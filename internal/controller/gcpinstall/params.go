package gcpinstall

import (
	"fmt"
	"os"
	"strings"

	"flightcrew.io/cli/internal/constants"
	"github.com/spf13/cobra"
)

var (
	// Declare the variables and then assign them in init() so that we don't have a cyclical dependency
	// since the installCmd references these variables, but we need to first instantiate the flags.
	tokenFlag, versionFlag, vmFlag, projectFlag, zoneFlag, platformFlag *string
	writeFlag                                                           *bool
)

const (
	flagProject        = "project"
	flagZone           = "zone"
	flagTowerVersion   = "version"
	flagToken          = "token"
	flagVirtualMachine = "vm"
	flagPlatform       = "platform"
	flagWrite          = "write"
)

const (
	keyProject            = "${GOOGLE_PROJECT_ID}"
	keyTowerVersion       = "${TOWER_VERSION}"
	keyZone               = "${ZONE}"
	keyVirtualMachine     = "${VIRTUAL_MACHINE}"
	keyAPIToken           = "${API_TOKEN}"
	keyIAMRole            = "${IAM_ROLE}"
	keyIAMServiceAccount  = "${SERVICE_ACCOUNT}"
	keyIAMFile            = "${IAM_FILE}"
	keyPermissions        = "${PERMISSIONS}"
	keyRPCHost            = "${RPC_HOST}"
	keyAppURL             = "${APP_URL}"
	keyPlatform           = "${PLATFORM}"
	keyTrafficRouter      = "${TRAFFIC_ROUTER}"
	keyImagePath          = "${IMAGE_PATH}"
	keyTempDir            = "${TEMP_DIR}"
	keyGAEMaxVersionCount = "${GAE_MAX_VERSION_COUNT}"
	keyGAEMaxVersionAge   = "${GAE_MAX_VERSION_AGE}"
)

var (
	flagToKey = map[string]string{
		flagProject:        keyProject,
		flagVirtualMachine: keyVirtualMachine,
		flagPlatform:       keyPlatform,
		flagToken:          keyAPIToken,
		flagTowerVersion:   keyTowerVersion,
		flagZone:           keyZone,
		flagWrite:          keyPermissions,
	}
)

var (
	allKeys = []string{
		keyProject,
		keyTowerVersion,
		keyZone,
		keyVirtualMachine,
		keyAPIToken,
		keyIAMServiceAccount,
		keyPermissions,
		keyPlatform,
		keyGAEMaxVersionCount,
		keyGAEMaxVersionAge,
	}
)

type Params struct {
	args    map[string]string
	tempDir string
}

func RegisterFlags(cmd *cobra.Command) {
	tokenFlag = cmd.Flags().StringP(flagToken, "t", "", "The Flightcrew API token to identify your organization.")
	versionFlag = cmd.Flags().StringP(flagTowerVersion, "v", "stable", "The Flightcrew image version to install.")
	vmFlag = cmd.Flags().String(flagVirtualMachine, "flightcrew-control-tower", "The name of the VM that will be created for the Flightcrew tower in your project.")
	writeFlag = cmd.Flags().BoolP(flagWrite, "w", false, "Whether the Flightcrew tower should be read-only (false) or read-write (true).")
	projectFlag = cmd.Flags().StringP(flagProject, "p", "", "Specify your Google Project ID.")
	zoneFlag = cmd.Flags().StringP(flagZone, "l", "us-central1-c", "The zone to put your Tower in.")
	platformFlag = cmd.Flags().String(flagPlatform, "gae_std", "specify what type of cloud resources you want to manage. ('gae_std' for App Engine, 'gce' for Compute Engine)")
}

func ParseFlags(cmd *cobra.Command) (Params, func(), error) {
	params := Params{
		args: make(map[string]string),
	}

	maybeAddEnv(params.args, keyProject, *projectFlag)
	maybeAddEnv(params.args, keyZone, *zoneFlag)
	maybeAddEnv(params.args, keyTowerVersion, *versionFlag)
	maybeAddEnv(params.args, keyAPIToken, *tokenFlag)
	maybeAddEnv(params.args, keyVirtualMachine, *vmFlag)

	if *writeFlag {
		params.args[keyPermissions] = constants.Write
	} else {
		params.args[keyPermissions] = constants.Read
	}

	displayName, ok := constants.KeyToDisplay[*platformFlag]
	if !ok {
		desired := make([]string, 0, len(constants.KeyToDisplay))
		for k := range constants.KeyToDisplay {
			desired = append(desired, k)
		}
		return Params{}, nil, fmt.Errorf("invalid --platform flag: %s", strings.Join(desired, ", "))
	}

	maybeAddEnv(params.args, keyPlatform, displayName)

	dir, err := os.MkdirTemp("", "flightcrew-gcp-install-*")
	if err != nil {
		return Params{}, nil, fmt.Errorf("create temp dir for installation: %v", err)
	}

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

	for flagName, keyName := range flagToKey {
		if val, ok := m[keyName]; ok {
			switch keyName {
			case keyPermissions:
				if val == constants.Read {
					continue
				} else {
					val = "true"
				}
			case keyPlatform:
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
