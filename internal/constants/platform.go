package constants

const (
	GoogleAppEngineStdKey      = "gae_std"
	GoogleAppEngineStdPlatform = "provider:gcp/platform:appengine/type:standard"
	GoogleAppEngineStdDisplay  = "App Engine"

	GoogleComputeEngineKey      = "gce"
	GoogleComputeEnginePlatform = "provider:gcp/platform:compute/type:instances"
	GoogleComputeEngineDisplay  = "Compute Engine"

	Read  = "Read"
	Write = "Write"
)

var (
	KeyToDisplay = map[string]string{
		GoogleAppEngineStdKey:  GoogleAppEngineStdDisplay,
		GoogleComputeEngineKey: GoogleComputeEngineDisplay,
	}
	DisplayToPlatform = map[string]string{
		GoogleAppEngineStdDisplay:  GoogleAppEngineStdPlatform,
		GoogleComputeEngineDisplay: GoogleComputeEnginePlatform,
	}
	PlatformPermissions = map[string]map[string]struct{}{
		GoogleAppEngineStdPlatform: {
			Read:  {},
			Write: {},
		},
		GoogleComputeEnginePlatform: {
			Read: {},
		},
	}
)
