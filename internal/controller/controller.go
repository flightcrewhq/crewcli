package controller

import (
	"flightcrew.io/cli/internal/view/command"
	"flightcrew.io/cli/internal/view/wrapinput"
)

const (
	DefaultHelpText = "> Edit a particular entry to see help text here.\n> Otherwise, press enter to proceed."
)

// The logical flow is
// 1. Define the inputs (InputsModel + Inputs)
// 2. Run the commands with the variables (RunModel + Run)
// 3. Show a summary and give additional information (EndModel + End)

// Inputs is a variable number of inputs that can be changed. This interface allows
// pre-defined defaults that can be modified according to cloud providers' specific
// requirements.
type Inputs interface {
	GetName() string

	// GetInputs will be called every time there is an update, so the implementation has
	// the opportunity to switch out the inputs as needed.
	GetInputs() []*wrapinput.Model

	// GetAllInputs will be called at the beginning for initial formatting to have a cohesive
	// look.
	GetAllInputs() []*wrapinput.Model

	// Validate gives the library a chance to convert any values that need to be converted
	// or make sure that all inputs have valid values.
	Validate(inputs []*wrapinput.Model) bool
	// Reset goes from Validation state to Edit state.
	Reset(inputs []*wrapinput.Model)

	// RecreateCommand should return the command the user can run to get back to the current state.
	RecreateCommand() string

	// GetRunController is called when the gathering inputs step has been completed. The view will
	// transition to the Run view, so the implementation should give the updated variable values here.
	GetRunController() Run
}

type Run interface {
	// Commands should return the commands to be run. The caller will modify the commands in place,
	// so the implementation will have access to the updated state for each command along the way.
	Commands() []*command.Model

	// RecreateCommand should return the command the user can run to get back to the current state.
	RecreateCommand() string

	// GetEndController being called signifies that the Run screen is now finished and will
	// proceed to the next screen.
	GetEndController() End
}

type End interface {
	// Name returns the name of the flow (e.g. gcp-install) and should be safe to write into
	// a file name.
	Name() string
	// The description for when we reach the end screen.
	EndDescription() string
	Commands() []*command.Model
}
