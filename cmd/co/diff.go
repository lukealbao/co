package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/lukealbao/co"
	"github.com/spf13/cobra"
)

var (
	diffRefFrom   string
	diffRulesFrom codeowners.Ruleset
	diffRefTo     string
	diffRulesTo   codeowners.Ruleset
)

var diffCmd = &cobra.Command{
	Use:   "diff [commit | commit..commit | commit commit]",
	Short: "Print a unified diff of file ownership",
	Long: `Diff calculates the ownership of all files at the given valid git refs, and prints a
unified diff of their text outputs.

Commit refs behave mostly like git diff, with subtle differences when comparing uncommitted changes.
The codeowners file will be read from disk, so any uncommitted changes will be considered. All other
files will be listed from the git index. That is, renamed, deleted, or added files will be considered
only after staging those changes.

Flags:
  -r, --renames Files that have been renamed will be printed as their latest name for both versions.`,
	Args: func(cmd *cobra.Command, args []string) error {
		refs := make([]string, 2, 2)

		for i, a := range args {
			refs[i] = a
		}

		first, second := refs[0], refs[1]

		// ref: sha1..sha2
		if strings.Contains(first, "..") {
			parts := strings.Split(first, "..")
			first, second = parts[0], parts[1]
			if strings.HasPrefix(second, ".") {
				fmt.Fprintln(cmd.ErrOrStderr(), fmt.Errorf("bad ref: %s", args))
				return cmd.Help()
			}
		}

		validRef := func(ref string) bool {
			_, err := exec.Command("git", "rev-parse", ref).Output()
			return err == nil
		}

		if first != "" && !validRef(first) {
			fmt.Fprintln(cmd.ErrOrStderr(), fmt.Errorf("bad ref %s", args))
			return cmd.Help()
		}

		if second != "" && !validRef(second) {
			fmt.Fprintln(cmd.ErrOrStderr(), fmt.Errorf("bad ref: %s", args))
			return cmd.Help()
		}

		// If we have only one ref, we may be looking to compare dirty file state with HEAD.
		if first == "" {
			first, second = "HEAD", ""
		}

		diffRefFrom, diffRefTo = first, second

		codeOwnersPath := cmd.Flag("file").Value.String()
		if codeOwnersPath == "" {
			var err error
			diffRulesFrom, err = codeowners.LoadFileFromStandardLocationAtRef(first)
			diffRulesTo, err = codeowners.LoadFileFromStandardLocationAtRef(second)
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "%s", err)
				os.Exit(1)
			}
		} else {
			var err error
			diffRulesFrom, err = codeowners.LoadFileAtRef(first, codeOwnersPath)
			diffRulesTo, err = codeowners.LoadFileAtRef(second, codeOwnersPath)
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "%s", err)
				os.Exit(1)
			}
		}

		return nil
	},
	Run: func(cmd *cobra.Command, files []string) {
		// TODO: user may want different refs. AND: may want to specify paths.
		trackedFilesFrom, err := codeowners.LsFiles(diffRefFrom)
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "%s", err)
			os.Exit(1)
		}

		refs, err := newFollower(".", diffRefFrom, diffRefTo)
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "%s", err)
			os.Exit(1)
		}

		trackedFilesTo, err := codeowners.LsFiles(diffRefTo)
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "%s", err)
			os.Exit(1)
		}

		owners1, err := listOwners(diffRulesFrom, trackedFilesFrom, nil, false)
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "%s", err)
			os.Exit(1)
		}

		owners2, err := listOwners(diffRulesTo, trackedFilesTo, nil, false)
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "%s", err)
			os.Exit(1)
		}

		Tmp, err := ioutil.TempDir("", "co-diff-")
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "%s", err)
			os.Exit(1)
		}

		rev1, _ := exec.Command("git", "rev-parse", diffRefFrom).Output()
		if string(rev1) == "\n" {
			rev1 = []byte("git-index")
		} else {
			rev1 = []byte(strings.Trim(fmt.Sprintf("%s", rev1), " \n"))
		}

		tmp1, err := os.Create(filepath.Join(Tmp, fmt.Sprintf("owners@%s.list", rev1)))
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "%s", err)
			os.Exit(1)
		}

		// Todo, this is a hack to sort.
		for _, result := range owners1 {
			if follow, err := cmd.Flags().GetBool("renames"); err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "%s", err)
				os.Exit(1)
			} else if follow {
				result.path = refs.latestName(result.path)
			}
		}

		var oowners1 rs = owners1
		sort.Sort(oowners1)

		for _, result := range oowners1 {
			fmt.Fprintf(tmp1, "%-70s %s\n", result.path, result.owners)
		}

		rev2, _ := exec.Command("git", "rev-parse", diffRefTo).Output()
		if string(rev2) == "\n" {
			rev2 = []byte("git-index")
		} else {
			rev2 = []byte(strings.Trim(fmt.Sprintf("%s", rev2), " \n"))
		}
		tmp2, err := os.Create(filepath.Join(Tmp, fmt.Sprintf("owners@%s.list", rev2)))
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "%s", err)
			os.Exit(1)
		}
		for _, result := range owners2 {
			fmt.Fprintf(tmp2, "%-70s %s\n", result.path, result.owners)
		}

		diff := exec.Command("git", "diff", "-u", "-w", "--no-index", tmp1.Name(), tmp2.Name())
		// exit code 1 on diff is not a real error in this use case.
		unifiedDiff, _ := diff.CombinedOutput()

		lines := strings.Split(string(unifiedDiff), "\n")
		var wasDiff bool

		var output io.Writer

		var pager string
		if pager = os.Getenv("PAGER"); pager == "" {
			pager = "less"
		}

		p := exec.Command(pager, "-R", "-F", "-X")
		p.Stdout = os.Stdout
		p.Stderr = os.Stderr

		stdin, err := p.StdinPipe()
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			os.Exit(1)
		}
		output = stdin

		if err := p.Start(); err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			os.Exit(1)
		}

		defer func() {
			if err := p.Wait(); err != nil {
				fmt.Fprintln(cmd.ErrOrStderr(), err)
				os.Exit(1)
			}

			if wasDiff {
				os.Exit(1)
			}
		}()

		for i, line := range lines {
			switch {
			case strings.HasPrefix(line, "+"):
				fmt.Fprintln(output, color.GreenString(line))
				wasDiff = true
				break
			case strings.HasPrefix(line, "-"):
				fmt.Fprintln(output, color.RedString(line))
				wasDiff = true
				break
			default:
				// Don't print empty final line.
				if i != len(lines)-1 || line != "" {
					fmt.Fprintln(output, line)
				}
			}
		}

		stdin.Close()
	},
}

