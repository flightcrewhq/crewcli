package gcp

import (
	"bytes"
	"fmt"
	"os/exec"
)

func HasGcloudInPath() bool {
	cmd := exec.Command("bash", "-c", `which gcloud`)
	var b bytes.Buffer
	cmd.Stdout = &b
	cmd.Stderr = &b
	hasGcloudInPath := cmd.Run() == nil

	if !hasGcloudInPath {
		fmt.Printf(`The "gcloud" CLI tool is a pre-requisite to run this script.

		If you haven't yet, please install the tool: https://cloud.google.com/sdk/docs/install

		If you already have, please add it to your path:
		  export PATH=<where it is>:$PATH

		`)
	}
	return hasGcloudInPath
}
