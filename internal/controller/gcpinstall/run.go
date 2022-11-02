package gcpinstall

import (
	"strings"

	"flightcrew.io/cli/internal/controller"
	"flightcrew.io/cli/internal/view/command"
)

type RunController struct {
	args     map[string]string
	commands []*command.Model
}

func NewRunController(args map[string]string) *RunController {
	replaceArgs := make([]string, 0, 2*len(args))
	for key, arg := range args {
		replaceArgs = append(replaceArgs, key, arg)
	}

	checkServiceAccount := command.NewReadModel(command.Opts{
		Description: "Check if a Flightcrew service account already exists or needs to be created.",
		Command:     `gcloud iam service-accounts describe --project="${GOOGLE_PROJECT_ID}" "${SERVICE_ACCOUNT}@${GOOGLE_PROJECT_ID}.iam.gserviceaccount.com" > /dev/null 2>&1`,
		Message: map[command.State]string{
			command.PassState: "The service account already exists.",
			command.FailState: "No service account found. Next step is to create one.",
		},
	})
	checkIAMRole := command.NewReadModel(command.Opts{
		Description: "Check if a Flightcrew IAM Role already exists or needs to be created.",
		Command:     `gcloud iam roles describe --project="${GOOGLE_PROJECT_ID}" "${IAM_ROLE}" >/dev/null 2>&1`,
		Message: map[command.State]string{
			command.PassState: "This Flightcrew IAM role already exists.",
			command.FailState: "No IAM role found. Next step is to create one.",
		},
	})
	checkVMExists := command.NewReadModel(command.Opts{
		Description: "Check if a Flightcrew VM already exists or needs to be created.",
		Command:     `gcloud compute instances list --format="csv(NAME,EXTERNAL_IP,STATUS)" --project=${GOOGLE_PROJECT_ID} --zones=${ZONE} | awk -F "," "/${VIRTUAL_MACHINE}/ {print f(2), f(3)} function f(n){return (\$n==\"\" ? \"null\" : \$n)}" | [ $(wc -c) -gt "0" ]`,
		Message: map[command.State]string{
			command.PassState: "This Flightcrew VM already exists. Nothing to install.",
			command.FailState: "No existing VM found. Next step is to create it.",
		},
	})

	commands := []*command.Model{
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
		checkIAMRole,
		command.NewWriteModel(command.Opts{
			SkipIfSucceed: checkIAMRole,
			Description:   "This command creates an IAM role from `${IAM_FILE}` for the Flightcrew VM to access configs and monitoring data.\n\nhttps://cloud.google.com/iam/docs/understanding-custom-roles",
			Command: `gcloud iam roles create ${IAM_ROLE} \
	--project=${GOOGLE_PROJECT_ID} \
	--file=${IAM_FILE}`,
		}),
		command.NewWriteModel(command.Opts{
			SkipIfSucceed: checkIAMRole,
			Description: `This command attaches the IAM role to Flightcrew's service account, which will give the IAM permissions to a new VM.

https://cloud.google.com/iam/docs/granting-changing-revoking-access`,
			Command: `gcloud projects add-iam-policy-binding "${GOOGLE_PROJECT_ID}" \
	--member=serviceAccount:"${SERVICE_ACCOUNT}@${GOOGLE_PROJECT_ID}.iam.gserviceaccount.com" \
	--role="projects/${GOOGLE_PROJECT_ID}/roles/${IAM_ROLE}" \
	--condition=None`,
		}),
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
	--container-env=CLOUD_PLATFORM=${PLATFORM} \${TRAFFIC_ROUTER}${GAE_MAX_VERSION_COUNT}${GAE_MAX_VERSION_AGE}
	--container-env=FC_PACKAGE_VERSION=${TOWER_VERSION} \
	--container-env=METRIC_PROVIDERS=stackdriver \
	--container-env=FC_RPC_CONNECT_HOST=${RPC_HOST} \
	--container-env=FC_RPC_CONNECT_PORT=443 \
	--container-env=FC_TOWER_PORT=8080 \
	--label=component=flightcrew \
	--machine-type=e2-micro \
	--scopes=cloud-platform \
	--service-account="${SERVICE_ACCOUNT}@${GOOGLE_PROJECT_ID}.iam.gserviceaccount.com" \
	--tags=http-server \
	--zone=${ZONE}`,
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

	replacer := strings.NewReplacer(replaceArgs...)
	for _, cmd := range commands {
		cmd.Replace(replacer)
	}

	return &RunController{
		args:     args,
		commands: commands,
	}
}

func (ctl RunController) Commands() []*command.Model {
	return ctl.commands
}

func (ctl *RunController) GetEndController() controller.End {
	return NewEndController(ctl.commands)
}