// follower maps older names to newer names.
type follower map[string]string

func newFollower(path, base, current string) (follower, error) {
	return findRenames(path, base, current)
}

func (f follower) latestName(path string) string {
	var ok bool
	cycleCheck := make(map[string]struct{})

	for {
		if _, ok := cycleCheck[path]; ok {
			return path
		} else {
			cycleCheck[path] = struct{}{}
		}

		curr := path
		path, ok = f[path]

		if !ok {
			return curr
		}
	}
}

func findRenames(path, base, current string) (follower, error) {
	rename := regexp.MustCompile(`^R(\d+)\t(.*)\t(.*)$`)
	var (
		out follower = make(map[string]string)
		err error
	)

	gitRange := fmt.Sprintf("%s..%s", base, current)
	log, err := exec.Command("git", "log", "--name-status", "--pretty=format:''", "--diff-filter=R", gitRange, path).Output()
	if err != nil {
		return out, err
	}

	scanner := bufio.NewScanner(bytes.NewReader(log))

	var i = 0
	for scanner.Scan() {
		i += 1
		if match := rename.FindStringSubmatch(scanner.Text()); match != nil {
			_ /* pct */, from, to := match[1], match[2], match[3]
			out[from] = to
		}
	}

	return out, err
}

type r struct {
	path   string
	owners []string
}

type rs []*r

func (x rs) Len() int           { return len(x) }
func (x rs) Less(a, b int) bool { return x[a].path < x[b].path }
func (x rs) Swap(a, b int) {
	x[a], x[b] = x[b], x[a]
}

var _ sort.Interface = (rs)(nil)

// listOwners returns a list of structured output. Callers must format for printing.
func listOwners(rules codeowners.Ruleset, files []string, ownerFilter []string, showUnowned bool) ([]*r, error) {
	var out []*r = make([]*r, 0)

	for _, file := range files {
		rule, err := rules.Match(file)
		if err != nil {
			return nil, err
		}

		if rule == nil || rule.Owners == nil || len(rule.Owners) == 0 {
			if len(ownerFilters) == 0 || showUnowned {
				out = append(out, &r{path: file, owners: []string{"(unowned)"}})
			}

			continue
		}

		owners := make([]string, 0, len(rule.Owners))
		for _, owner := range rule.Owners {
			filterMatch := len(ownerFilters) == 0 && !showUnowned
			for _, filter := range ownerFilters {
				if filter == owner { // TODO: This is "Value" in hmarr. Are we losing info?
					filterMatch = true
				}
			}
			if filterMatch {
				owners = append(owners, owner)
			}
		}

		if len(owners) > 0 {
			out = append(out, &r{path: file, owners: owners})
		}
	}

	return out, nil
}
