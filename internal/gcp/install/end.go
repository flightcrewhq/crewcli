package gcpinstall

import (
	"strings"

	"flightcrew.io/cli/internal/style"
)

func (inputs Inputs) EndDescription() string {
	if len(inputs.endDescription) > 0 {
		return inputs.endDescription
	}

	var description = `## 🕊  Welcome to Flightcrew!

GCE takes a couple of minutes to start up the VM.

See your VM in the console:
https://console.cloud.google.com/compute/instancesDetail/zones/${ZONE}/instances/${VIRTUAL_MACHINE}?project=${GOOGLE_PROJECT_ID}

Alternatively, see your new VM in action:
${CODE_START}
# SSH into the created VM.
gcloud compute ssh ${VIRTUAL_MACHINE} --project ${GOOGLE_PROJECT_ID} --zone ${ZONE}
# Follow the new container's logs.
docker logs --follow \$(docker ps -f name=tower --format=\"{{.ID}}\")
${CODE_END}

Head on over to ${APP_URL} to see the info your Tower collected.
`

	description = inputs.replacer.Replace(description)
	description = strings.Replace(description, "${CODE_START}", "```sh", -1)
	description = strings.Replace(description, "${CODE_END}", "```", -1)

	out, _ := style.Glamour.Render(description)
	return out
}
