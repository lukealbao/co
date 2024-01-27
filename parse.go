package codeowners

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strings"
)

var (
	commentRegexp  = regexp.MustCompile(`\A\s*#`)
	emailRegexp    = regexp.MustCompile(`\A[A-Z0-9a-z\._%\+\-]+@[A-Za-z0-9\.\-]+\.[A-Za-z]{2,6}\z`)
	teamRegexp     = regexp.MustCompile(`\A@([a-zA-Z0-9\-]+\/[a-zA-Z0-9_\-]+)\z`)
	usernameRegexp = regexp.MustCompile(`\A@([a-zA-Z0-9\-]+)\z`)
)

const (
	statePattern = iota + 1
	stateOwners
)

// ParseFile parses a CODEOWNERS file, returning a set of rules.
func ParseFile(f io.Reader) ([]Rule, error) {
	rules := make([]Rule, 0)
	scanner := bufio.NewScanner(f)

	r := newRule()

	for lineNo := 1; scanner.Scan(); lineNo++ {
		line := scanner.Text()

		if line == "" || commentRegexp.MatchString(line) {
			r.leadingComment += line + "\n"
			continue
		}

		if err := parseRule(line, r); err != nil {
			return rules, fmt.Errorf("line %d: %v", lineNo, err)
		} else {
			r.SourceLine = lineNo
			rules = append(rules, *r)
			r = newRule()
		}
	}

	return rules, nil
}

func parseRule(ruleStr string, r *Rule) error {
	state := statePattern
	escaped := false

	buf := bytes.Buffer{}
	for i, ch := range strings.TrimSpace(ruleStr) {
		// Comments consume the rest of the line and stop further parsing
		if ch == '#' {
			r.trailingComment = strings.TrimSpace(ruleStr[i:])
			break
		}

		switch state {
		case statePattern:
			switch {
			case ch == '\\':
				// Escape the next character (important for whitespace while parsing), but
				// don't lose the backslash as it's part of the pattern
				escaped = true
				buf.WriteRune(ch)
				continue

			case isWhitespace(ch) && !escaped:
				// Unescaped whitespace means this is the end of the pattern
				pattern, err := newPattern(buf.String())
				if err != nil {
					return err
				}
				r.pattern = pattern
				buf.Reset()
				state = stateOwners

			case isPatternChar(ch) || (isWhitespace(ch) && escaped):
				// Keep any valid pattern characters and escaped whitespace
				buf.WriteRune(ch)

			default:
				return fmt.Errorf("unexpected character '%c' at position %d", ch, i+1)
			}
			// Escaping only applies to one character
			escaped = false

		case stateOwners:
			switch {
			case isWhitespace(ch):
				// Whitespace means we've reached the end of the owner or we're just chomping
				// through whitespace before or after owner declarations
				if buf.Len() > 0 {
					ownerStr := buf.String()
					owner, err := newOwner(ownerStr)
					if err != nil {
						return fmt.Errorf("%s at position %d", err.Error(), i+1-len(ownerStr))
					}
					r.Owners = append(r.Owners, owner.String())
					buf.Reset()
				}

			case isOwnersChar(ch):
				// Write valid owner characters to the buffer
				buf.WriteRune(ch)

			default:
				return fmt.Errorf("unexpected character '%c' at position %d", ch, i+1)
			}
		}
	}

	// We've finished consuming the line, but we might still have content in the buffer
	// if the line didn't end with a separator (whitespace)
	switch state {
	case statePattern:
		if buf.Len() == 0 { // We should have non-empty pattern
			return fmt.Errorf("unexpected end of rule")
		}

		pattern, err := newPattern(buf.String())
		if err != nil {
			return err
		}
		r.pattern = pattern

	case stateOwners:
		// If there's an owner left in the buffer, don't leave it behind
		if buf.Len() > 0 {
			ownerStr := buf.String()
			owner, err := newOwner(ownerStr)
			if err != nil {
				return fmt.Errorf("%s at position %d", err.Error(), len(ruleStr)+1-len(ownerStr))
			}
			r.Owners = append(r.Owners, owner.String())
		}
	}

	return nil
}

// newOwner figures out which kind of owner this is and returns an Owner struct
func newOwner(s string) (Owner, error) {
	match := emailRegexp.FindStringSubmatch(s)
	if match != nil {
		return Owner{Value: match[0], Type: EmailOwner}, nil
	}

	match = teamRegexp.FindStringSubmatch(s)
	if match != nil {
		return Owner{Value: match[1], Type: TeamOwner}, nil
	}

	match = usernameRegexp.FindStringSubmatch(s)
	if match != nil {
		return Owner{Value: match[1], Type: UsernameOwner}, nil
	}

	return Owner{}, fmt.Errorf("invalid owner format '%s'", s)
}

func isWhitespace(ch rune) bool {
	return ch == ' ' || ch == '\t' || ch == '\n'
}

func isAlphanumeric(ch rune) bool {
	return (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9')
}

// isPatternChar matches characters that are allowed in patterns
func isPatternChar(ch rune) bool {
	switch ch {
	case '*', '?', '.', '/', '@', '_', '+', '-', '\\':
		return true
	}
	return isAlphanumeric(ch)
}

// isOwnersChar matches characters that are allowed in owner definitions
func isOwnersChar(ch rune) bool {
	switch ch {
	case '.', '@', '/', '_', '%', '+', '-':
		return true
	}
	return isAlphanumeric(ch)
}
