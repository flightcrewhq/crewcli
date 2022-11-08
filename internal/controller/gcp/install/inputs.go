package gcpinstall

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"flightcrew.io/cli/internal/constants"
	"flightcrew.io/cli/internal/controller"
	"flightcrew.io/cli/internal/controller/gcp"
	gconst "flightcrew.io/cli/internal/controller/gcp/constants"
	"flightcrew.io/cli/internal/debug"
	"flightcrew.io/cli/internal/timeconv"
	"flightcrew.io/cli/internal/view/wrapinput"
)

var (
	filenameReplacer = strings.NewReplacer(
		".", "_",
		"/", "_",
		":", "_",
		" ", "_",
	)

	initialInputKeys = []string{
		gconst.KeyProject,
		gconst.KeyVirtualMachine,
		gconst.KeyAPIToken,
		gconst.KeyPlatform,
		gconst.KeyPermissions,
		gconst.KeyZone,
		gconst.KeyTowerVersion,
		gconst.KeyIAMServiceAccount,
	}

	writeAppEngineInputKeys = []string{
		gconst.KeyProject,
		gconst.KeyVirtualMachine,
		gconst.KeyAPIToken,
		gconst.KeyPlatform,
		gconst.KeyPermissions,
		gconst.KeyGAEMaxVersionAge,
		gconst.KeyGAEMaxVersionCount,
		gconst.KeyZone,
		gconst.KeyTowerVersion,
		gconst.KeyIAMServiceAccount,
	}
)

type InputsController struct {
	tempDir   string
	inputs    map[string]*wrapinput.Model
	args      map[string]string
	inputKeys []string
}

func NewInputsController(params Params) *InputsController {
	ctl := &InputsController{
		inputKeys: initialInputKeys,
		inputs:    make(map[string]*wrapinput.Model),
		args:      params.args,
		tempDir:   params.tempDir,
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

	baseURL := gcp.GetHostBaseURL("", "")
	ctl.args[gconst.KeyAppURL] = constants.GetAppHostName(baseURL)
	ctl.args[gconst.KeyRPCHost] = constants.GetAPIHostName(baseURL)
	ctl.args[gconst.KeyTrafficRouter] = ""
	ctl.args[gconst.KeyGAEMaxVersionAge] = ""
	ctl.args[gconst.KeyGAEMaxVersionCount] = ""
	ctl.args[gconst.KeyImagePath] = gcp.ImagePath
	ctl.args[gconst.KeyProjectOrOrgFlag] = ""
	ctl.args[gconst.KeyProjectOrOrgSlash] = ""

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

		case gconst.KeyAPIToken:
			input = wrapinput.NewFreeForm()
			input.Freeform.Placeholder = "api-token"
			input.Freeform.CharLimit = 0
			input.Title = "API Token"
			input.Required = true
			input.HelpText = "API token is the value provided by Flightcrew to identify your organization."
			maybeSetValue(gconst.KeyAPIToken)

		case gconst.KeyIAMServiceAccount:
			input = wrapinput.NewFreeForm()
			input.Title = "Service Account"
			input.Freeform.CharLimit = 64
			input.Default = "flightcrew-runner"
			input.SetValue("flightcrew-runner")
			input.HelpText = "Service Account is the name of the (to be created) IAM service account to run the Flightcrew Tower."

		case gconst.KeyPlatform:
			input = wrapinput.NewRadio([]string{
				constants.GoogleAppEngineStdDisplay,
				constants.GoogleComputeEngineDisplay})
			input.Title = "Platform"
			input.HelpText = "Platform is which Google Cloud Provider resources Flightcrew will read in."
			maybeSetValue(gconst.KeyPlatform)

		case gconst.KeyPermissions:
			input = wrapinput.NewRadio([]string{
				constants.Read,
				constants.Write})
			input.Title = "Permissions"
			input.HelpText = "Permissions is whether Flightcrew will only read in your resources, or if Flightcrew can modify (if you ask us to) your resources."
			maybeSetValue(gconst.KeyPermissions)

		case gconst.KeyGAEMaxVersionAge:
			input = wrapinput.NewFreeForm()
			input.Title = "Max Version Age"
			input.Freeform.Placeholder = "168h"
			input.HelpText = "The Tower (App Engine + Write) will prune old versions that are receiving no traffic when they become older than this age (in h,m,s).\nLeave blank to disable."

		case gconst.KeyGAEMaxVersionCount:
			input = wrapinput.NewFreeForm()
			input.Title = "Max Version Count"
			input.Freeform.Placeholder = "30"
			input.HelpText = "The Tower (App Engine + Write) will prune old versions that are receiving no traffic when the number of old versions exceeds this count.\nLeave blank to disable."

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
		case gconst.KeyProject:
			projectID := input.Value()
			orgID, err := gcp.GetOrganizationID(projectID)
			if err != nil {
				input.SetInfo("no organization found")
				ctl.args[gconst.KeyProjectOrOrgFlag] = fmtFlagForReplace("project", projectID)
				ctl.args[gconst.KeyProjectOrOrgSlash] = fmt.Sprintf(`projects/%s`, projectID)
			} else {
				input.SetInfo("found organization ID '" + orgID + "'")
				ctl.args[gconst.KeyProjectOrOrgFlag] = fmtFlagForReplace("organization", orgID)
				ctl.args[gconst.KeyProjectOrOrgSlash] = fmt.Sprintf(`organizations/%s`, orgID)
			}

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
			baseURL := gcp.GetHostBaseURL(projectInput.Value(), input.Value())
			ctl.args[gconst.KeyRPCHost] = constants.GetAPIHostName(baseURL)
			ctl.args[gconst.KeyAppURL] = constants.GetAppHostName(baseURL)

		case gconst.KeyPlatform:
			displayName := input.Value()
			platform, ok := constants.DisplayToPlatform[displayName]
			if !ok {
				setError(errors.New("invalid platform"))
				break
			}

			input.SetConverted(platform)

		case gconst.KeyPermissions:
			platformInput := ctl.inputs[gconst.KeyPlatform]
			platform, ok := constants.DisplayToPlatform[platformInput.Value()]
			if !ok {
				if _, ok := constants.PlatformPermissions[platformInput.Value()]; !ok {
					setError(errors.New("need to set platform first"))
					break
				} else {
					platform = platformInput.Value()
				}
			}

			perms, ok := constants.PlatformPermissions[platform]
			if !ok {
				setError(errors.New("platform has no permissions"))
				break
			}

			permission := input.Value()
			if _, ok := perms[permission]; !ok {
				setError(fmt.Errorf("%s permissions are not supported for platform '%s'", permission, platformInput.Value()))
				break
			}

			var err error
			readSettings := perms[constants.Read]
			ctl.args[gconst.KeyIAMRoleRead] = readSettings.Role
			ctl.args[gconst.KeyIAMFileRead], err = ctl.createFileWithContents(platform, constants.Read, readSettings.Content, "yaml")
			if setError(err) {
				break
			}

			if permission == constants.Write {
				writeSettings := perms[constants.Write]
				ctl.args[gconst.KeyTrafficRouter] = fmtContainerEnvForReplace("TRAFFIC_ROUTER", platform)
				ctl.args[gconst.KeyIAMRoleWrite] = writeSettings.Role
				ctl.args[gconst.KeyIAMFileWrite], err = ctl.createFileWithContents(platform, constants.Write, writeSettings.Content, "yaml")
				if setError(err) {
					break
				}
			} else {
				ctl.args[gconst.KeyTrafficRouter] = ""
				ctl.args[gconst.KeyIAMRoleWrite] = ""
				ctl.args[gconst.KeyIAMFileWrite] = ""
			}

		case gconst.KeyGAEMaxVersionCount:
			value := input.Value()
			if len(value) == 0 {
				break
			}

			numMaxVersions, err := strconv.Atoi(value)
			if err != nil {
				setError(errors.New("must be a positive integer"))
				break
			}

			if numMaxVersions < 1 {
				setError(errors.New("must be positive"))
			}

			input.SetInfo("")
			input.SetConverted(fmtContainerEnvForReplace("APPENGINE_MAX_VERSION_COUNT", fmt.Sprintf("%d", numMaxVersions)))

		case gconst.KeyGAEMaxVersionAge:
			value := input.Value()
			if len(value) == 0 {
				break
			}

			dur, err := timeconv.ParseDuration(value)
			if err != nil {
				setError(errors.New("must be a duration (mo, w, d, h, m, s) (e.g. 1mo, 2w, 5d3h)"))
				break
			}

			converted, err := convertDuration(dur)
			if setError(err) {
				break
			}

			input.SetInfo(converted)
			input.SetConverted(fmtContainerEnvForReplace("APPENGINE_MAX_VERSION_AGE", converted))

		}
	}

	return !hasErrors
}

