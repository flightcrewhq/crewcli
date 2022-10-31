package gcpinstall

import (
	"fmt"
	"os"
	"strings"

	"flightcrew.io/cli/internal/constants"
	"flightcrew.io/cli/internal/gcp"
	"github.com/spf13/cobra"
)

var (
	// Declare the variables and then assign them in init() so that we don't have a cyclical dependency
	// since the installCmd references these variables, but we need to first instantiate the flags.
	tokenFlag, versionFlag, vmFlag, projectFlag, zoneFlag, platformFlag *string
	writeFlag                                                           *bool
)

const (
	KeyProject           = "${GOOGLE_PROJECT_ID}"
	KeyTowerVersion      = "${TOWER_VERSION}"
	KeyZone              = "${ZONE}"
	KeyVirtualMachine    = "${VIRTUAL_MACHINE}"
	KeyAPIToken          = "${API_TOKEN}"
	KeyIAMRole           = "${IAM_ROLE}"
	KeyIAMServiceAccount = "${SERVICE_ACCOUNT}"
	KeyIAMFile           = "${IAM_FILE}"
	KeyPermissions       = "${PERMISSIONS}"
	KeyRPCHost           = "${RPC_HOST}"
	KeyPlatform          = "${PLATFORM}"
	KeyTrafficRouter     = "${TRAFFIC_ROUTER}"
	KeyImagePath         = "${IMAGE_PATH}"
	KeyTempDir           = "${TEMP_DIR}"
)

type Params struct {
	args    map[string]string
	tempDir string
}

func RegisterFlags(cmd *cobra.Command) {
	tokenFlag = cmd.Flags().StringP("token", "t", "", "The Flightcrew API token to identify your organization.")
	versionFlag = cmd.Flags().StringP("version", "v", "stable", "The Flightcrew image version to install.")
	vmFlag = cmd.Flags().StringP("vm", "", "flightcrew-control-tower", "The name of the VM that will be created for the Flightcrew tower in your project.")
	writeFlag = cmd.Flags().BoolP("write", "w", false, "Whether the Flightcrew tower should be readonly (true) or read-write (false).")
	projectFlag = cmd.Flags().StringP("project", "p", "", "specify your Google Cloud Platform project name")
	zoneFlag = cmd.Flags().StringP("zone", "l", "us-central", "The zone to put your Tower in.")
	platformFlag = cmd.Flags().StringP("platform", "c", "gae_std", "specify what type of params.argsources you want to manage. (gae_std, gce)")
}

func ParseFlags(cmd *cobra.Command) (Params, func(), error) {
	params := Params{
		args: make(map[string]string),
	}

	maybeAddEnv(params.args, KeyProject, *projectFlag)
	if !contains(params.args, KeyProject) {
		if projects, err := gcp.GetProjectsFromEnvironment(); err == nil {
			maybeAddEnv(params.args, KeyProject, strings.Join(projects, ","))
		}
	}

	maybeAddEnv(params.args, KeyZone, *zoneFlag)
	maybeAddEnv(params.args, KeyTowerVersion, *versionFlag)
	maybeAddEnv(params.args, KeyAPIToken, *tokenFlag)
	maybeAddEnv(params.args, KeyVirtualMachine, *vmFlag)

	if *writeFlag {
		params.args[KeyPermissions] = constants.Write
	} else {
		params.args[KeyPermissions] = constants.Read
	}

	if displayName, ok := constants.KeyToDisplay[*platformFlag]; !ok {
		desired := make([]string, 0, len(constants.KeyToDisplay))
		for k := range constants.KeyToDisplay {
			desired = append(desired, k)
		}
		return Params{}, nil, fmt.Errorf("invalid --platform flag: %s", strings.Join(desired, ", "))
	} else {
		maybeAddEnv(params.args, KeyPlatform, displayName)
	}

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

func contains(m map[string]string, key string) bool {
	_, ok := m[key]
	return ok
}
