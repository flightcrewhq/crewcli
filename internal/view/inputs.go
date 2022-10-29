package view

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Inputs is a variable number of inputs that can be changed. This interface allows
// pre-defined defaults that can be modified according to cloud providers' specific
// requirements.
type Inputs interface {
	// Len returns the number of inputs in this grouping.
	Len() int
	// Validate gives the library a chance to convert any values that need to be converted
	// or make sure that all inputs have valid values.
	Validate() bool
	Reset()
	// Args turns the inputs into a map from key to value so that the inputted values can
	// be replaced in the subsequent commands to be run.
	Args() map[string]string

	View() string
	Update(msg tea.Msg) tea.Cmd

	Focus(i int) tea.Cmd
	NextEmpty(i int) int
}
