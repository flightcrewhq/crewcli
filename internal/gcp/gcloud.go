package gcp

import (
	"bytes"
	"os/exec"
)

func HasGcloudInPath() bool {
	cmd := exec.Command("bash", "-c", `which gcloud`)
	var b bytes.Buffer
	cmd.Stdout = &b
	cmd.Stderr = &b
	return cmd.Run() == nil
}
