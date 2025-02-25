package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/lukealbao/co"
	"github.com/spf13/cobra"
)

func exitIf(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
}

var whoCmd = &cobra.Command{
	Use:   "who [filepath]...",
	Short: "List code owners for file(s)",
	Long: `List code owners for file(s)

Default format displays a table mapping files to the list of owners:

    path/to/file/a                            [@backend @infrastructure]
    path/to/file/b                            [(unowned)]

Note that unowned files are displayed as belonging to the dummy "(unowned)" group.

JSON-formatted output displays an array of objects. Unowned files have a null owners list:

    [
      {
        "path": "path/to/file/a",
        "owners": ["@backend", "@infrastructure"]
      },
      {
        "path": "path/to/file/b",
        "owners": null
      }
    ]
`,
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		var filesToCheck []string

		if len(args) > 0 {
			filesToCheck = expandAllFiles(args[:])
		} else {
			filesToCheck, err = codeowners.LsFiles("HEAD")
			exitIf(err)
		}

		files, err := listOwners(sessionRules, filesToCheck, ownerFilters, showUnowned)
		exitIf(err)

		displayStats, err := cmd.Flags().GetBool("stats")
		exitIf(err)
		if displayStats {
			displayOwnershipStats(files)
			return
		}

		formatJson, err := cmd.Flags().GetBool("json")
		exitIf(err)

		if formatJson {
			bytes, err := json.MarshalIndent(files, "", "  ")
			exitIf(err)

			fmt.Fprintf(cmd.OutOrStdout(), "%s\n", bytes)
			return
		}

		for _, result := range files {
			fmt.Fprintf(cmd.OutOrStdout(), "%-70s %s\n", result.Path, result.Owners)
		}
	},
}

func displayOwnershipStats(files []*r) {
	stats := make(map[string]int)
	for _, file := range files {
		for _, owner := range file.Owners {
			stats[owner]++
		}
	}

	fileCount := float64(len(files))
	unownedCount := float64(stats["(unowned)"])
	ownedCount := float64(fileCount - unownedCount)

	fmt.Printf("Total files: %.0f (%.2f%%)\n", fileCount, (fileCount/fileCount)*100)
	fmt.Printf("Total files with owners: %.0f (%.2f%%)\n", ownedCount, (ownedCount/fileCount)*100)
	fmt.Printf("Total unowned files: %.0f (%.2f%%)\n", unownedCount, (unownedCount/fileCount)*100)
	fmt.Printf("Total owners: %d\n", len(stats)-1)

	fmt.Println()

	// sort owners by count
	type FilesPerOwner struct {
		Owner string
		Count int
	}

	var ss []FilesPerOwner
	for Owner, Count := range stats {
		ss = append(ss, FilesPerOwner{Owner, Count})
	}

	sort.Slice(ss, func(i, j int) bool {
		return ss[i].Count > ss[j].Count
	})

	for _, kv := range ss {
		percentage := (float64(kv.Count) / fileCount) * 100
		fmt.Printf("%s: %d (%.2f%%)\n", kv.Owner, kv.Count, percentage)
	}
}

func expandAllFiles(paths []string) []string {
	out := make([]string, 0)

	isDir := func(path string) bool {
		info, err := os.Stat(path)
		if os.IsNotExist(err) {
			return false
		}
		return info.IsDir()
	}

	for _, p := range paths {
		if !isDir(p) {
			out = append(out, p)
			continue
		}

		filepath.WalkDir(p, func(path string, d os.DirEntry, err error) error {
			if path == ".git" {
				return filepath.SkipDir
			}

			if !isDir(path) {
				out = append(out, path)
			}

			return nil
		})
	}

	return out
}
