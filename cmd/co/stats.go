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
	Long: `List code owners for file(s)

Note that unowned files are displayed as belonging to the dummy "(unowned)" group.
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
			bytes, err := json.MarshalIndent(stats.FilesPerOwner, "", "  ")
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

	fmt.Printf("Total files: %.0f (%.2f%%)\n", fileCount, (fileCount/fileCount)*100)
	fmt.Printf("Total files with owners: %.0f (%.2f%%)\n", ownedCount, (ownedCount/fileCount)*100)
	fmt.Printf("Total unowned files: %.0f (%.2f%%)\n", unownedCount, (unownedCount/fileCount)*100)
	fmt.Printf("Total owners: %d\n", totalOwners)
	fmt.Println()
	for _, kv := range filesPerOwner {
		fmt.Printf("%s: %d (%.2f%%)\n", kv.Owner, kv.Count, kv.Percentage)
	}
}
