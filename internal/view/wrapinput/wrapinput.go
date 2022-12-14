package wrapinput

import (
	"strings"

	"flightcrew.io/cli/internal/style"
	"flightcrew.io/cli/internal/view/radioinput"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type ValidateParams struct {
	Converted    string
	InfoMessage  *string
	ErrorMessage string
}

type Model struct {
	// Types of inputs. Pointers so that we can tell which one is being used.
	Freeform *textinput.Model
	Radio    *radioinput.Model

	Title    string
	HelpText string
	Default  string

	validation ValidateParams
	validating bool

	Required bool
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
	ShowValue bool
}

func (m Model) View(params ViewParams) string {
	var b strings.Builder

	b.WriteString(m.Title)
	if m.Required {
		b.WriteString(style.Required("*"))
	} else {
		b.WriteRune(' ')
	}
	b.WriteString(": ")

	if params.ShowValue {
		if m.Radio != nil {
			b.WriteString(m.Radio.Value())
		} else if m.Freeform != nil {
			if val := m.Freeform.Value(); len(val) > 0 {
				b.WriteString(m.Freeform.Value())
			} else {
				b.WriteString(m.Default)
			}
		}
		if len(m.validation.ErrorMessage) > 0 {
			b.WriteString(" ❗️ ")
			b.WriteString(style.Error(m.validation.ErrorMessage))
		} else if m.validation.InfoMessage != nil {
			if len(*m.validation.InfoMessage) > 0 {
				b.WriteString(" → ")
				b.WriteString(style.Convert(*m.validation.InfoMessage))
			}
		} else if len(m.validation.Converted) > 0 {
			b.WriteString(" → ")
			b.WriteString(style.Convert(m.validation.Converted))
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
	if len(m.validation.Converted) > 0 {
		return m.validation.Converted
	}

	if m.Radio != nil {
		if val := m.Radio.Value(); len(val) > 0 {
			return val
		}
	}

	if m.Freeform != nil {
		if val := m.Freeform.Value(); len(val) > 0 {
			return val
		}
	}

	return m.Default
}

func (m *Model) SetInfo(infoMsg string) {
	m.validating = true
	m.validation.InfoMessage = &infoMsg
}

func (m *Model) SetConverted(convertedVal string) {
	m.validating = true
	m.validation.Converted = convertedVal
}

func (m *Model) SetError(err error) {
	m.validating = true
	if err != nil {
		m.validation.ErrorMessage = err.Error()
	}
}

func (m *Model) ResetValidation() {
	m.validating = false
	m.validation = ValidateParams{}
}
