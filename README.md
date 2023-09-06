# `crewcli`

![Go](https://github.com/flightcrewhq/cli/actions/workflows/go.yaml/badge.svg)

This is the command line interface for installing Flightcrew.

## Get Started

1. Specify where you'd like `crewcli` to exist.
2. Run the bootstrap script to download the latest release and verify that the checksums match. \
   Depending on where you want to download to, you may need `sudo` to write the binary there.

   To inspect the contents, go to <https://github.com/flightcrewhq/crewcli/blob/main/bootstrap.sh/>

```sh
# with sudo
export OUTDIR=/usr/local/bin
curl --location https://raw.githubusercontent.com/flightcrewhq/crewcli/main/bootstrap.sh | sudo OUTDIR=${OUTDIR} bash
```

```sh
# no sudo
export OUTDIR=./bin
curl --location https://raw.githubusercontent.com/flightcrewhq/crewcli/main/bootstrap.sh | bash
```

## Usage

`crewcli` currently supports Google Cloud Platform.

To use, run `crewcli gcp install` or `crewcli gcp upgrade` to get started. This will start up an interactive terminal to get you set up.

Commands that modify your GCP state will NOT be run until user permission is given. However, some commands to get additional details to make the process smoother may be run. Nothing is being logged.

For more details, the commands that are run can be found below:

* `gcp install`: <https://github.com/flightcrewhq/crewcli/blob/main/internal/controller/gcp/install/run.go/>
* `gcp upgrade`: <https://github.com/flightcrewhq/crewcli/blob/main/internal/controller/gcp/upgrade/run.go/>

For Kubernetes, please use our Helm chart. Reach out to [hello@flightcrew.io](mailto:hello@flightcrew.io) for access.

## Contact

For help, reach out to [support@flightcrew.io](mailto:support@flightcrew.io).

Sign up for our newsletter at <https://flightcrew.io/>!
