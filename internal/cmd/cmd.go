package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"

	"flightcrew.io/cli/internal/gcp"
	gcpinstall "flightcrew.io/cli/internal/gcp/install"
	"flightcrew.io/cli/internal/view"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

func init() {
	gcpinstall.RegisterFlags(gcpInstallCmd)
}

// Do runs the command logic.
func Do(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) int {
	rootCmd := &cobra.Command{Use: "flycli", SilenceUsage: true}

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

var version string

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number for Flightcrew's CLI",
	Run: func(cmd *cobra.Command, args []string) {
		// TODO(chris)
		//	if debug.Traced {
		//		defer trace.StartRegion(cmd.Context(), "version").End()
		//	}
		//	if version == "" {
		//		fmt.Printf("%s\n", info.Version)
		//	} else {
		//		fmt.Printf("%s\n", version)
		//	}
	},
}

var gcpCmd = &cobra.Command{
	Use:   "gcp",
	Short: "Manage Flightcrew for Google Cloud Platform (GCP).",
}

var gcpInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install a Flightcrew tower into Google Cloud Platform (GCP).",
	RunE: func(cmd *cobra.Command, args []string) error {
		//	if debug.Traced {
		//		defer trace.StartRegion(cmd.Context(), "compile").End()
		//	}
		//stderr := cmd.ErrOrStderr()
		//dir, name := getConfigPath(stderr, cmd.Flag("file"))
		//if _, err := Generate(cmd.Context(), ParseEnv(cmd), dir, name, stderr); err != nil {
		//	os.Exit(1)
		//}

		if err := gcp.InitArtifactRegistry(); err != nil {
			return fmt.Errorf("init artifact registry: %w", err)
		}

		env, cleanup, err := gcpinstall.ParseFlags(cmd)
		if err != nil {
			return err
		}
		defer cleanup()

		p := tea.NewProgram(view.NewInputsModel(gcpinstall.NewInputs(env)))
		if err := p.Start(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		return err
	},
}

/*
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
