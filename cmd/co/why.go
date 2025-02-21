package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

type filematch struct {
	Path   string   `json:"path"`
	Line   int      `json:"line"`
	Rule   *string  `json:"rule"`
	Owners []string `json:"owners"`
}

var whyCmd = &cobra.Command{
	Use:   "why file",
	Short: "Identify which rule effects ownership for a single file.",
	Long: `Identify which rule effects ownership for a single file.

Default format displays the line number of the effective rule along with the spec:

    42 backend/**/*.test.ts                            [@backend @qa]

Unowned files will have a negative line number:

    -1 (no match)                                      (unowned)

JSON-formatted output displays an object that includes the file path:

    {
      "path": "backend/db/users.test.ts",
      "line": 42,
      "owners": ["@backend", "@qa"]
    }

If the file is unowned, the owners list will be null:
    {
      "path": "path/to/file/b",
      "line": -1,
      "owners": null
    }
`,
	Run: func(cmd *cobra.Command, files []string) {
		if len(files) != 1 {
			cmd.Help()
			os.Exit(1)
		}

		formatJson, err := cmd.Flags().GetBool("json")
		exitIf(err)

		if formatJson {
			match := filematch{Line: -1}

			// Yes, files is actually restricted to len(1) above.
			for _, file := range files {
				rule, err := sessionRules.Match(file)
				exitIf(err)

				if rule != nil {
					match.Path = file
					match.Line = rule.SourceLine
					pat := rule.RawPattern()
					match.Rule = &pat
					match.Owners = rule.Owners
				}
			}

			bytes, err := json.MarshalIndent(match, "", "  ")
			exitIf(err)

			fmt.Fprintf(cmd.OutOrStdout(), "%s\n", bytes)
			return
		}

		for _, file := range files {
			rule, err := sessionRules.Match(file)
			exitIf(err)

			if rule == nil {
				fmt.Fprintf(cmd.OutOrStdout(), "  %4d %-70s %s\n", -1, "(no match)", "(unowned)")
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "  %4d %-70s %s\n", rule.SourceLine, rule.RawPattern(), rule.Owners)
			}
		}
	},
}
