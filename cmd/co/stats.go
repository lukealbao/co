package main

import (
	"encoding/json"
	"fmt"

	codeowners "github.com/lukealbao/co"
	"github.com/spf13/cobra"
)

var statsCmd = &cobra.Command{
	Use:   "stats [filepath]...",
	Short: "List code ownership statistics for file(s)",
	Long: `List code ownership statistics for file(s)

Total files, owned files, unowned files, and owner count are displayed.

Per owner file count and percentage are also displayed, sorted by file count.
Unowned files are displayed as belonging to the dummy "(unowned)" group.

Ownership percentages may add up to more than 100%, as there can be more than one owner per file.
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
	totalOwners := stats.TotalOwners

	fmt.Printf("%-30s %.0f (%.2f%%)\n", "Total files", fileCount, (fileCount/fileCount)*100)
	fmt.Printf("%-30s %.0f (%.2f%%)\n", "Total files with owners", ownedCount, (ownedCount/fileCount)*100)
	fmt.Printf("%-30s %.0f (%.2f%%)\n", "Total unowned files", unownedCount, (unownedCount/fileCount)*100)
	fmt.Printf("%-30s %d\n", "Total owners", totalOwners)
	fmt.Println("----------------------------------------------")
	for _, kv := range filesPerOwner {
		fmt.Printf("%-50s %d (%.2f%%)\n", kv.Owner, kv.Count, kv.Percentage)
	}
}
