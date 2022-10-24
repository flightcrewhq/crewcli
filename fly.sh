#!/bin/bash

# Script for users to provision a GCE VM and run our Control Tower.
# Docs: https://cloud.google.com/sdk/gcloud/reference/compute/instances/create-with-container
# Look for "Equivalent Command Line" at the bottom: https://console.cloud.google.com/compute/instancesAdd

set -e

# Color Codes.
CODE_WHITE="37"
CODE_GREEN="32"
CODE_YELLOW="33"
CODE_RED="31"

CODE_BOLD="1"
CODE_NORMAL="0"

COLOR_RESET="\e[0m"
COLOR_BOLDWHITE="\e[${CODE_BOLD};${CODE_WHITE}m"
COLOR_BOLDGREEN="\e[${CODE_BOLD};${CODE_GREEN}m"
COLOR_BOLDRED="\e[${CODE_BOLD};${CODE_RED}m"
COLOR_YELLOW="\e[${CODE_NORMAL};${CODE_YELLOW}m"

ERR_INVALID_ARGUMENT=2
ERR_GCLOUD_FAILED=3

# Default flag settings
PERMISSIONS=readonly
IAM_ROLE=flightcrew.gae.read.only
IAM_FILE=gcp/gae/iam_readonly.yaml
FC_SERVICE_ACCOUNT=flightcrew-runner

VIRTUAL_MACHINE=flightcrew-control-tower
ZONE=us-central1-c
TOWER_IMAGE_VERSION=stable
ENV_FILE=container.env

help() {
	SCRIPT_NAME=`basename "$0"`
	echo "This script sets up Flightcrew as a VM in your Google Cloud project to
access metrics and configs for display in the Flightcrew UI:
https://app.flightcrew.io

Usage: sh ${SCRIPT_NAME} --project project_name --token your_api_key --permissions readwrite

common flags:
  --project          required always      Google Cloud project name.
  --token            required on create   The Flightcrew API token.
  --zone                                  Which zone to run the VM in.
                                            Default: us-central1-c
  --vm_name                               Name for the virtual machine.
                                            Default: flightcrew-control-tower
  --image_version                         The image tag that runs in the container.
                                            Default: stable
  --permissions                           Permissions to grant the VM \"readonly\" or \"readwrite\".
                                            Default: readonly
"
}

# Parse arguments.
while true; do
	case $1 in
		-h|--help|help)
			help
			exit 0
			;;
		-p|--project)
			GOOGLE_PROJECT_ID=$2
			;;
		--token)
			FLIGHTCREW_TOKEN=$2
			;;
		--image_version)
			TOWER_IMAGE_VERSION=$2
			;;
		--vm_name)
			VIRTUAL_MACHINE=$2
			;;
		--zone)
			ZONE=$2
			;;
		--permissions)
			PERMISSIONS=$2
			;;
		*)
			if [[ ! -z "$1" ]]; then
				printf "${COLOR_BOLDRED}[Error]${COLOR_RESET} Unknown option: $1\n\n"
				help
				exit $ERR_INVALID_ARGUMENT
			fi
			break
			;;
	esac

  if [[ -z "$2" ]]; then
    printf "${COLOR_BOLDRED}[Error]${COLOR_RESET} $1 can't be empty.\n\n"
  	help
  	exit $ERR_INVALID_ARGUMENT
  else
    shift 2
  fi
done

# Prints the command, asks for permission, and then executes the command.
#
# Usage: print_ask_exec "command" "description"
#   $1 = the command to run
#   $2 = the description for the command that will be printed to STDOUT
print_ask_exec() {
  COMMAND=$1
	DESCRIPTION=$2
	while true; do
	  echo "$DESCRIPTION"
	  printf "Run \`${COLOR_BOLDWHITE}${COMMAND}${COLOR_RESET}\`? ${COLOR_YELLOW}[yes/no]${COLOR_RESET}: "
		read LINE
		case $LINE in
			yes)
				break
				;;
			no)
				echo "User cancelled execution. Exiting..."
				exit 0
				;;
			*)
				printf "${COLOR_BOLDRED}[Error]${COLOR_RESET} Invalid input. Trying again.\n"
				;;
		esac
	done

	eval "$COMMAND"

	echo "Done"
	echo
}

