package codeowners

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConsolidateSubTree(t *testing.T) {
	type test struct {
		label         string
		rules         map[string][]string
		expectedRules map[string][]string
		fsStat        FsStat
	}

	tests := []test{
		{
			label: "singleton as leaf",
			rules: map[string][]string{
				"root": {"a"},
			},
			expectedRules: map[string][]string{
				"root": {"a"},
			},
			fsStat: statMock(),
		},
		{
			label: "singleton as dir",
			rules: map[string][]string{
				"root/": {"a"},
			},
			expectedRules: map[string][]string{
				"root/": {"a"},
			},
			fsStat: statMock(),
		},
		{
			label: "singleton as glob",
			rules: map[string][]string{
				"root/*": {"a"},
			},
			expectedRules: map[string][]string{
				"root/*": {"a"},
			},
			fsStat: statMock(),
		},
		{
			label: "siblings w/ root as leaf",
			rules: map[string][]string{
				"root":   {"a"},
				"root/a": {"a"},
				"root/b": {"a"},
				"root/c": {"a"},
			},
			expectedRules: map[string][]string{
				"root": {"a"},
			},
			fsStat: statMock(),
		},
		{
			label: "siblings w/ root as dir",
			rules: map[string][]string{
				"root/":  {"a"},
				"root/a": {"a"},
				"root/b": {"a"},
				"root/c": {"a"},
			},
			expectedRules: map[string][]string{
				"root/": {"a"},
			},
			fsStat: statMock(),
		},
		{
			label: "siblings w/ root as glob",
			rules: map[string][]string{
				"root/*": {"a"},
				"root/a": {"a"},
				"root/b": {"a"},
				"root/c": {"a"},
			},
			expectedRules: map[string][]string{
				"root/*": {"a"},
			},
			fsStat: statMock(),
		},
		{
			label: "deep tree w/ root as leaf",
			rules: map[string][]string{
				"root":     {"a"},
				"root/a":   {"a"},
				"root/b/c": {"a"},
				"root/d":   {"a"},
			},
			expectedRules: map[string][]string{
				"root": {"a"},
			},
			fsStat: statMock(),
		},
		{
			label: "deep tree w/ root as dir",
			rules: map[string][]string{
				"root/":    {"a"},
				"root/a":   {"a"},
				"root/b/c": {"a"},
				"root/d":   {"a"},
			},
			expectedRules: map[string][]string{
				"root/": {"a"},
			},
			fsStat: statMock(),
		},
		{
			label: "deep tree w/ root as glob",
			rules: map[string][]string{
				"root/*":   {"a"},
				"root/a":   {"a"},
				"root/b/c": {"a"},
				"root/d":   {"a"},
			},
			expectedRules: map[string][]string{
				"root/*":   {"a"},
				"root/b/c": {"a"},
			},
			fsStat: statMock(),
		},
		{
			// This is known to be flaky due to map comparison in stretchr.
			label: "no match results in no changes",
			rules: map[string][]string{
				"root":     {"a"},
				"root/a":   {"a"},
				"root/b/c": {"a", "b"},
				"root/d":   {"a"},
			},
			expectedRules: map[string][]string{
				"root":     {"a"},
				"root/a":   {"a"},
				"root/b/c": {"a", "b"},
				"root/d":   {"a"},
			},
			fsStat: statMock(),
		},
		{
			label: "don't over-consolidate",
			rules: map[string][]string{
				"root/a": {"a"},
				"root/b": {"a"},
			},
			expectedRules: map[string][]string{
				"root/a": {"a"},
				"root/b": {"a"},
			},
			fsStat: statMock(),
		},
		{
			label: "Handle unicode",
			rules: map[string][]string{
				"root/😃":     {"a"},
				"root/😃/*":   {"a"},
				"root/😃/a/b": {"a"},
				"root/😃😃/*":  {"a"},
			},
			expectedRules: map[string][]string{
				"root/😃":    {"a"},
				"root/😃😃/*": {"a"},
			},
			fsStat: statMock(),
		},
		{
			label: "implicit root is honored",
			rules: map[string][]string{
				"*":          {"a"},
				"root/a/*":   {"a"},
				"root/a/a/b": {"a"},
			},
			expectedRules: map[string][]string{
				"*": {"a"},
			},
			fsStat: statMock(),
		},
		{
			label: "directory dominates glob",
			rules: map[string][]string{
				"root/":     {"a"},
				"root/**/*": {"a"},
			},
			expectedRules: map[string][]string{
				"root/": {"a"},
			},
			fsStat: statMock(),
		},
		{
			label: "partial prefix match",
			rules: map[string][]string{
				"root":           {"a"},
				"root-toot/**/*": {"a"},
			},
			expectedRules: map[string][]string{
				"root":           {"a"},
				"root-toot/**/*": {"a"},
			},
			fsStat: statMock(),
		},
		{
			label: "shallow root glob doesn't dominate subdirectories",
			rules: map[string][]string{
				"root/*":     {"a"},
				"root/dir/x": {"a"},
				"root/file":  {"a"}, // shallow glob matches non dirs
			},
			expectedRules: map[string][]string{
				"root/*":     {"a"},
				"root/dir/x": {"a"},
			},
			fsStat: statMock(map[string]bool{"root/dir": true, "root/file": false}),
		},
	}

	for _, test := range tests {
		t.Run(test.label, func(t *testing.T) {
			// We will extract the consolidated rules from tree after consolidation.
			var (
				tree              *FileTree
				consolidatedRules map[string][]string
			)

			// Set up tree.
			{
				var rules []Rule

				for path, owners := range test.rules {
					pat, err := newPattern(path)
					assert.NoError(t, err)

					r := newRule()
					r.pattern = pat

					for _, owner := range owners {
						r.Owners = append(r.Owners, owner)
					}

					rules = append(rules, *r)
				}

				tree = NewFileTree(rules)
			}

			// System under test.
			ConsolidateTree(tree, test.fsStat)

			// Extract actual changes.
			{
				consolidatedRules = make(map[string][]string)
				for pat, rule := range tree.rules {
					for _, o := range rule.Owners {
						consolidatedRules[pat] = append(consolidatedRules[pat], o)
					}
				}
			}

			assert.Equal(t, test.expectedRules, consolidatedRules)
		})
	}
}
