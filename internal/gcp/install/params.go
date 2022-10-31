package gcpinstall

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
)

type InstallParams struct {
	VirtualMachineName  string
	ProjectName         string
	Zone                string
	TowerVersion        string
	Token               string
	IAMRole             string
	IAMFile             string
	ServiceAccount      string
	PlatformDisplayName string
	Write               bool
}