# Argument validation.
if [[ -z "$GOOGLE_PROJECT_ID" ]]; then
	printf "${COLOR_BOLDRED}[Error]${COLOR_RESET} Google Cloud project name cannot be empty.\n\n"
	help
	exit $ERR_INVALID_ARGUMENT
fi

if [[ "${PERMISSIONS}" == "readonly" ]]; then
  echo "VM will have read permissions only."

elif [[ "${PERMISSIONS}" == "readwrite" ]]; then
  echo "VM will have read and write permissions."

  IAM_ROLE=flightcrew.gae.write
  IAM_FILE=gcp/gae/iam_readonly.yaml

else
	printf "${COLOR_BOLDRED}[Error]${COLOR_RESET} --permissions value is unrecognized.\n\n"
	help
	exit $ERR_INVALID_ARGUMENT
fi

SERVICE_ACCT_EMAIL="${FC_SERVICE_ACCOUNT}"@"${GOOGLE_PROJECT_ID}".iam.gserviceaccount.com
FULL_IMAGE_PATH=us-west1-docker.pkg.dev/flightcrew-artifacts/client/tower:"${TOWER_IMAGE_VERSION}"

echo "Checking if Flightcrew's service account (${FC_SERVICE_ACCOUNT}) exists."
if gcloud iam service-accounts describe --project="${GOOGLE_PROJECT_ID}" "${SERVICE_ACCT_EMAIL}" >/dev/null 2>&1 ; then
  echo "Service account ${FC_SERVICE_ACCOUNT} exists."

else
  echo "Service account ${FC_SERVICE_ACCOUNT} wasn't found."

  print_ask_exec \
    "gcloud iam service-accounts create ${FC_SERVICE_ACCOUNT} \\
    --project=${GOOGLE_PROJECT_ID} \\
    --display-name=${FC_SERVICE_ACCOUNT} \\
    --description=\"Runs Flightcrew's Control Tower VM.\"" \
    "This command will create a service account, and follow-up commands will attach read and/or write permissions.
    See https://cloud.google.com/iam/docs/creating-managing-service-accounts"
fi

echo "Checking if Flightcrew's IAM role (${IAM_ROLE}) exists."
if gcloud iam roles describe --project="${GOOGLE_PROJECT_ID}" "${IAM_ROLE}" >/dev/null 2>&1 ; then
  echo "${IAM_ROLE} exists."
else
  echo "IAM role ${IAM_ROLE} wasn't found."

  print_ask_exec \
    "gcloud iam roles create ${IAM_ROLE} \\
    --project=${GOOGLE_PROJECT_ID} \\
    --file=${IAM_FILE}" \
    "This command creates an IAM role from ${IAM_FILE} for the Flightcrew VM to access configs and monitoring data.
    See https://cloud.google.com/iam/docs/understanding-custom-roles"
  
  print_ask_exec \
    "gcloud projects add-iam-policy-binding \"${GOOGLE_PROJECT_ID}\" \\
    --member=serviceAccount:\"${SERVICE_ACCT_EMAIL}\" \\
    --role=projects/${GOOGLE_PROJECT_ID}/roles/${IAM_ROLE} \\
    --condition=None \\
    > /dev/null" \
    "This command attaches the IAM role to Flightcrew's service account, which will give the IAM permissions to a new VM.
    See https://cloud.google.com/iam/docs/granting-changing-revoking-access"
fi

