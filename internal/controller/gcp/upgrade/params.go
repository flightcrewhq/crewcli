package gcpupgrade

import (
	"errors"
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
	versionFlag, vmFlag, projectFlag, zoneFlag *string
)

var (
	allKeys = []string{
		gconst.KeyProject,
		gconst.KeyTowerVersion,
		gconst.KeyZone,
		gconst.KeyVirtualMachine,
	}
)

type Params struct {
	args map[string]string
}

func RegisterFlags(cmd *cobra.Command) {
	versionFlag = cmd.Flags().StringP(gconst.FlagTowerVersion, "v", "stable", "The Flightcrew image version to install.")
	vmFlag = cmd.Flags().String(gconst.FlagVirtualMachine, "flightcrew-control-tower", "The name of the VM that will be created for the Flightcrew tower in your project.")
	projectFlag = cmd.Flags().StringP(gconst.FlagProject, "p", "", "Specify your Google Project ID.")
	zoneFlag = cmd.Flags().StringP(gconst.FlagZone, "l", "us-central1-c", "The zone to put your Tower in.")
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
	maybeAddEnv(params.args, gconst.KeyVirtualMachine, *vmFlag)

	return params, func() {}, nil
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
	buf.WriteString(" gcp upgrade")

	for flagName, keyName := range gconst.FlagToKey {
		if val, ok := m[keyName]; ok && len(val) > 0 {
			buf.WriteString(" --")
			buf.WriteString(flagName)
			buf.WriteRune('=')
			buf.WriteString(val)
		}
	}

	return buf.String()
}
