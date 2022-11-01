package view

import (
	"strings"

	"flightcrew.io/cli/internal/debug"
	"flightcrew.io/cli/internal/style"
	"flightcrew.io/cli/internal/view/button"
	"flightcrew.io/cli/internal/view/command"
	"github.com/charmbracelet/bubbles/paginator"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type LazyCommands func() []*command.Model

type cmdFinishedErr struct {
	err error
}

type runModel struct {
	params    map[string]string
	paginator paginator.Model

	commands     []*command.Model
	currentIndex int

	userInput bool
	yesButton *button.Button
	noButton  *button.Button

	quitting bool
}

func NewRunModel(params map[string]string, getCommands LazyCommands) *runModel {
	debug.Output("New run model time!")

	yesButton, _ := button.New("yes", 10)
	noButton, _ := button.New("no", 10)

	m := &runModel{
		params:    params,
		commands:  getCommands(),
		yesButton: yesButton,
		noButton:  noButton,
	}

	replaceArgs := make([]string, 0, 2*len(params))
	for key, param := range params {
		replaceArgs = append(replaceArgs, key, param)
	}
	replacer := strings.NewReplacer(replaceArgs...)

	for i := range m.commands {
		m.commands[i].Replace(replacer)
	}

	p := paginator.New()
	p.Type = paginator.Dots
	p.PerPage = 1
	p.UseHLKeys = true
	p.UseJKKeys = false
	p.UseLeftRightKeys = false
	p.UsePgUpPgDownKeys = false
	p.UseUpDownKeys = false
	p.ActiveDot = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "235", Dark: "252"}).Render("•")
	p.InactiveDot = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "250", Dark: "238"}).Render("•")
	p.SetTotalPages(len(m.commands))
	m.paginator = p

	m.nextCommand()

	return m
}

func (m *runModel) Init() tea.Cmd {
	return nil
}

func (m *runModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		// Allow user to quit at any time.
		case "ctrl+c", "esc":
			return m, tea.Quit

		case "h":
			var cmd tea.Cmd
			m.paginator, cmd = m.paginator.Update(msg)
			return m, cmd

		case "l":
			// Don't allow user to advance to future commands that haven't been run or prompted yet.
			if m.paginator.Page < m.currentIndex {
				var cmd tea.Cmd
				m.paginator, cmd = m.paginator.Update(msg)
				return m, cmd
			}
		}
	}

	cmd := m.commands[m.currentIndex]
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
			cmd := m.commands[m.currentIndex]
			cmd.Complete(msg.err == nil)
			return m, nil
			// TODO if we're at the end --> go to another screen

		}
		return m, nil

	case command.PassState:
		switch msg.(type) {
		case tea.KeyMsg:
			if m.quitting {
				return m, tea.Quit
			}

			if !m.nextCommand() {
				m.quitting = true
				return m, nil
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

func (m *runModel) View() string {
	var b strings.Builder
	cmd := m.commands[m.paginator.Page]
	b.WriteString("State is currently: ")
	b.WriteString(string(cmd.State()))
	b.WriteRune('\n')
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
	} else {
		b.WriteString(style.Help("ctrl+c/esc: quit • h/l: page • ←/→/enter: run command"))
	}

	b.WriteRune('\n')
	return b.String()
}

func (m *runModel) nextCommand() bool {
	for ; m.currentIndex < len(m.commands); m.currentIndex++ {
		m.paginator.Page = m.currentIndex
		current := m.commands[m.currentIndex]
		if current.State() != command.NoneState {
			continue
		}

		if current.ShouldRun() {
			return true
		}
	}

	return false
}
