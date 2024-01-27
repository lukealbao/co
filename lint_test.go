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
		},
		{
			label: "singleton as dir",
			rules: map[string][]string{
				"root/": {"a"},
			},
			expectedRules: map[string][]string{
				"root/": {"a"},
			},
		},
		{
			label: "singleton as glob",
			rules: map[string][]string{
				"root/*": {"a"},
			},
			expectedRules: map[string][]string{
				"root/*": {"a"},
			},
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
			ConsolidateTree(tree)

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
