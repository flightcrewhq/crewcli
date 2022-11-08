package gcpupgrade

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"flightcrew.io/cli/internal/controller"
	"flightcrew.io/cli/internal/controller/gcp"
	gconst "flightcrew.io/cli/internal/controller/gcp/constants"
	"flightcrew.io/cli/internal/debug"
	"flightcrew.io/cli/internal/view/wrapinput"
)

var (
	initialInputKeys = []string{
		gconst.KeyProject,
		gconst.KeyVirtualMachine,
		gconst.KeyZone,
		gconst.KeyTowerVersion,
	}
)

type InputsController struct {
	inputs    map[string]*wrapinput.Model
	args      map[string]string
	inputKeys []string
}

func NewInputsController(params Params) *InputsController {
	ctl := &InputsController{
		inputKeys: initialInputKeys,
		inputs:    make(map[string]*wrapinput.Model),
		args:      params.args,
	}

	if !contains(ctl.args, gconst.KeyVirtualMachine) {
		ctl.args[gconst.KeyVirtualMachine] = "flightcrew-control-tower"
	}
	if !contains(ctl.args, gconst.KeyZone) {
		ctl.args[gconst.KeyZone] = "us-central1-c"
	}
	if !contains(ctl.args, gconst.KeyTowerVersion) {
		ctl.args[gconst.KeyTowerVersion] = "stable"
	}

	ctl.args[gconst.KeyImagePath] = gcp.ImagePath

	for _, key := range allKeys {
		var input wrapinput.Model
		maybeSetValue := func(key string) {
			if val, ok := ctl.args[key]; ok {
				input.SetValue(val)
			}
		}

		switch key {
		case gconst.KeyProject:
			input = wrapinput.NewFreeForm()
			input.Freeform.CharLimit = 0
			input.Freeform.Placeholder = "project-id-1234"
			if project, err := gcp.GetProjectFromEnvironment(); err == nil && len(project) > 0 {
				input.Freeform.Placeholder = project
			}
			input.Title = "Project ID"
			input.HelpText = "Project ID is the unique string identifier for your Google Cloud Platform project."
			input.Required = true
			maybeSetValue(gconst.KeyProject)

		case gconst.KeyVirtualMachine:
			input = wrapinput.NewFreeForm()
			input.Freeform.Placeholder = "flightcrew-control-tower"
			input.Freeform.CharLimit = 64
			input.Title = "VM Name"
			input.Default = "flightcrew-control-tower"
			input.Required = true
			input.HelpText = "VM Name is what the (to be installed) Flightcrew virtual machine instance will be named."
			maybeSetValue(gconst.KeyVirtualMachine)

		case gconst.KeyZone:
			input = wrapinput.NewFreeForm()
			input.Freeform.Placeholder = "us-central1-c"
			input.Freeform.CharLimit = 32
			input.Title = "Zone"
			input.Default = "us-central1-c"
			input.HelpText = "Zone is the Google zone where the (to be installed) Flightcrew virtual machine instance will be located."
			maybeSetValue(gconst.KeyZone)

		case gconst.KeyTowerVersion:
			input = wrapinput.NewFreeForm()
			input.Freeform.Placeholder = "stable"
			input.Freeform.CharLimit = 32
			input.Title = "Tower Version"
			input.HelpText = "Tower Version is the version of the Tower image that will be installed. (recommended: `stable`)"
			maybeSetValue(gconst.KeyTowerVersion)

		}

		input.Blur()
		ctl.inputs[key] = &input
	}
	return ctl
}

func (ctl InputsController) GetAllInputs() []*wrapinput.Model {
	res := make([]*wrapinput.Model, 0, len(ctl.inputs))
	for _, v := range ctl.inputs {
		res = append(res, v)
	}
	return res
}

func (ctl *InputsController) Reset(inputs []*wrapinput.Model) {
	for k := range ctl.inputs {
		ctl.inputs[k].ResetValidation()
	}
}