echo "Checking if Flightcrew's VM (${VIRTUAL_MACHINE}) exists and is running."
IP_STATUS=( $(gcloud compute instances list --format="csv(NAME,EXTERNAL_IP,STATUS)" --project=${GOOGLE_PROJECT_ID} --zones=${ZONE} | awk -F "," "/${VIRTUAL_MACHINE}/ {print f(2), f(3)} function f(n){return (\$n==\"\" ? \"null\" : \$n)}") )

if [[ -z "${IP_STATUS[1]}" ]]; then
  echo "VM ${VIRTUAL_MACHINE} wasn't found."

  if [[ -z "$FLIGHTCREW_TOKEN" ]]; then
    printf "${COLOR_BOLDRED}[Error]${COLOR_RESET} The Flightcrew token cannot be empty when creating the VM, set --token.\n\n"
    help
    exit $ERR_INVALID_ARGUMENT
  fi

  print_ask_exec \
    "gcloud compute instances create-with-container ${VIRTUAL_MACHINE} \\
    --project=${GOOGLE_PROJECT_ID} \\
    --container-command=\"/ko-app/tower\" \\
    --container-image=${FULL_IMAGE_PATH} \\
    --container-arg=\"--debug=true\" \\
    --container-env-file=${ENV_FILE} \\
    --container-env=FC_API_KEY=${FLIGHTCREW_TOKEN} \\
    --container-env=FC_PACKAGE_VERSION=${TOWER_IMAGE_VERSION} \\
    --machine-type=e2-micro \\
    --scopes=cloud-platform \\
    --service-account=\"${SERVICE_ACCT_EMAIL}\" \\
    --tags=http-server \\
    --zone=${ZONE}" \
    "Create a VM instance attached to Flightcrew's service account, and run the Control Tower image."

  print_ask_exec \
    "gcloud compute instances add-metadata ${VIRTUAL_MACHINE} \\
      --project=${GOOGLE_PROJECT_ID} \\
      --zone=${ZONE}  \\
      --metadata=google-logging-enabled=false" \
      "Disable the VM's builtin logger because it has a memory leak and will cause the VM to crash after 1-2 weeks."

else
  echo "VM \"${VIRTUAL_MACHINE}\" is ${IP_STATUS[1]}."

  if nc -w 1 -z "${IP_STATUS[0]}" 22 ; then
    print_ask_exec \
      "gcloud compute ssh ${VIRTUAL_MACHINE} \\
      --project ${GOOGLE_PROJECT_ID} \\
      --zone ${ZONE} \\
      --command 'docker system prune -f -a'" \
      "SSH to the VM is available - This command prunes old images to save space before downloading a new one."
  fi

  # update-container keeps the existing args / envs from when the VM was created,
  # but they can be overwritten here.
  print_ask_exec \
    "gcloud compute instances update-container ${VIRTUAL_MACHINE} \\
    --project=${GOOGLE_PROJECT_ID} \\
    --zone=${ZONE} \\
    --container-env-file=\"${ENV_FILE}\" \\
    --container-image=\"${FULL_IMAGE_PATH}\" \\
    --container-env=\"FC_PACKAGE_VERSION=${TOWER_IMAGE_VERSION}\"" \
    "This command updates the VM to the newest stable Control Tower image."
fi

printf "${COLOR_BOLDGREEN}[SUCCESS]${COLOR_RESET} Script finished!

GCE takes a couple minutes to start up the VM.

See your VM in the console:
https://console.cloud.google.com/compute/instancesDetail/zones/${ZONE}/instances/${VIRTUAL_MACHINE}?project=${GOOGLE_PROJECT_ID}

Alternatively, see your new VM in action:
# SSH into the created VM.
${COLOR_BOLDWHITE}\$${COLOR_RESET} gcloud compute ssh ${VIRTUAL_MACHINE} --project ${GOOGLE_PROJECT_ID} --zone ${ZONE}
# Follow the new container's logs.
${COLOR_BOLDWHITE}\$${COLOR_RESET} docker logs --follow \$(docker ps -f name=tower --format=\"{{.ID}}\")
"