var convertDuration = timeconv.GetDurationFormatter([]string{"h", "m", "s"})

func (ctl InputsController) GetRunController() controller.Run {
	for _, k := range ctl.inputKeys {
		ctl.args[k] = ctl.inputs[k].Value()
	}

	return NewRunController(ctl.args)
}

func (ctl InputsController) GetName() string {
	return "Google Cloud Platform Installation"
}

func (ctl *InputsController) GetInputs() []*wrapinput.Model {
	platformInput := ctl.inputs[gconst.KeyPlatform]
	permissionsInput := ctl.inputs[gconst.KeyPermissions]
	if platformInput.Radio.Value() == constants.GoogleAppEngineStdDisplay &&
		permissionsInput.Radio.Value() == constants.Write {
		ctl.inputKeys = writeAppEngineInputKeys
	} else {
		ctl.inputKeys = initialInputKeys
	}

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

func fmtContainerEnvForReplace(env string, value string) string {
	return fmtForReplace(fmt.Sprintf(`--container-env="%s=%s"`, env, value))
}

func fmtFlagForReplace(flag string, value string) string {
	return fmtForReplace(fmt.Sprintf(`--%s=%s`, flag, value))
}

func fmtForReplace(value string) string {
	return fmt.Sprintf(`
	%s \`, value)
}

func (ctl *InputsController) createFileWithContents(platform string, permissions string, contents string, extension string) (string, error) {
	fn := filepath.Join(ctl.tempDir, fmt.Sprintf("%s.%s", filenameReplacer.Replace(fmt.Sprintf("%s_%s", permissions, platform)), extension))
	if _, err := os.Stat(fn); err == nil {
		return fn, nil
	}

	f, err := os.Create(fn)
	if err != nil {
		return "", err
	}
	defer f.Close()

	if _, err := f.WriteString(contents); err != nil {
		return "", err
	}

	return f.Name(), nil
}
