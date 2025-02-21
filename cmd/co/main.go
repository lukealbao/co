package main

import (
	"os"

	"github.com/lukealbao/co"
	"github.com/spf13/cobra"
)

var root = &cobra.Command{
	Use: "co",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		path := cmd.Flag("file").Value.String()
		var err error

		whoami := cmd.Name()
		if whoami == "help" || whoami == "version" {
			return nil
		}

		if path == "" {
			sessionRules, err = codeowners.LoadFileFromStandardLocationAtRef("")
		} else {
			sessionRules, err = codeowners.LoadFileAtRef("", path)
		}

		return err
	},
}

// Globals
var (
	codeownersPath string
	ownerFilters   []string
	showUnowned    bool
	sessionRules   codeowners.Ruleset
	// Ldflags passed in by goreleaser's defaults:
	version string
	commit  string
	date    string
)

func init() {
	root.PersistentFlags().StringVarP(&codeownersPath, "file", "f", "", "CODEOWNERS file path")
	whoCmd.Flags().StringSliceVarP(&ownerFilters, "owner", "o", nil, "filter results by owner")
	whoCmd.Flags().BoolVarP(&showUnowned, "unowned", "u", false, "only show unowned files (can be combined with -o)")
	whoCmd.Flags().BoolP("json", "j", false, "format output as json. output is Array<{path: string; owners: Array<string>}>.")

	root.AddCommand(whoCmd)
	root.AddCommand(whyCmd)

	diffCmd.Flags().BoolP("renames", "r", false, "follow file renames")
	root.AddCommand(diffCmd)

	fmtCmd.Flags().StringP("file", "f", "", "path to CODEOWNERS file (required)")
	fmtCmd.Flags().BoolP("trim", "t", false, "rollup rules into matching parent globs, if any exist")
	fmtCmd.MarkFlagRequired("file")
	root.AddCommand(fmtCmd)

	lintCmd.Flags().StringP("file", "f", "", "path to CODEOWNERS file (required)")
	lintCmd.Flags().Bool("fix", false, "edit CODEOWNERS file to remove unused rules")
	lintCmd.MarkFlagRequired("file")
	root.AddCommand(lintCmd)

	root.AddCommand(versionCmd)

	// TODO: add completion
	root.CompletionOptions.DisableDefaultCmd = true
}

func main() {
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
