package cmd

import (
	"fmt"

	internalversion "flightcrew.io/cli/internal/version"
	"github.com/spf13/cobra"
)

// These variables are stamped by ldflags on build, configured through goreleaser.
// In the event that they aren't available on init(), the values will be populated
// through debug.ReadBuildInfo in internal/version.
var (
	version   = ""
	commit    = ""
	date      = ""
	goVersion = ""
)

func init() {
	if version == "" {
		version = internalversion.Version()
	} else {
		internalversion.MarkModified(&version)
		internalversion.SetFromCmd(version)
	}

	if commit == "" {
		commit = internalversion.Commit()
	}

	if date == "" {
		date = internalversion.Date()
	}

	if goVersion == "" {
		goVersion = internalversion.GoVersion()
	}
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print build version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("version: %s (%s)\ncommit: %s\ndate: %s\n", version, goVersion, commit, date)
	},
}
