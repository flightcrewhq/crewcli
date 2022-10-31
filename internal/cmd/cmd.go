package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"flightcrew.io/cli/internal/constants"
	"flightcrew.io/cli/internal/gcp"
	gcpview "flightcrew.io/cli/internal/gcp/view"
	"flightcrew.io/cli/internal/view"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var (
	// Declare the variables and then assign them in init() so that we don't have a cyclical dependency
	// since the installCmd references these variables, but we need to first instantiate the flags.
	tokenFlag, versionFlag, vmFlag, projectFlag, zoneFlag, platformFlag *string
	writeFlag                                                           *bool
)

func init() {
	tokenFlag = gcpInstallCmd.Flags().StringP("token", "t", "", "The Flightcrew API token to identify your organization.")
	versionFlag = gcpInstallCmd.Flags().StringP("version", "v", "stable", "The Flightcrew image version to install.")
	vmFlag = gcpInstallCmd.Flags().StringP("vm", "", "flightcrew-control-tower", "The name of the VM that will be created for the Flightcrew tower in your project.")
	writeFlag = gcpInstallCmd.Flags().BoolP("write", "w", false, "Whether the Flightcrew tower should be readonly (true) or read-write (false).")
	projectFlag = gcpInstallCmd.Flags().StringP("project", "p", "", "specify your Google Cloud Platform project name")
	zoneFlag = gcpInstallCmd.Flags().StringP("zone", "l", "us-central", "The zone to put your Tower in.")
	platformFlag = gcpInstallCmd.Flags().StringP("platform", "c", "gae_std", "specify what type of resources you want to manage. (gae_std, gce)")
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

type Env struct {
	ExperimentalFeatures bool
	DryRun               bool
}

func ParseEnv(c *cobra.Command) Env {
	x := c.Flag("experimental")
	dr := c.Flag("dry-run")
	return Env{
		ExperimentalFeatures: x != nil && x.Changed,
		DryRun:               dr != nil && dr.Changed,
	}
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

		params := gcp.InstallParams{}

		params.ProjectName = *projectFlag
		if len(params.ProjectName) == 0 {
			if projects, err := gcp.GetProjectsFromEnvironment(); err == nil {
				params.ProjectName = strings.Join(projects, ",")
			}
		}
		params.Zone = *zoneFlag
		params.TowerVersion = *versionFlag
		params.Token = *tokenFlag
		params.Write = *writeFlag
		params.VirtualMachineName = *vmFlag

		displayName, ok := constants.KeyToDisplay[*platformFlag]
		if !ok {
			desired := make([]string, 0, len(constants.KeyToDisplay))
			for k := range constants.KeyToDisplay {
				desired = append(desired, k)
			}
			return fmt.Errorf("invalid --platform flag: %s", strings.Join(desired, ", "))
		}

		params.PlatformDisplayName = displayName

		dir, err := os.MkdirTemp("", "flightcrew-install-*")
		if err != nil {
			return fmt.Errorf("create temp dir for installation: %v", err)
		}
		defer func() {
			err = os.RemoveAll(dir)
		}()

		p := tea.NewProgram(view.NewInstallModel(
			gcpview.NewInstallInputs(params, dir),
			gcpview.NewInstallCommands))
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
