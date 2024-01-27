package codeowners

import (
	"fmt"
	"strings"

	"github.com/google/btree"
)

// Rule is a CODEOWNERS rule that maps a gitignore-style path pattern to a set of owners.
type Rule struct {
	SourceLine      int
	leadingComment  string
	trailingComment string
	pattern         pattern
	Owners          []string
}

// RawPattern returns the rule's gitignore-style path pattern.
func (r *Rule) RawPattern() string {
	if r == nil {
		return ""
	}
	return r.pattern.pattern
}

// Match tests whether the provided matches the rule's pattern.
func (r *Rule) Match(path string) (bool, error) {
	return r.pattern.match(path)
}

func (r *Rule) String() string {
	var b strings.Builder
	if r.leadingComment != "" {
		b.WriteString(r.leadingComment)
	}
	b.WriteString(r.pattern.pattern)
	for _, owner := range r.Owners {
		b.WriteString(" " + owner)
	}
	if r.trailingComment != "" {
		b.WriteString(" " + r.trailingComment)
	}

	return b.String()
}

// Ruleset is a collection of CODEOWNERS rules.
type Ruleset []Rule

// Match finds the last rule in the ruleset that matches the path provided. When determining the
// ownership of a file using CODEOWNERS, order matters, and the last matching rule takes precedence.
func (r Ruleset) Match(path string) (*Rule, error) {
	for i := len(r) - 1; i >= 0; i-- {
		rule := &r[i]
		match, err := rule.Match(path)
		if match || err != nil {
			return rule, err
		}
	}
	return nil, nil
}

func newRule() *Rule {
	r := Rule{
		Owners: make([]string, 0),
	}
	return &r
}

type FileTree struct {
	rules map[string]Rule
	index *btree.BTreeG[string]
}

// String will print all rules in lexicographical order.
func (f *FileTree) String() string {
	var b strings.Builder
	f.index.Ascend(func(name string) bool {
		rule, ok := f.rules[name]
		if !ok {
			panic(fmt.Errorf("unexpectedly missing rule(%s)", name))
		}
		b.WriteString(rule.String() + "\n")
		return true
	})

	return b.String()
}

func NewFileTree(rules []Rule) *FileTree {
	tree := FileTree{make(map[string]Rule), btree.NewG[string](10, func(a, b string) bool {
		return a < b
	})}

	for _, r := range rules {
		tree.rules[r.pattern.pattern] = r
		tree.index.ReplaceOrInsert(r.pattern.pattern)
	}

	return &tree
}

func ConsolidateTree(tree *FileTree) {
	sameSliceContents := func(s1, s2 []string) bool {
		if len(s1) != len(s2) {
			return false
		}

		for _, v1 := range s1 {
			var found bool
			for _, v2 := range s2 {
				if v1 == v2 {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}

		return true
	}

	// Traverse the entire tree in lexicographical order. For each node, treat it as a directory
	// and traverse its subtree. If all descendants have the same owner, remove them from the tree.
	tree.index.Ascend(func(root string) bool {
		var toRemove []string

		rule, ok := tree.rules[root]
		if !ok {
			panic(fmt.Errorf("cannot traverse %s", root))
		}

		var start string
		if root[len(root)-1] == '/' {
			start = root
		} else {
			start = root + "/"
		}

		tree.index.AscendGreaterOrEqual(start, func(pat string) bool {
			if rule.RawPattern() == pat {
				return true
			}

			if ok, _ := rule.Match(pat); !ok {
				return true
			}

			if !sameSliceContents(tree.rules[pat].Owners, tree.rules[root].Owners) {
				toRemove = nil
				return false
			}

			toRemove = append(toRemove, pat)
			return true
		})

		for _, descendant := range toRemove {
			delete(tree.rules, descendant)
			tree.index.Delete(descendant)
		}

		return true
	})
}
