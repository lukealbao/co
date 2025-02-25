// Package codeowners needs documentation.
package codeowners

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// LoadFileFromStandardLocation loads and parses a CODEOWNERS file at one of the
// standard locations for CODEOWNERS files (./, .github/, docs/). If run from a
// git repository, all paths are relative to the repository root.
func LoadFileFromStandardLocation() ([]Rule, error) {
	path := FindFileAtStandardLocation()
	if path == "" {
		return nil, fmt.Errorf("could not find CODEOWNERS file at any of the standard locations")
	}
	return LoadFile(path)
}

// LoadFile loads and parses a CODEOWNERS file at the path specified.
func LoadFile(path string) ([]Rule, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return ParseFile(f)
}

func LsFiles(ref string) ([]string, error) {
	var (
		files []byte
		err   error
	)
	if ref == "" {
		files, err = exec.Command("git", "ls-files").Output()
	} else {
		files, err = exec.Command("git", "ls-tree", "-r", "--name-only", ref).Output()
	}
	if err != nil {
		return nil, err
	}

	return strings.Split(strings.TrimSpace(string(files)), "\n"), nil
}

func LoadFileFromStandardLocationAtRef(ref string) ([]Rule, error) {
	if ref == "" {
		return LoadFileFromStandardLocation()
	}

	files, err := LsFiles(ref)
	if err != nil {
		return nil, err
	}

	standards := []string{"CODEOWNERS", ".github/CODEOWNERS", ".gitlab/CODEOWNERS", "docs/CODEOWNERS"}

	for _, file := range files {
		for _, known := range standards {
			if file == known {
				return LoadFileAtRef(ref, file)
			}
		}
	}

	return nil, fmt.Errorf("could not find CODEOWNERS file at any of the standard locations (ref: %s)", ref)
}

// LoadFileAtRef loads and parses a CODEOWNERS file from a historical commit. If ref is an empty string,
// file will be read from disk.
func LoadFileAtRef(ref, path string) ([]Rule, error) {
	if ref == "" {
		return LoadFile(path)
	}

	spec := fmt.Sprintf("%s:%s", ref, path)
	f, err := exec.Command("git", "show", spec).Output()
	if err != nil {
		return nil, fmt.Errorf("%s: could not load codeowners at %s", err, spec)
	}
	return ParseFile(bytes.NewReader(f))
}

// FindFileAtStandardLocation loops through the standard locations for
// CODEOWNERS files (./, .github/, docs/), and returns the first place a
// CODEOWNERS file is found. If run from a git repository, all paths are
// relative to the repository root.
func FindFileAtStandardLocation() string {
	pathPrefix := ""
	repoRoot, inRepo := findRepositoryRoot()
	if inRepo {
		pathPrefix = repoRoot
	}

	for _, path := range []string{"CODEOWNERS", ".github/CODEOWNERS", ".gitlab/CODEOWNERS", "docs/CODEOWNERS"} {
		fullPath := filepath.Join(pathPrefix, path)
		if fileExists(fullPath) {
			return fullPath
		}
	}
	return ""
}

// fileExist checks if a normal file exists at the path specified.
func fileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// findRepositoryRoot returns the path to the root of the git repository, if
// we're currently in one. If we're not in a git repository, the boolean return
// value is false.
func findRepositoryRoot() (string, bool) {
	output, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", false
	}
	return strings.TrimSpace(string(output)), true
}

const (
	// EmailOwner is the owner type for email addresses.
	EmailOwner string = "email"
	// TeamOwner is the owner type for GitHub teams.
	TeamOwner string = "team"
	// UsernameOwner is the owner type for GitHub usernames.
	UsernameOwner string = "username"
)

// Owner represents an owner found in a rule.
type Owner struct {
	// Value is the name of the owner: the email addres, team name, or username.
	Value string
	// Type will be one of 'email', 'team', or 'username'.
	Type string
}

// String returns a string representation of the owner. For email owners, it
// simply returns the email address. For user and team owners it prepends an '@'
// to the owner.
func (o Owner) String() string {
	if o.Type == EmailOwner {
		return o.Value
	}
	return "@" + o.Value
}
