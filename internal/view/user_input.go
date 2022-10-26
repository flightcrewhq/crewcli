package view

import (
	"strings"

	"flightcrew.io/cli/internal/style"
	tea "github.com/charmbracelet/bubbletea"
)

type userInput struct {
	yes bool
}

func NewUserInput() *userInput {
	return &userInput{
		yes: false,
	}
}

func (ui *userInput) Value() bool {
	return ui.yes
}

func (ui *userInput) Reset() {
	ui.yes = false
}

func (ui *userInput) View() string {
	var b strings.Builder
	if ui.yes {
		b.WriteString(style.Focused.Render(confirmYesText))
	} else {
		b.WriteString(style.Blurred.Render(confirmYesText))
	}

	b.WriteRune(' ')

	if ui.yes {
		b.WriteString(style.Blurred.Render(confirmNoText))
	} else {
		b.WriteString(style.Focused.Render(confirmNoText))
	}

	return b.String()
}

func (ui *userInput) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return tea.Quit
		case "left":
			if !ui.yes {
				ui.yes = true
			}
			return nil
		case "right":
			if ui.yes {
				ui.yes = false
			}
			return nil
		case "enter", "tab":

		}
	}
	return nil
}
