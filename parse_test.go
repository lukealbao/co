package codeowners

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatting(t *testing.T) {
	examples := []struct {
		label      string
		ownersFile string
		expected   string
	}{
		{
			label: "end-to-end",
			ownersFile: `# Alpha should be consolidated.
alpha/ @test/user

alpha/a/b @test/user

# Comment on delta
delta @test/a @test/b
delta/a @test/a
delta/b/* @test/b

## Beta should come before delta.

beta @test/beta ## With a trailing comment.
`,
			expected: `# Alpha should be consolidated.
alpha/ @test/user

## Beta should come before delta.

beta @test/beta ## With a trailing comment.

# Comment on delta
delta @test/a @test/b
delta/a @test/a
delta/b/* @test/b
`,
		},
	}

	for _, tc := range examples {
		t.Run(tc.label, func(t *testing.T) {
			file := strings.NewReader(tc.ownersFile)
			rules, err := ParseFile(file)
			assert.NoError(t, err)

			tree := NewFileTree(rules)
			ConsolidateTree(tree)

			formatted := tree.String()
			assert.Equal(t, tc.expected, formatted)
		})
	}
}

// TestParseRule is ported from the original github.com/hmarr/codeowners test suite.
func TestParseRule(t *testing.T) {
	examples := []struct {
		name     string
		rule     string
		expected []Rule
		err      string
	}{
		// Success cases
		{
			name: "leading comments",
			rule: "# comment a\n\n# comment b\n\nfile.txt @user",
			expected: []Rule{
				{
					SourceLine:     5,
					pattern:        mustBuildPattern(t, "file.txt"),
					Owners:         []string{"@user"},
					leadingComment: "# comment a\n\n# comment b\n\n",
				},
			},
		},
		{
			name: "username owners",
			rule: "file.txt @user",
			expected: []Rule{
				{
					SourceLine: 1,
					pattern:    mustBuildPattern(t, "file.txt"),
					Owners:     []string{"@user"},
				},
			},
		},
		{
			name: "team owners",
			rule: "file.txt @org/team",
			expected: []Rule{
				{
					SourceLine: 1,
					pattern:    mustBuildPattern(t, "file.txt"),
					Owners:     []string{"@org/team"},
				},
			},
		},
		{
			name: "email owners",
			rule: "file.txt foo@example.com",
			expected: []Rule{
				{
					SourceLine: 1,
					pattern:    mustBuildPattern(t, "file.txt"),
					Owners:     []string{"foo@example.com"},
				},
			},
		},
		{
			name: "multiple owners",
			rule: "file.txt @user @org/team foo@example.com",
			expected: []Rule{{
				SourceLine: 1,
				pattern:    mustBuildPattern(t, "file.txt"),
				Owners: []string{
					"@user",
					"@org/team",
					"foo@example.com",
				},
			},
			},
		},
		{
			name: "complex patterns",
			rule: "d?r/* @user",
			expected: []Rule{{
				SourceLine: 1,
				pattern:    mustBuildPattern(t, "d?r/*"),
				Owners:     []string{"@user"},
			},
			},
		},
		{
			name: "pattern with space",
			rule: "foo\\ bar @user",
			expected: []Rule{{
				SourceLine: 1,
				pattern:    mustBuildPattern(t, "foo\\ bar"),
				Owners:     []string{"@user"},
			}},
		},
		{
			name: "comments",
			rule: "file.txt @user # some comment",
			expected: []Rule{{
				SourceLine:      1,
				pattern:         mustBuildPattern(t, "file.txt"),
				Owners:          []string{"@user"},
				trailingComment: "# some comment",
			}},
		},
		{
			name: "pattern with no owners",
			rule: "pattern",
			expected: []Rule{{
				SourceLine:      1,
				pattern:         mustBuildPattern(t, "pattern"),
				Owners:          []string{},
				trailingComment: "",
			}},
		},
		{
			name: "pattern with no owners and comment",
			rule: "pattern # but no more",
			expected: []Rule{{
				SourceLine:      1,
				pattern:         mustBuildPattern(t, "pattern"),
				Owners:          []string{},
				trailingComment: "# but no more",
			}},
		},
		{
			name: "pattern with no owners with whitespace",
			rule: "pattern ",
			expected: []Rule{{
				SourceLine:      1,
				pattern:         mustBuildPattern(t, "pattern"),
				Owners:          []string{},
				trailingComment: "",
			}},
		},
		{
			name: "pattern with leading and trailing whitespace",
			rule: " pattern @user ",
			expected: []Rule{{
				SourceLine:      1,
				pattern:         mustBuildPattern(t, "pattern"),
				Owners:          []string{"@user"},
				trailingComment: "",
			}},
		},
		{
			name: "pattern with leading and trailing whitespace and no owner",
			rule: " pattern ",
			expected: []Rule{{
				SourceLine:      1,
				pattern:         mustBuildPattern(t, "pattern"),
				Owners:          []string{},
				trailingComment: "",
			}},
		},

		// Error cases

		{
			name: "malformed patterns",
			rule: "file.{txt @user",
			err:  "line 1: unexpected character '{' at position 6",
		},
		{
			name: "patterns with brackets",
			rule: "file.[cC] @user",
			err:  "line 1: unexpected character '[' at position 6",
		},
		{
			name: "malformed owners",
			rule: "file.txt missing-at-sign",
			err:  "line 1: invalid owner format 'missing-at-sign' at position 10",
		},
	}

	for _, e := range examples {
		t.Run("parses "+e.name, func(t *testing.T) {
			file := strings.NewReader(e.rule)
			actual, err := ParseFile(file)
			if e.err != "" {
				assert.EqualError(t, err, e.err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, e.expected, actual)
			}
		})
	}
}

func mustBuildPattern(t *testing.T, pat string) pattern {
	p, err := newPattern(pat)
	if err != nil {
		t.Fatal(err)
	}
	return p
}
