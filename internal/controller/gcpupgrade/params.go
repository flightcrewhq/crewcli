package gcpupgrade

import (
	"os"
	"strings"

	"flightcrew.io/cli/internal/constants"
	"github.com/spf13/cobra"
)

var (
	// Declare the variables and then assign them in init() so that we don't have a cyclical dependency
	// since the installCmd references these variables, but we need to first instantiate the flags.
	versionFlag, vmFlag, projectFlag, zoneFlag *string
)

const (
	flagProject        = "project"
	flagZone           = "zone"
	flagTowerVersion   = "version"
	flagVirtualMachine = "vm"
)

const (
	keyProject           = "${GOOGLE_PROJECT_ID}"
	keyTowerVersion      = "${TOWER_VERSION}"
	keyZone              = "${ZONE}"
	keyVirtualMachine    = "${VIRTUAL_MACHINE}"
	keyIAMServiceAccount = "${SERVICE_ACCOUNT}"
	keyTrafficRouter     = "${TRAFFIC_ROUTER}"
	keyImagePath         = "${IMAGE_PATH}"
	keyTempDir           = "${TEMP_DIR}"
	keyProjectOrOrgFlag  = "${PROJECT_OR_ORG_FLAG}"
	keyProjectOrOrgSlash = "${PROJECT_OR_ORG_SLASH}"
	keyVirtualMachineIP  = "${VIRTUAL_MACHINE_IP}"
)

var (
	flagToKey = map[string]string{
		flagProject:        keyProject,
		flagVirtualMachine: keyVirtualMachine,
		flagTowerVersion:   keyTowerVersion,
		flagZone:           keyZone,
	}
)

var (
	allKeys = []string{
		keyProject,
		keyTowerVersion,
		keyZone,
		keyVirtualMachine,
	}
)

type Params struct {
	args map[string]string
}

func RegisterFlags(cmd *cobra.Command) {
	versionFlag = cmd.Flags().StringP(flagTowerVersion, "v", "stable", "The Flightcrew image version to install.")
	vmFlag = cmd.Flags().String(flagVirtualMachine, "flightcrew-control-tower", "The name of the VM that will be created for the Flightcrew tower in your project.")
	projectFlag = cmd.Flags().StringP(flagProject, "p", "", "Specify your Google Project ID.")
	zoneFlag = cmd.Flags().StringP(flagZone, "l", "us-central1-c", "The zone to put your Tower in.")
}

func ParseFlags(cmd *cobra.Command) (Params, func(), error) {
	params := Params{
		args: make(map[string]string),
	}

	maybeAddEnv(params.args, keyProject, *projectFlag)
	maybeAddEnv(params.args, keyZone, *zoneFlag)
	maybeAddEnv(params.args, keyTowerVersion, *versionFlag)
	maybeAddEnv(params.args, keyVirtualMachine, *vmFlag)

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

	for flagName, keyName := range flagToKey {
		if val, ok := m[keyName]; ok && len(val) > 0 {
			buf.WriteString(" --")
			buf.WriteString(flagName)
			buf.WriteRune('=')
			buf.WriteString(val)
		}
	}

	return buf.String()
}
