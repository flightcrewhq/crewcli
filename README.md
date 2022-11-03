# `crewcli`

![Go](https://github.com/flightcrewhq/cli/actions/workflows/go.yaml/badge.svg)

This is the command line interface for installing Flightcrew.

## Get Started

1. Specify where you'd like `crewcli` to exist.
2. Run the bootstrap script to download the latest release and verify that the checksums match.

   To inspect the contents, go to <https://github.com/flightcrewhq/crewcli/blob/main/bootstrap.sh/>

```sh
# Where you'd like crewcli to be installed.
export OUTDIR=/usr/local/bin
# Please look at the file before you run it.
curl --location https://raw.githubusercontent.com/flightcrewhq/crewcli/main/bootstrap.sh | bash
```

## Usage

`crewcli` currently supports Google Cloud Platform.

To use, run `crewcli gcp install` to get started. This will start up an interactive terminal to get you set up.

For Kubernetes, please use our Helm chart. Reach out to sam@flightcrew.io for access.

## Contact

For help, reach out to support@flightcrew.io.

Sign up to our newsletter at <https://flightcrew.io/>!
