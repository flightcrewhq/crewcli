package view

import (
	"github.com/charmbracelet/bubbles/textinput"
)

type wrappedInput struct {
	Title    string
	HelpText string
	Required bool
	Default  string
	Input    textinput.Model
}
