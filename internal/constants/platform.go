package constants

const (
	GoogleAppEngineStdKey      = "gae_std"
	GoogleAppEngineStdPlatform = "provider:gcp/platform:appengine/type:standard"
	GoogleAppEngineStdDisplay  = "App Engine"

	GoogleComputeEngineKey      = "gce"
	GoogleComputeEnginePlatform = "provider:gcp/platform:compute/type:instances"
	GoogleComputeEngineDisplay  = "Compute Engine"
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
)
