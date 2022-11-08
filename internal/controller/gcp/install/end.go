package gcpinstall

import (
	"bytes"
	"os/exec"
	"strings"

	"flightcrew.io/cli/internal/style"
	"flightcrew.io/cli/internal/view/command"
)

type EndController struct {
	replacer       *strings.Replacer
	endDescription string
	commands       []*command.Model
	vmIsUp         bool
}

func NewEndController(commands []*command.Model, replacer *strings.Replacer) *EndController {
	return &EndController{
		commands: commands,
		replacer: replacer,
	}
}

func (ctl EndController) Commands() []*command.Model {
	return ctl.commands
}

func (ctl EndController) Name() string {
	return "Google Cloud Platform Installation"
}

func (ctl *EndController) EndDescription() string {
	rerender := false
	if !ctl.vmIsUp {
		// Checks to see if the SSH port (22) is open for the VM. If it is, then the user
		// should be able to SSH into the machine.
		cmd := exec.Command("bash", "-c", `$(gcloud compute instances list --format="csv(NAME,EXTERNAL_IP,STATUS)" --project=${PROJECT_ID} --zones=${ZONE} | awk -F "," "/${VIRTUAL_MACHINE}/{print f(2)} function f(n){return (\$n==\"\" ? \"null\" : \$n)}" | nc -w 1 -z`)
		var b bytes.Buffer
		cmd.Stdout = &b
		cmd.Stderr = &b
		err := cmd.Run()
		if err == nil {
			ctl.vmIsUp = true
			rerender = true
		}
	}

	if !rerender && len(ctl.endDescription) > 0 {
		return ctl.endDescription
	}

	var link = "https://console.cloud.google.com/compute/instancesDetail/zones/${ZONE}/instances/${VIRTUAL_MACHINE}?project=${GOOGLE_PROJECT_ID}"
	var description = `## Welcome to Flightcrew! 🕊

${MESSAGE}

See your VM in the console:
http://replace.me

Alternatively, see your new VM in action:
${CODE_START}
# SSH into the created VM.
gcloud compute ssh ${VIRTUAL_MACHINE} --project ${GOOGLE_PROJECT_ID} --zone ${ZONE}
# Follow the new container's logs.
docker logs --follow $(docker ps -f name="${IMAGE_PATH}:${TOWER_VERSION}" --format="{{.ID}}")
${CODE_END}

Once your Tower is up, head on over to ${APP_URL} to see the info your Tower collected.
`

	link = ctl.replacer.Replace(link)
	description = ctl.replacer.Replace(description)
	description = strings.Replace(description, "${CODE_START}", "```sh", 1)
	description = strings.Replace(description, "${CODE_END}", "```", 1)
	if ctl.vmIsUp {
		description = strings.Replace(description, "${MESSAGE}", "⏱ Your VM is still starting up.", 1)
	} else {
		description = strings.Replace(description, "${MESSAGE}", "✅ Your VM is available and running!", 1)
	}

	out, _ := style.Glamour.Render(description)
	out = strings.Replace(out, "http://replace.me", link, 1)

	return out
}
