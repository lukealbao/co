package codeowners

import (
	"sort"
)

type r struct {
	Path   string   `json:"path"`
	Owners []string `json:"owners"`
}

type Owners []*r

func (x Owners) Len() int           { return len(x) }
func (x Owners) Less(a, b int) bool { return x[a].Path < x[b].Path }
func (x Owners) Swap(a, b int) {
	x[a], x[b] = x[b], x[a]
}

var _ sort.Interface = (Owners)(nil)

// ListOwners returns a list of structured output. Callers must format for printing.
func ListOwners(rules Ruleset, files []string, ownerFilters []string, showUnowned bool) (Owners, error) {
	var out []*r = make([]*r, 0)

	for _, file := range files {
		rule, err := rules.Match(file)
		if err != nil {
			return nil, err
		}

		if rule == nil || rule.Owners == nil || len(rule.Owners) == 0 {
			if len(ownerFilters) == 0 || showUnowned {
				out = append(out, &r{Path: file, Owners: []string{"(unowned)"}})
			}

			continue
		}

		owners := make([]string, 0, len(rule.Owners))
		for _, owner := range rule.Owners {
			filterMatch := len(ownerFilters) == 0 && !showUnowned
			for _, filter := range ownerFilters {
				if filter == owner { // TODO: This is "Value" in hmarr. Are we losing info?
					filterMatch = true
				}
			}
			if filterMatch {
				owners = append(owners, owner)
			}
		}

		if len(owners) > 0 {
			out = append(out, &r{Path: file, Owners: owners})
		}
	}

	return out, nil
}