func (ctl *InputsController) Validate(inputs []*wrapinput.Model) bool {
	hasErrors := false
	for k, input := range ctl.inputs {
		setError := func(err error) bool {
			if err != nil {
				input.SetError(err)
				debug.Output(err.Error())
				hasErrors = true
				return true
			}
			return false
		}

		if input.Required && len(input.Value()) == 0 {
			setError(errors.New("required"))
			continue
		}

		switch k {
		case gconst.KeyTowerVersion:
			version, err := gcp.GetTowerImageVersion(input.Value())
			if setError(err) {
				debug.Output("convert tower version got error: %v", err)
				break
			}

			input.SetConverted(version)
			debug.Output("convert tower version is %s", version)

		case gconst.KeyVirtualMachine:
			projectInput := ctl.inputs[gconst.KeyProject]
			var err error
			ctl.args[gconst.KeyVirtualMachineIP], err = ctl.getVirtualMachineIP(projectInput.Value(), ctl.inputs[gconst.KeyZone].Value(), input.Value())
			if setError(err) {
				break
			}

			if ipAddr := ctl.args[gconst.KeyVirtualMachineIP]; len(ipAddr) > 0 {
				input.SetInfo(fmt.Sprintf("found VM with IP %s", ipAddr))
			} else {
				input.SetInfo("found stopped VM")
			}

		}
	}

	return !hasErrors
}

func (ctl InputsController) GetRunController() controller.Run {
	for _, k := range ctl.inputKeys {
		ctl.args[k] = ctl.inputs[k].Value()
	}

	return NewRunController(ctl.args)
}

func (ctl InputsController) GetName() string {
	return "Google Cloud Platform Upgrade"
}

func (ctl *InputsController) GetInputs() []*wrapinput.Model {
	ctl.inputKeys = initialInputKeys

	inputs := make([]*wrapinput.Model, 0, len(ctl.inputKeys))
	for _, k := range ctl.inputKeys {
		inputs = append(inputs, ctl.inputs[k])
	}
	return inputs
}

func (ctl *InputsController) RecreateCommand() string {
	for _, key := range ctl.inputKeys {
		ctl.args[key] = ctl.inputs[key].Value()
	}
	return recreateCommand(ctl.args)
}

func contains(m map[string]string, key string) bool {
	_, ok := m[key]
	return ok
}

func (ctl *InputsController) getVirtualMachineIP(projectID string, zone string, vmName string) (string, error) {
	var notFoundErr = errors.New("no VM with this name and location")

	cmdStr := `gcloud compute instances list --format="csv(NAME,EXTERNAL_IP,STATUS)" --project=${GOOGLE_PROJECT_ID} --zones=${ZONE} --filter="name=${VIRTUAL_MACHINE}"`
	cmdStr = strings.Replace(cmdStr, "${GOOGLE_PROJECT_ID}", projectID, 1)
	cmdStr = strings.Replace(cmdStr, "${VIRTUAL_MACHINE}", vmName, 1)
	cmdStr = strings.Replace(cmdStr, "${ZONE}", zone, 1)
	debug.Output("running `%s`", cmdStr)
	cmd := exec.Command("bash", "-c", cmdStr)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", notFoundErr
	}
	debug.Output(stdout.String())

	r := csv.NewReader(&stdout)
	headers, err := r.Read()
	if err != nil {
		return "", notFoundErr
	}

	nameIdx := -1
	ipIdx := -1
	for i, header := range headers {
		switch header {
		case "name":
			nameIdx = i
		case "external_ip":
			ipIdx = i
		}

		if nameIdx >= 0 && ipIdx >= 0 {
			break
		}
	}
	if nameIdx < 0 || ipIdx < 0 {
		return "", notFoundErr
	}

	for {
		entry, err := r.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return "", notFoundErr
		}

		if entry[nameIdx] == vmName {
			return entry[ipIdx], nil
		}
	}

	return "", notFoundErr
}
