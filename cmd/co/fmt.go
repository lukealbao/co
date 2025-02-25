package main

import (
	"fmt"
	"os"

	"github.com/lukealbao/co"
	"github.com/spf13/cobra"
)

var fmtCmd = &cobra.Command{
	Use:   "fmt",
	Short: "Normalize CODEOWNERS format",
	Long: `Format CODEOWNERS file in place.

The file will be lexicographically sorted. Use --trim to remove redundant rules.`,
	Run: func(cmd *cobra.Command, files []string) {
		tree := codeowners.NewFileTree(sessionRules)

		trim, err := cmd.Flags().GetBool("trim")
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			os.Exit(1)
		}

		if trim {
			codeowners.ConsolidateTree(tree)
		}

		file, err := os.OpenFile(codeownersPath, os.O_TRUNC|os.O_WRONLY, os.ModePerm)
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			os.Exit(1)
		}

		if err := file.Truncate(0); err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			os.Exit(1)
		}

		if _, err := file.WriteString(tree.String()); err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			os.Exit(1)
		}
	},
}
