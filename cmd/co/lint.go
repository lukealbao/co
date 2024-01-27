package main

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/lukealbao/co"
	"github.com/spf13/cobra"
)

var lintCmd = &cobra.Command{
	Use:   "lint",
	Short: "Validate codeowners file",
	Long:  `Check for unused rules.`,
	Run: func(cmd *cobra.Command, _ []string) {
		files, err := codeowners.LsFiles("")
		exitIf(err)

		matchAny := func(rule codeowners.Rule, paths []string) bool {
			for _, path := range paths {
				if ok, _ := rule.Match(path); ok {
					return true
				}
			}
			return false
		}

		var errors codeowners.Ruleset

		for _, rule := range sessionRules {
			if !matchAny(rule, files) {
				errors = append(errors, rule)
			}
		}

		if fix, err := cmd.Flags().GetBool("fix"); err != nil {
			exitIf(err)
		} else if !fix {
			if len(errors) > 0 {
				fmt.Println(color.HiRedString("Error"), "Unused Rules:")
				for _, rule := range errors {
					fmt.Fprintf(os.Stdout, "%4d %-70s %s\n", rule.SourceLine, rule.RawPattern(), rule.Owners)
				}
				os.Exit(1)
			} else {
				return
			}
		}

		// Will attempt to fix.
		if len(errors) == 0 {
			return
		}

		for _, unusedRule := range errors {
			for i, rule := range sessionRules {
				if rule.SourceLine == unusedRule.SourceLine {
					sessionRules = append(sessionRules[:i], sessionRules[i+1:]...)
				}
			}
		}

		filepath, err := cmd.Flags().GetString("file")
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			os.Exit(1)
		}

		file, err := os.OpenFile(filepath, os.O_TRUNC|os.O_WRONLY, os.ModePerm)
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			os.Exit(1)
		}

		if err := file.Truncate(0); err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			os.Exit(1)
		}

		for _, rule := range sessionRules {
			_, err := fmt.Fprintln(file, rule.String())
			exitIf(err)
		}
	},
}
