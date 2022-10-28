package wrapinput

import (
	"strings"

	"flightcrew.io/cli/internal/style"
	"flightcrew.io/cli/internal/view/radioinput"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	Title    string
	HelpText string
	Required bool
	Default  string
	// If the value is to be converted, this is only valid when model.confirming is true.
	Converted string
	Message   string
	Error     string
	Freeform  *textinput.Model
	Radio     *radioinput.Model
}

func NewRadio(options []string) Model {
	var radio = radioinput.NewModel(options)
	return Model{
		Radio: &radio,
	}
}

func NewFreeForm() Model {
	var freeform = textinput.New()
	freeform.CursorStyle = style.Focused.Copy()
	freeform.CharLimit = 32
	return Model{
		Freeform: &freeform,
	}
}

type ViewParams struct {
	ShowValue  bool
	TitleStyle lipgloss.Style
}

func (m Model) View(params ViewParams) string {
	var b strings.Builder

	b.WriteString(params.TitleStyle.Render(m.Title))
	if m.Required {
		b.WriteString(style.Required("*"))
	} else {
		b.WriteRune(' ')
	}
	b.WriteString(": ")

	if params.ShowValue {
		if m.Radio != nil {
			b.WriteString(m.Radio.Value())
		} else {
			b.WriteString(m.Freeform.Value())
		}

		if len(m.Error) > 0 {
			b.WriteString(" ❗️ ")
			b.WriteString(style.Error(m.Error))
		} else if len(m.Message) > 0 {
			b.WriteString(" → ")
			b.WriteString(style.Convert(m.Message))
		} else if len(m.Converted) > 0 {
			b.WriteString(" → ")
			b.WriteString(style.Convert(m.Converted))
		}

	} else {
		if m.Radio != nil {
			b.WriteString(m.Radio.View())
		} else {
			b.WriteString(m.Freeform.View())
		}
	}

	return b.String()
}

func (m *Model) Focus() tea.Cmd {
	if m.Radio != nil {
		m.Radio.Focus()
		return nil
	} else if m.Freeform != nil {
		cmd := m.Freeform.Focus()
		m.Freeform.PromptStyle = style.Focused
		m.Freeform.TextStyle = style.Focused
		return cmd
	}

	return nil
}

func (m *Model) Blur() {
	if m.Radio != nil {
		m.Radio.Blur()
	} else if m.Freeform != nil {
		m.Freeform.Blur()
		m.Freeform.PromptStyle = style.None
		m.Freeform.TextStyle = style.None
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	if m.Radio != nil {
		*m.Radio, cmd = m.Radio.Update(msg)
	} else if m.Freeform != nil {
		*m.Freeform, cmd = m.Freeform.Update(msg)
	}
	return m, cmd
}

func (m *Model) SetValue(val string) {
	if m.Radio != nil {
		m.Radio.SetValue(val)
		return
	}

	if m.Freeform != nil {
		m.Freeform.SetValue(val)
		return
	}
}

func (m Model) Value() string {
	if m.Radio != nil {
		return m.Radio.Value()
	}

	if m.Freeform != nil {
		if len(m.Converted) > 0 {
			return m.Converted
		}

		if val := m.Freeform.Value(); len(val) > 0 {
			return val
		}

		return m.Default
	}

	return ""
}

func (m *Model) Validate() {
}

func (m *Model) ResetValidation() {
	m.Converted = ""
	m.Error = ""
	m.Message = ""
}
