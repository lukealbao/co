package codeowners

import (
	"sort"
)

type FilesPerOwner struct {
	Owner      string  `json:"owner"`
	Count      int     `json:"fileCount"`
	Percentage float64 `json:"percentage"`
}

type OwnerStats struct {
	TotalFiles    int
	OwnedFiled    int
	UnownedFiles  int
	TotalOwners   int
	FilesPerOwner []FilesPerOwner
}

func CalculateOwnershipStats(files Owners) OwnerStats {
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
	totalOwners := len(stats)
	if _, hasUnowned := stats["(unowned)"]; hasUnowned {
		totalOwners--
	}

	return OwnerStats{
		TotalFiles:    fileCount,
		OwnedFiled:    ownedCount,
		UnownedFiles:  unownedCount,
		FilesPerOwner: filesPerOwner,
		TotalOwners:   totalOwners,
	}
}
