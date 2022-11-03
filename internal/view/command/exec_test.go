package command

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeCommandForExec(t *testing.T) {
	type testCase struct {
		input string
		want  string
	}
	for _, tc := range []testCase{
		{
			input: `gcloud compute instances create-with-container fc-ct-gae-std-dev \
	--project=project-id-1234 \
	--container-command="/ko-app/tower" \
	--container-image="us-west1-docker.pkg.dev/flightcrew-artifacts/client/tower:0.2.15" \
	--container-arg="--debug=true" \
	--container-env="FC_API_KEY=example-token-1234" \
	--container-env="CLOUD_PLATFORM=provider:gcp/platform:compute/type:instances" \
	--container-env="FC_PACKAGE_VERSION=0.2.15" \
	--container-env="METRIC_PROVIDERS=stackdriver" \
	--container-env="FC_RPC_CONNECT_HOST=host.example.com" \
	--container-env="FC_RPC_CONNECT_PORT=443" \
	--container-env="FC_TOWER_PORT=8080" \
	--labels="component=flightcrew" \
	--machine-type="e2-micro" \
	--scopes="cloud-platform" \
	--service-account="flightcrew-runner@project-id-1234.iam.gserviceaccount.com" \
	--tags="http-server" \
	--zone="us-west2-a"`,
			want: `gcloud compute instances create-with-container fc-ct-gae-std-dev --project=project-id-1234 --container-command="/ko-app/tower" --container-image="us-west1-docker.pkg.dev/flightcrew-artifacts/client/tower:0.2.15" --container-arg="--debug=true" --container-env="FC_API_KEY=example-token-1234" --container-env="CLOUD_PLATFORM=provider:gcp/platform:compute/type:instances" --container-env="FC_PACKAGE_VERSION=0.2.15" --container-env="METRIC_PROVIDERS=stackdriver" --container-env="FC_RPC_CONNECT_HOST=host.example.com" --container-env="FC_RPC_CONNECT_PORT=443" --container-env="FC_TOWER_PORT=8080" --labels="component=flightcrew" --machine-type="e2-micro" --scopes="cloud-platform" --service-account="flightcrew-runner@project-id-1234.iam.gserviceaccount.com" --tags="http-server" --zone="us-west2-a"`,
		},
		{
			input: `gcloud compute instances list --format="csv(NAME,EXTERNAL_IP,STATUS)" \
	--organization="1234567890" \
	--zones=us-west2-a | awk -F "," "/vm-name/ {print f(2), f(3)} function f(n){return(\$n==\"\" ? \"null\" : \$n)}" | [ $(wc -c) -gt "0" ]`,
			want: `gcloud compute instances list --format="csv(NAME,EXTERNAL_IP,STATUS)" --organization="1234567890" --zones=us-west2-a | awk -F "," "/vm-name/ {print f(2), f(3)} function f(n){return(\$n==\"\" ? \"null\" : \$n)}" | [ $(wc -c) -gt "0" ]`,
		},
		{
			input: `ls`,
			want:  `ls`,
		},
	} {
		assert.Equal(t, tc.want, sanitizeForExec(tc.input))
	}

}
