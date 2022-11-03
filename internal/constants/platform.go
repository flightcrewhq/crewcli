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

type Permissions struct {
	Content string
	Role    string
}

var (
	KeyToDisplay = map[string]string{
		GoogleAppEngineStdKey:  GoogleAppEngineStdDisplay,
		GoogleComputeEngineKey: GoogleComputeEngineDisplay,
	}
	DisplayToPlatform = map[string]string{
		GoogleAppEngineStdDisplay:  GoogleAppEngineStdPlatform,
		GoogleComputeEngineDisplay: GoogleComputeEnginePlatform,
	}
	// TODO: Use a local file or get from a URL that the user provides.
	PlatformPermissions = map[string]map[string]*Permissions{
		GoogleAppEngineStdPlatform: {
			Read: &Permissions{
				Role: "flightcrew.gae_std.read.only",
				Content: `title: Flightcrew AppEngine (Read-Only)
description: Grants Flightcrew's Control Tower VM read access to AppEngine configs and monitoring.
stage: ALPHA
# These permissions are pared down from those specified in the GCP docs
# for how to read AppEngine versions:
# https://cloud.google.com/iam/docs/understanding-roles#app-engine-roles
includedPermissions:
# Read the Application config
- appengine.applications.get
- appengine.operations.get

# Read the Service config
- appengine.services.get
- appengine.services.list

# Read the Version config
- appengine.versions.get
- appengine.versions.list

# Read monitoring data
- monitoring.metricDescriptors.get
- monitoring.metricDescriptors.list
- monitoring.timeSeries.list`,
			},

			Write: &Permissions{
				Role: "flightcrew.gae_std.read.write",
				Content: `title: Flightcrew AppEngine (Write)
description: Grants Flightcrew's Control Tower VM write access to AppEngine configs and monitoring.
stage: ALPHA
# These permissions add to the Read-Only role to deploy new AppEngine versions.
includedPermissions:
# Change the service's traffic splitting:
- appengine.services.update

# CRUD for Versions
- appengine.versions.create
- appengine.versions.update
- appengine.versions.delete

# CloudBuild is needed to create new versions.
# These permissions come from Cloud Build Editor and Cloud Storage Object Admin:
# https://cloud.google.com/iam/docs/understanding-roles#cloudbuild.builds.editor
- cloudbuild.builds.create
- cloudbuild.builds.get
- storage.buckets.get
- storage.objects.create
- storage.objects.get
- storage.objects.list
- storage.objects.delete
- logging.logEntries.create
- iam.serviceAccounts.actAs`,
			},
		},

		GoogleComputeEnginePlatform: {
			Read: &Permissions{
				Role: "flightcrew.gce.read.only",
				Content: `title: Flightcrew GCE (Read-Only)
description: Grants Flightcrew read access to GCE VM instance configs and monitoring.
stage: ALPHA
# These permissions are pared down from those specified in the GCP docs:
# https://cloud.google.com/iam/docs/understanding-roles#compute-engine-roles
includedPermissions:
# Read VM specs and configs
- compute.zones.list
- compute.instances.list

# Read monitoring data
- monitoring.metricDescriptors.get
- monitoring.metricDescriptors.list
- monitoring.timeSeries.list`,
			},
		},
	}
)

func GetPlatformKey(text string) string {
	if _, ok := KeyToDisplay[text]; ok {
		return text
	}

	if _, ok := DisplayToPlatform[text]; ok {
		for key, display := range KeyToDisplay {
			if display == text {
				return key
			}
		}
	}

	for display, platform := range DisplayToPlatform {
		if platform == text {
			for key, maybeDisplay := range KeyToDisplay {
				if display == maybeDisplay {
					return key
				}
			}
		}
	}

	return ""
}
