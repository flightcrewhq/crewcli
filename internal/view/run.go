package view

import (
	"strings"

	"flightcrew.io/cli/internal/controller"
	"flightcrew.io/cli/internal/debug"
	"flightcrew.io/cli/internal/style"
	"flightcrew.io/cli/internal/view/button"
	"flightcrew.io/cli/internal/view/command"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/paginator"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type cmdFinishedErr struct {
	err error
}

type RunModel struct {
	commands  []*command.Model
	yesButton *button.Button
	noButton  *button.Button

	controller controller.Run
	paginator  paginator.Model
	index      int

	userInput bool
}

func NewRunModel(controller controller.Run) *RunModel {
	debug.Output("New run model time!")

	yesButton, _ := button.New("yes", 10)
	noButton, _ := button.New("no", 10)

	m := &RunModel{
		controller: controller,
		commands:   controller.Commands(),
		yesButton:  yesButton,
		noButton:   noButton,
	}

	p := paginator.New()
	p.Type = paginator.Dots
	p.PerPage = 1
	p.KeyMap.PrevPage = key.NewBinding(key.WithKeys("h"))
	p.KeyMap.NextPage = key.NewBinding(key.WithKeys("l"))
	p.ActiveDot = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "235", Dark: "252"}).Render("•")
	p.InactiveDot = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "250", Dark: "238"}).Render("•")
	p.SetTotalPages(len(m.commands))
	m.paginator = p

	m.nextCommand()

	return m
}

func (m *RunModel) Init() tea.Cmd {
	return nil
}

func (m *RunModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.index >= len(m.commands) {
		printRecreatedCommand(m.controller.RecreateCommand())
		return m, tea.Quit
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		// Allow user to quit at any time.
		case "ctrl+c", "esc":
			printRecreatedCommand(m.controller.RecreateCommand())
			return m, tea.Quit

		case "h":
			var cmd tea.Cmd
			m.paginator, cmd = m.paginator.Update(msg)
			return m, cmd

		case "l":
			// Don't allow user to advance to future commands that haven't been run or prompted yet.
			if m.paginator.Page < m.index {
				var cmd tea.Cmd
				m.paginator, cmd = m.paginator.Update(msg)
				return m, cmd
			}
		}
	}

	cmd := m.commands[m.index]
	switch cmd.State() {
	case command.PromptState:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "enter":
				if cmd.State() == command.PromptState && !m.userInput {
					return m, tea.Quit
				}

				if cmd.State() == command.PromptState && m.userInput {
					m.userInput = false
					return m, tea.Exec(cmd.GetCommandToRun(), func(err error) tea.Msg {
						return cmdFinishedErr{err}
					})
				}

			case "tab":
				if cmd.State() == command.PromptState {
					m.userInput = !m.userInput
				}

			case "left":
				if cmd.State() == command.PromptState && !m.userInput {
					m.userInput = true
				}

			case "right":
				if cmd.State() == command.PromptState && m.userInput {
					m.userInput = false
				}
			}
		}

	case command.RunningState:
		switch msg := msg.(type) {
		case cmdFinishedErr:
			cmd := m.commands[m.index]
			cmd.Complete(msg.err == nil)
			return m, nil
		}
		return m, nil

	case command.PassState:
		switch msg.(type) {
		case tea.KeyMsg:
			if !m.nextCommand() {
				return NewEndModel(m.controller.GetEndController()), nil
			}

			return m, nil
		}

	case command.FailState:
		switch msg.(type) {
		case tea.KeyMsg:
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m *RunModel) View() string {
	var b strings.Builder
	cmd := m.commands[m.paginator.Page]
	b.WriteRune('\n')

	b.WriteString(lipgloss.NewStyle().PaddingLeft(2).Render(m.paginator.View()))
	b.WriteRune('\n')

	b.WriteString(cmd.View())
	b.WriteString("\n\n")

	if cmd.State() == command.PromptState {
		b.WriteString(style.Action("[ACTION REQUIRED]"))
		b.WriteString(" Run the command? ")
		b.WriteString(m.yesButton.View(m.userInput))
		b.WriteString("  ")
		b.WriteString(m.noButton.View(!m.userInput))
		b.WriteRune('\n')
	}

	if cmd.State() == command.FailState && !cmd.IsRead() {
		b.WriteString(style.Help("(press any key to quit)"))
	} else if cmd.State() == command.PassState {
		b.WriteString(style.Help("(press any key to continue)"))
	} else if cmd.State() == command.SkipState {
		b.WriteString(style.Help("(nothing to do, press any key to quit)"))
	} else {
		b.WriteString(style.Help("ctrl+c/esc: quit • h/l: page • ←/→/enter: run command"))
	}

	b.WriteRune('\n')
	return b.String()
}

func (m *RunModel) nextCommand() bool {
	for ; m.index < len(m.commands); m.index++ {
		m.paginator.Page = m.index
		current := m.commands[m.index]
		if current.State() != command.NoneState {
			continue
		}

		if current.ShouldPrompt() {
			return true
		}
	}

	return false
}
