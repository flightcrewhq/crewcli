package gcpinstall

import (
	"strings"

	"flightcrew.io/cli/internal/constants"
	"flightcrew.io/cli/internal/controller"
	gconst "flightcrew.io/cli/internal/controller/gcp/constants"
	"flightcrew.io/cli/internal/view/command"
)

type RunController struct {
	args     map[string]string
	replacer *strings.Replacer
	commands []*command.Model
}

func NewRunController(args map[string]string) *RunController {
	commands := make([]*command.Model, 0)
	commands = append(commands, getIAMRoleCommands(args)...)
	commands = append(commands, getServiceAccountCommands(args)...)
	commands = append(commands, getBindIAMPolicyCommands(args)...)
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

func getIAMRoleCommands(args map[string]string) []*command.Model {
	newCheckIAMRole := func(replacer *strings.Replacer) *command.Model {
		cmd := command.NewReadModel(command.Opts{
			Description: "Check if a ${PERMISSIONS} Flightcrew IAM Role already exists or needs to be created.",
			Command: `gcloud iam roles describe \${PROJECT_OR_ORG_FLAG}
	"${ROLE}" >/dev/null 2>&1`,
			Message: map[command.State]string{
				command.PassState: "This Flightcrew IAM role already exists.",
				command.FailState: "No IAM role found. Next step is to create one.",
			},
		})
		cmd.Replace(replacer)
		return cmd
	}
	newCreateIAMRole := func(replacer *strings.Replacer, skipIfSucceed *command.Model) *command.Model {
		cmd := command.NewWriteModel(command.Opts{
			SkipIfSucceed: skipIfSucceed,
			Description:   "This command creates a ${PERMISSIONS} IAM role from `${FILE}` for the Flightcrew VM to access configs and monitoring data.\n\nhttps://cloud.google.com/iam/docs/understanding-custom-roles",
			Command: `gcloud iam roles create ${ROLE} \${PROJECT_OR_ORG_FLAG}
	--file=${FILE}`,
		})
		cmd.Replace(replacer)
		return cmd
	}

	commands := make([]*command.Model, 0)

	readReplacer := strings.NewReplacer(
		"${ROLE}", args[gconst.KeyIAMRoleRead],
		"${FILE}", args[gconst.KeyIAMFileRead],
		"${PERMISSIONS}", constants.Read)
	checkReadIAMRole := newCheckIAMRole(readReplacer)
	commands = append(commands,
		checkReadIAMRole,
		newCreateIAMRole(readReplacer, checkReadIAMRole))

	if args[gconst.KeyPermissions] == constants.Write {
		writeReplacer := strings.NewReplacer(
			"${ROLE}", args[gconst.KeyIAMRoleWrite],
			"${FILE}", args[gconst.KeyIAMFileWrite],
			"${PERMISSIONS}", constants.Write)
		checkWriteIAMRole := newCheckIAMRole(writeReplacer)
		commands = append(commands,
			checkWriteIAMRole,
			newCreateIAMRole(writeReplacer, checkWriteIAMRole))
	}

	return commands
}

func getServiceAccountCommands(args map[string]string) []*command.Model {
	checkServiceAccount := command.NewReadModel(command.Opts{
		Description: "Check if a Flightcrew service account already exists or needs to be created.",
		Command:     `gcloud iam service-accounts describe --project="${GOOGLE_PROJECT_ID}" "${SERVICE_ACCOUNT}@${GOOGLE_PROJECT_ID}.iam.gserviceaccount.com" > /dev/null 2>&1`,
		Message: map[command.State]string{
			command.PassState: "The service account already exists.",
			command.FailState: "No service account found. Next step is to create one.",
		},
	})
	return []*command.Model{
		checkServiceAccount,
		command.NewWriteModel(command.Opts{
			SkipIfSucceed: checkServiceAccount,
			Description: `This command will create a service account, and follow-up commands will attach ${PERMISSIONS} permissions.

https://cloud.google.com/iam/docs/creating-managing-service-accounts`,
			Command: `gcloud iam service-accounts create "${SERVICE_ACCOUNT}" \
	--project="${GOOGLE_PROJECT_ID}" \
	--display-name="${SERVICE_ACCOUNT}" \
	--description="Runs Flightcrew's Control Tower VM."`,
		}),
	}
}

func getBindIAMPolicyCommands(args map[string]string) []*command.Model {
	newCheckPolicy := func(replacer *strings.Replacer) *command.Model {
		cmd := command.NewReadModel(command.Opts{
			Description: "Check if IAM policy binding already exists or needs to be created.",
			Command:     `gcloud projects get-iam-policy ${GOOGLE_PROJECT_ID} --filter="bindings.role=${PROJECT_OR_ORG_SLASH}/roles/${ROLE}"  --flatten=bindings --format="table(bindings.members,bindings.role)" | grep --quiet "'serviceAccount:${SERVICE_ACCOUNT}@${GOOGLE_PROJECT_ID}.iam.gserviceaccount.com'"`,
			Message: map[command.State]string{
				command.PassState: "Binding already exists.",
				command.FailState: "Binding doesn't exist. Next step is to add the binding.",
			},
		})
		cmd.Replace(replacer)
		return cmd
	}

	newAttachPolicy := func(replacer *strings.Replacer, skipIfSucceed *command.Model) *command.Model {
		cmd := command.NewWriteModel(command.Opts{
			SkipIfSucceed: skipIfSucceed,
			Description: `This command binds the ${PERMISSIONS} IAM role to the created service account, which grants the associated permissions to Flightcrew's service account for running the to-be-created VM.

https://cloud.google.com/iam/docs/granting-changing-revoking-access`,
			Command: `gcloud projects add-iam-policy-binding "${GOOGLE_PROJECT_ID}" \
	--member=serviceAccount:"${SERVICE_ACCOUNT}@${GOOGLE_PROJECT_ID}.iam.gserviceaccount.com" \
	--role="${PROJECT_OR_ORG_SLASH}/roles/${ROLE}" \
	--condition=None`,
		})
		cmd.Replace(replacer)
		return cmd
	}

	commands := make([]*command.Model, 0)

	readReplacer := strings.NewReplacer(
		"${ROLE}", gconst.KeyIAMRoleRead,
		"${PERMISSIONS}", constants.Read,
	)
	readCheck := newCheckPolicy(readReplacer)
	commands = append(commands, readCheck, newAttachPolicy(readReplacer, readCheck))

	if args[gconst.KeyPermissions] == constants.Write {
		writeReplacer := strings.NewReplacer(
			"${ROLE}", gconst.KeyIAMRoleWrite,
			"${PERMISSIONS}", constants.Write,
		)
		writeCheck := newCheckPolicy(writeReplacer)
		commands = append(commands, writeCheck, newAttachPolicy(writeReplacer, writeCheck))
	}

	return commands
}

func getVMCommands(args map[string]string) []*command.Model {
	checkVMExists := command.NewReadModel(command.Opts{
		Description: "Check if a Flightcrew VM already exists or needs to be created.",
		Command: `gcloud compute instances list --format="csv(NAME,EXTERNAL_IP,STATUS)" \${PROJECT_OR_ORG_FLAG}
	--zones=${ZONE} | awk -F "," "/${VIRTUAL_MACHINE}/ {print f(2), f(3)} function f(n){return (\$n==\"\" ? \"null\" : \$n)}" | [ $(wc -c) -gt "0" ]`,
		Message: map[command.State]string{
			command.PassState: "This Flightcrew VM already exists. Nothing to install.",
			command.FailState: "No existing VM found. Next step is to create it.",
		},
	})

	return []*command.Model{
		checkVMExists,
		command.NewWriteModel(command.Opts{
			SkipIfSucceed: checkVMExists,
			Description:   "Create a VM instance attached to Flightcrew's service account, and run the Control Tower image.",
			Command: `gcloud compute instances create-with-container ${VIRTUAL_MACHINE} \
	--project=${GOOGLE_PROJECT_ID} \
	--container-command="/ko-app/tower" \
	--container-image="${IMAGE_PATH}:${TOWER_VERSION}" \
	--container-arg="--debug=true" \
	--container-env="FC_API_KEY=${API_TOKEN}" \
	--container-env="CLOUD_PLATFORM=${PLATFORM}" \${TRAFFIC_ROUTER}${GAE_MAX_VERSION_COUNT}${GAE_MAX_VERSION_AGE}
	--container-env="FC_PACKAGE_VERSION=${TOWER_VERSION}" \
	--container-env="METRIC_PROVIDERS=stackdriver" \
	--container-env="FC_RPC_CONNECT_HOST=${RPC_HOST}" \
	--container-env="FC_RPC_CONNECT_PORT=443" \
	--container-env="FC_TOWER_PORT=8080" \
	--labels="component=flightcrew" \
	--machine-type="e2-micro" \
	--scopes="cloud-platform" \
	--service-account="${SERVICE_ACCOUNT}@${GOOGLE_PROJECT_ID}.iam.gserviceaccount.com" \
	--tags="http-server" \
	--zone="${ZONE}"`,
		}),
		command.NewWriteModel(command.Opts{
			SkipIfSucceed: checkVMExists,
			Command: `gcloud compute instances add-metadata ${VIRTUAL_MACHINE} \
	--project=${GOOGLE_PROJECT_ID} \
	--zone=${ZONE}  \
	--metadata=google-logging-enabled=false`,
			Description: `Disable the VM's builtin logger because it has a memory leak and will cause the VM to crash after 1-2 weeks.

https://serverfault.com/questions/980569/disable-fluentd-on-on-container-optimized-os-gce`,
		}),
	}
}
