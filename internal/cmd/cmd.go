package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"

	"flightcrew.io/cli/internal/constants"
	"flightcrew.io/cli/internal/controller/gcpinstall"
	"flightcrew.io/cli/internal/gcp"
	"flightcrew.io/cli/internal/view"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

func init() {
	gcpinstall.RegisterFlags(gcpInstallCmd)
}

// Do runs the command logic.
func Do(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) int {
	rootCmd := &cobra.Command{Use: constants.CLIName, SilenceUsage: true}

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(gcpCmd)
	//	rootCmd.AddCommand(gcpUpgradeCmd)
	//	rootCmd.AddCommand(gcpUninstallCmd)

	gcpCmd.AddCommand(gcpInstallCmd)

	rootCmd.SetArgs(args)
	rootCmd.SetIn(stdin)
	rootCmd.SetOut(stdout)
	rootCmd.SetErr(stderr)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var err error
	if err := rootCmd.ExecuteContext(ctx); err == nil {
		return 0
	}
	if exitError, ok := err.(*exec.ExitError); ok {
		return exitError.ExitCode()
	}
	return 1
}

var gcpCmd = &cobra.Command{
	Use:   "gcp",
	Short: "Manage Flightcrew for Google Cloud Platform (GCP).",
}

var gcpInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install a Flightcrew tower into Google Cloud Platform (GCP).",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !gcp.HasGcloudInPath() {
			fmt.Printf(`The "gcloud" CLI tool is a pre-requisite to run this script.

If you haven't yet, please install the tool: https://cloud.google.com/sdk/docs/install

If you already have, please add it to your path:
  export PATH=<where it is>:$PATH

`)
			return errors.New("gcloud is not in path")
		}

		if err := gcp.InitArtifactRegistry(); err != nil {
			return fmt.Errorf("init artifact registry: %w", err)
		}

		env, cleanup, err := gcpinstall.ParseFlags(cmd)
		if err != nil {
			return err
		}
		defer cleanup()

		p := tea.NewProgram(view.NewInputsModel(gcpinstall.NewInputsController(env)))
		if err := p.Start(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		return err
	},
}

/* TODO: Uncomment these when we need to implement them.
var gcpUpgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "upgrade an existing Flightcrew tower in Google Cloud Provider (GCP) to another version.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("upgrade is not yet implemented")
	},
}

var gcpUninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstalls an existing Flightcrew tower and associated resources from Google Cloud Provider (GCP).",
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("uninstall is not yet implemented")
	},
}
*/
