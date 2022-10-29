package gcp

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
	ReadOnly            bool
}
