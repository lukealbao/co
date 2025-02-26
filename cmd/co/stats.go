package main

import (
	"encoding/json"
	"fmt"
	"sort"

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

		files, err := listOwners(sessionRules, filesToCheck, ownerFilters, showUnowned)
		exitIf(err)

		stats := calculateOwnershipStats(files)
		if formatJson {
			bytes, err := json.MarshalIndent(stats.filesPerOwner, "", "  ")
			exitIf(err)

			fmt.Fprintf(cmd.OutOrStdout(), "%s\n", bytes)
		} else {
			displayOwnershipStats(stats)
		}
	},
}

type FilesPerOwner struct {
	Owner      string  `json:"owner"`
	Count      int     `json:"fileCount"`
	Percentage float64 `json:"percentage"`
}

type OwnerStats struct {
	totalFiles    int
	ownedFiled    int
	unownedFiles  int
	totalOwners   int
	filesPerOwner []FilesPerOwner
}

func calculateOwnershipStats(files []*r) OwnerStats {
	fileCount := len(files)

	stats := make(map[string]int)
	for _, file := range files {
		for _, owner := range file.Owners {
			stats[owner]++
		}
	}

	var filesPerOwner []FilesPerOwner
	for owner, count := range stats {
		percentage := (float64(count) / float64(fileCount)) * 100
		filesPerOwner = append(filesPerOwner, FilesPerOwner{owner, count, percentage})
	}

	sort.Slice(filesPerOwner, func(i, j int) bool {
		return filesPerOwner[i].Count > filesPerOwner[j].Count
	})

	unownedCount := stats["(unowned)"]
	ownedCount := fileCount - unownedCount
	totalOwners := len(stats) - 1

	return OwnerStats{
		totalFiles:    fileCount,
		ownedFiled:    ownedCount,
		unownedFiles:  unownedCount,
		filesPerOwner: filesPerOwner,
		totalOwners:   totalOwners,
	}
}

func displayOwnershipStats(stats OwnerStats) {
	fileCount := float64(stats.totalFiles)
	ownedCount := float64(stats.ownedFiled)
	unownedCount := float64(stats.unownedFiles)
	filesPerOwner := stats.filesPerOwner
	totalOwners := stats.totalOwners

	fmt.Printf("Total files: %.0f (%.2f%%)\n", fileCount, (fileCount/fileCount)*100)
	fmt.Printf("Total files with owners: %.0f (%.2f%%)\n", ownedCount, (ownedCount/fileCount)*100)
	fmt.Printf("Total unowned files: %.0f (%.2f%%)\n", unownedCount, (unownedCount/fileCount)*100)
	fmt.Printf("Total owners: %d\n", totalOwners)
	fmt.Println()
	for _, kv := range filesPerOwner {
		fmt.Printf("%s: %d (%.2f%%)\n", kv.Owner, kv.Count, kv.Percentage)
	}
}
