package gcpupgrade

import (
	"strings"

	"flightcrew.io/cli/internal/controller"
	"flightcrew.io/cli/internal/view/command"
)

type RunController struct {
	args     map[string]string
	replacer *strings.Replacer
	commands []*command.Model
}

func NewRunController(args map[string]string) *RunController {
	commands := make([]*command.Model, 0)
	commands = append(commands, getVMCommands(args)...)

	replaceArgs := make([]string, 0, 2*len(args))
	for key, arg := range args {
		replaceArgs = append(replaceArgs, key, arg)
	}

	replacer := strings.NewReplacer(replaceArgs...)
	for _, cmd := range commands {
		cmd.Replace(replacer)
	}

	return &RunController{
		args:     args,
		replacer: replacer,
		commands: commands,
	}
}

func (ctl RunController) Commands() []*command.Model {
	return ctl.commands
}

func (ctl *RunController) GetEndController() controller.End {
	return NewEndController(ctl.commands, ctl.replacer)
}

func (ctl RunController) RecreateCommand() string {
	return recreateCommand(ctl.args)
}

func getVMCommands(args map[string]string) []*command.Model {
	commands := make([]*command.Model, 0)

	checkSSH := command.NewReadModel(command.Opts{
		Command:     `RETURN=$(nc -w 1 -z "${VIRTUAL_MACHINE_IP}" 22) if [[ $RETURN -eq 0 ]]; then exit 1; else exit $RETURN; fi`,
		Description: "This command checks whether the virtual machine is reachable through SSH.",
		Message: map[command.State]string{
			command.PassState: "VM is unavailable for SSH access, skipping image pruning.",
			command.FailState: "SSH access is available. Next step is to prune images from the machine.",
		},
	})

	commands = append(commands,
		checkSSH,
		command.NewWriteModel(command.Opts{
			SkipIfSucceed: checkSSH,
			Command: `gcloud compute ssh ${VIRTUAL_MACHINE} \
	--project ${GOOGLE_PROJECT_ID} \
	--zone ${ZONE} \
	--command 'docker system prune -f -a'`,
			Description: "This command prunes old images on the virtual machine (through SSH) to save space before downloading a new one.",
		}),
		// update-container keeps the existing args / envs from when the VM was created,
		// but they can be overwritten here.
		command.NewWriteModel(command.Opts{
			Command: `gcloud compute instances update-container ${VIRTUAL_MACHINE} \
	--project=${GOOGLE_PROJECT_ID} \
	--zone=${ZONE} \
	--container-image="${IMAGE_PATH}:${TOWER_VERSION}" \
	--container-env="FC_PACKAGE_VERSION=${TOWER_VERSION}"`,
			Description: "This command updates the VM to the newest stable Control Tower image.",
		}),
	)
	return commands
}
