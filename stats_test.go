package codeowners

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculateOwnershipStats(t *testing.T) {
	tests := []struct {
		name     string
		files    Owners
		expected OwnerStats
	}{
		{
			name:  "empty files list",
			files: Owners{},
			expected: OwnerStats{
				TotalFiles:    0,
				OwnedFiled:    0,
				UnownedFiles:  0,
				TotalOwners:   0,
				FilesPerOwner: []FilesPerOwner{},
			},
		},
		{
			name: "single owned file",
			files: Owners{
				{
					Path:   "file1.txt",
					Owners: []string{"@teamA"},
				},
			},
			expected: OwnerStats{
				TotalFiles:   1,
				OwnedFiled:   1,
				UnownedFiles: 0,
				TotalOwners:  1,
				FilesPerOwner: []FilesPerOwner{
					{
						Owner:      "@teamA",
						Count:      1,
						Percentage: 100.0,
					},
				},
			},
		},
		{
			name: "single unowned file",
			files: Owners{
				{
					Path:   "file1.txt",
					Owners: []string{"(unowned)"},
				},
			},
			expected: OwnerStats{
				TotalFiles:   1,
				OwnedFiled:   0,
				UnownedFiles: 1,
				TotalOwners:  0,
				FilesPerOwner: []FilesPerOwner{
					{
						Owner:      "(unowned)",
						Count:      1,
						Percentage: 100.0,
					},
				},
			},
		},
		{
			name: "multiple files with different owners",
			files: Owners{
				{
					Path:   "file1.txt",
					Owners: []string{"@teamA", "@teamB"},
				},
				{
					Path:   "file2.txt",
					Owners: []string{"@teamA"},
				},
				{
					Path:   "file3.txt",
					Owners: []string{"(unowned)"},
				},
			},
			expected: OwnerStats{
				TotalFiles:   3,
				OwnedFiled:   2,
				UnownedFiles: 1,
				TotalOwners:  2,
				FilesPerOwner: []FilesPerOwner{
					{
						Owner:      "@teamA",
						Count:      2,
						Percentage: 66.67,
					},
					{
						Owner:      "@teamB",
						Count:      1,
						Percentage: 33.33,
					},
					{
						Owner:      "(unowned)",
						Count:      1,
						Percentage: 33.33,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateOwnershipStats(tt.files)

			assert.Equal(t, tt.expected.TotalFiles, result.TotalFiles)
			assert.Equal(t, tt.expected.OwnedFiled, result.OwnedFiled)
			assert.Equal(t, tt.expected.UnownedFiles, result.UnownedFiles)
			assert.Equal(t, tt.expected.TotalOwners, result.TotalOwners)

			// For FilesPerOwner, we need to check each field separately due to floating point comparison
			assert.Equal(t, len(tt.expected.FilesPerOwner), len(result.FilesPerOwner))
			for i, expectedOwner := range tt.expected.FilesPerOwner {
				assert.Equal(t, expectedOwner.Owner, result.FilesPerOwner[i].Owner)
				assert.Equal(t, expectedOwner.Count, result.FilesPerOwner[i].Count)
				// Use InDelta for floating point comparison with a small delta
				assert.InDelta(t, expectedOwner.Percentage, result.FilesPerOwner[i].Percentage, 0.01)
			}
		})
	}
}
