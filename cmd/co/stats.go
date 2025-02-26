package main

import (
	"encoding/json"
	"fmt"

	codeowners "github.com/lukealbao/co"
	"github.com/spf13/cobra"
)

var statsCmd = &cobra.Command{
	Use:   "stats [filepath]...",
	Short: "Display code ownership statistics",
	Long: `Display code ownership statistics

Default format displays a table:

  Total files                    5 (100.00%)
  Owned files                    2 (40.00%)
  Unowned files                  3 (60.00%)
  Owner count                    2
  ----------------------------------------------
  (unowned)                                          3 (60.00%)
  owner-a                                            2 (40.00%)
  owner-b                                            1 (20.00%)

Unowned files are displayed as belonging to the dummy "(unowned)" group.
Ownership percentages may add up to more than 100%, as there can be more than one owner per file.

If filepaths are provided, only files matching the provided paths are considered.
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

		formatJson, err := cmd.Flags().GetBool("json")
		exitIf(err)

		files, err := codeowners.ListOwners(sessionRules, filesToCheck, ownerFilters, showUnowned)
		exitIf(err)

		stats := codeowners.CalculateOwnershipStats(files)
		if formatJson {
			bytes, err := json.MarshalIndent(stats, "", "  ")
			exitIf(err)

			fmt.Fprintf(cmd.OutOrStdout(), "%s\n", bytes)
		} else {
			displayOwnershipStats(stats)
		}
	},
}

func displayOwnershipStats(stats codeowners.OwnerStats) {
	fileCount := float64(stats.TotalFiles)
	ownedCount := float64(stats.OwnedFiles)
	unownedCount := float64(stats.UnownedFiles)
	filesPerOwner := stats.FilesPerOwner
	totalOwners := stats.OwnerCount

	fmt.Printf("%-30s %.0f (%.2f%%)\n", "Total files", fileCount, (fileCount/fileCount)*100)
	fmt.Printf("%-30s %.0f (%.2f%%)\n", "Owned files", ownedCount, (ownedCount/fileCount)*100)
	fmt.Printf("%-30s %.0f (%.2f%%)\n", "Unowned files", unownedCount, (unownedCount/fileCount)*100)
	fmt.Printf("%-30s %d\n", "Owner count", totalOwners)
	fmt.Println("----------------------------------------------")
	for _, kv := range filesPerOwner {
		fmt.Printf("%-50s %d (%.2f%%)\n", kv.Owner, kv.Count, kv.Percentage)
	}
}
