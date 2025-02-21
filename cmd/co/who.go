package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/lukealbao/co"
	"github.com/spf13/cobra"
)

func exitIf(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
}

var whoCmd = &cobra.Command{
	Use:   "who [filepath]...",
	Short: "List code owners for file(s)",
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		var filesToCheck []string

		if len(args) > 0 {
			filesToCheck = expandAllFiles(args[:])
		} else {
			filesToCheck, err = codeowners.LsFiles("HEAD")
			exitIf(err)
		}

		files, err := listOwners(sessionRules, filesToCheck, ownerFilters, showUnowned)
		exitIf(err)

		for _, result := range files {
			fmt.Fprintf(cmd.OutOrStdout(), "%-70s %s\n", result.Path, result.Owners)
		}
	},
}

func expandAllFiles(paths []string) []string {
	out := make([]string, 0)

	isDir := func(path string) bool {
		info, err := os.Stat(path)
		if os.IsNotExist(err) {
			return false
		}
		return info.IsDir()
	}

	for _, p := range paths {
		if !isDir(p) {
			out = append(out, p)
			continue
		}

		filepath.WalkDir(p, func(path string, d os.DirEntry, err error) error {
			if path == ".git" {
				return filepath.SkipDir
			}

			if !isDir(path) {
				out = append(out, path)
			}

			return nil
		})
	}

	return out
}
