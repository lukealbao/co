package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var whyCmd = &cobra.Command{
	Use:   "why file",
	Short: "Show rule used for a file",
	Run: func(cmd *cobra.Command, files []string) {
		if len(files) != 1 {
			cmd.Help()
			os.Exit(1)
		}

		for _, file := range files {
			rule, err := sessionRules.Match(file)
			exitIf(err)

			if rule == nil {
				fmt.Fprintf(cmd.OutOrStdout(), "  %4d %-70s %s\n", -1, "[no match]", "[unowned]")
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "  %4d %-70s %s\n", rule.SourceLine, rule.RawPattern(), rule.Owners)
			}
		}
	},
}
