package gcpinstall

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"flightcrew.io/cli/internal/constants"
	"flightcrew.io/cli/internal/controller/gcp"
	gconst "flightcrew.io/cli/internal/controller/gcp/constants"
	"github.com/spf13/cobra"
)

var (
	// Declare the variables and then assign them in init() so that we don't have a cyclical dependency
	// since the installCmd references these variables, but we need to first instantiate the flags.
	tokenFlag, versionFlag, vmFlag, projectFlag, zoneFlag, platformFlag *string
	writeFlag                                                           *bool
)

var (
	allKeys = []string{
		gconst.KeyProject,
		gconst.KeyTowerVersion,
		gconst.KeyZone,
		gconst.KeyVirtualMachine,
		gconst.KeyAPIToken,
		gconst.KeyIAMServiceAccount,
		gconst.KeyPermissions,
		gconst.KeyPlatform,
		gconst.KeyGAEMaxVersionCount,
		gconst.KeyGAEMaxVersionAge,
	}
)

type Params struct {
	args    map[string]string
	tempDir string
}

func RegisterFlags(cmd *cobra.Command) {
	tokenFlag = cmd.Flags().StringP(gconst.FlagToken, "t", "", "The Flightcrew API token to identify your organization.")
	versionFlag = cmd.Flags().StringP(gconst.FlagTowerVersion, "v", "stable", "The Flightcrew image version to install.")
	vmFlag = cmd.Flags().String(gconst.FlagVirtualMachine, "flightcrew-control-tower", "The name of the VM that will be created for the Flightcrew tower in your project.")
	writeFlag = cmd.Flags().BoolP(gconst.FlagWrite, "w", false, "Whether the Flightcrew tower should be read-only (false) or read-write (true).")
	projectFlag = cmd.Flags().StringP(gconst.FlagProject, "p", "", "Specify your Google Project ID.")
	zoneFlag = cmd.Flags().StringP(gconst.FlagZone, "l", "us-central1-c", "The zone to put your Tower in.")
	platformFlag = cmd.Flags().String(gconst.FlagPlatform, "gae_std", "specify what type of cloud resources you want to manage. ('gae_std' for App Engine, 'gce' for Compute Engine)")
}

func ParseFlags(cmd *cobra.Command) (Params, func(), error) {
	if !gcp.HasGcloudInPath() {
		return Params{}, nil, errors.New("gcloud is not in path")
	}

	params := Params{
		args: make(map[string]string),
	}

	maybeAddEnv(params.args, gconst.KeyProject, *projectFlag)
	maybeAddEnv(params.args, gconst.KeyZone, *zoneFlag)
	maybeAddEnv(params.args, gconst.KeyTowerVersion, *versionFlag)
	maybeAddEnv(params.args, gconst.KeyAPIToken, *tokenFlag)
	maybeAddEnv(params.args, gconst.KeyVirtualMachine, *vmFlag)

	if *writeFlag {
		params.args[gconst.KeyPermissions] = constants.Write
	} else {
		params.args[gconst.KeyPermissions] = constants.Read
	}

	displayName, ok := constants.KeyToDisplay[*platformFlag]
	if !ok {
		desired := make([]string, 0, len(constants.KeyToDisplay))
		for k := range constants.KeyToDisplay {
			desired = append(desired, k)
		}
		return Params{}, nil, fmt.Errorf("invalid --platform flag: %s", strings.Join(desired, ", "))
	}

	maybeAddEnv(params.args, gconst.KeyPlatform, displayName)

	dir, err := os.MkdirTemp("/tmp", "flightcrew-gcp-install-*")
	if err != nil {
		return Params{}, nil, fmt.Errorf("create temp dir for installation: %v", err)
	}
	params.tempDir = dir

	return params, func() {
		err = os.RemoveAll(dir)
		if err != nil {
			fmt.Printf("delete temporary directory `%s`: %v\n", dir, err)
		}
	}, nil
}

func maybeAddEnv(m map[string]string, key, value string) {
	if len(value) > 0 {
		m[key] = value
	}
}

func recreateCommand(m map[string]string) string {
	var commandName string
	if len(os.Args) > 0 {
		commandName = os.Args[0]
	} else {
		commandName = constants.CLIName
	}

	var buf strings.Builder
	buf.WriteString(commandName)
	buf.WriteString(" gcp install")

	for flagName, keyName := range gconst.FlagToKey {
		if val, ok := m[keyName]; ok && len(val) > 0 {
			switch keyName {
			case gconst.KeyPermissions:
				if val == constants.Read {
					continue
				} else {
					val = "true"
				}
			case gconst.KeyPlatform:
				val = constants.GetPlatformKey(val)
			}

			buf.WriteString(" --")
			buf.WriteString(flagName)
			buf.WriteRune('=')
			buf.WriteString(val)
		}
	}

	return buf.String()
}
