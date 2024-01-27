package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print code version",
	Run: func(cmd *cobra.Command, files []string) {
		buildinfo := "About: https://github.com/lukealbao/co\n" +
			"Version: " + version + "\n" +
			"Built: " + date + "\n" +
			"Commit: " + commit
		fmt.Fprintln(cmd.OutOrStdout(), buildinfo)
	},
}
