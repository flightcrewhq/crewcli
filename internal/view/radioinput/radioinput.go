package radioinput

import (
	"fmt"
	"strings"

	"flightcrew.io/cli/internal/style"
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	prevKeys     map[string]struct{}
	nextKeys     map[string]struct{}
	options      []string
	currentIndex int
	focused      bool
}

func NewModel(opts []string) Model {
	return Model{
		options:      opts,
		currentIndex: 0,
		prevKeys:     map[string]struct{}{"left": {}},
		nextKeys:     map[string]struct{}{"right": {}},
		focused:      false,
	}
}

func (m *Model) SetPrevKeys(msgs []string) {
	m.prevKeys = make(map[string]struct{})
	for _, k := range msgs {
		m.prevKeys[k] = struct{}{}
	}
}

func (m *Model) SetNextKeys(msgs []string) {
	m.nextKeys = make(map[string]struct{})
	for _, k := range msgs {
		m.nextKeys[k] = struct{}{}
	}
}

func (m Model) Value() string {
	return m.options[m.currentIndex]
}

func (m *Model) SetValue(val string) {
	for i := 0; i < len(m.options); i++ {
		if m.options[i] == val {
			m.currentIndex = i
			return
		}
	}
	panic(fmt.Sprintf("Invalid default value passed in. Got '%s', but only have '%s'", val, strings.Join(m.options, "', '")))
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.focused {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		s := msg.String()
		if _, ok := m.prevKeys[s]; ok {
			m.currentIndex--
		} else if _, ok := m.nextKeys[s]; ok {
			m.currentIndex++
		}

		if m.currentIndex < 0 {
			m.currentIndex = len(m.options) - 1
		} else if m.currentIndex >= len(m.options) {
			m.currentIndex = 0
		}
	}

	return m, nil
}

func (m Model) View() string {
	var b strings.Builder
	if m.focused {
		b.WriteString(style.Focused.Render("> "))
	} else {
		b.WriteString("> ")
	}

	numOpts := len(m.options)
	for i, opt := range m.options {
		if i == m.currentIndex {
			if m.focused {
				b.WriteString(style.Highlight(opt))
			} else {
				b.WriteString(style.BlurHighlight(opt))
			}
		} else {
			b.WriteString(style.Blurred.Render(opt))
		}
		if i < numOpts-1 {
			b.WriteString(" • ")
		}
	}

	return b.String()
}

func (m *Model) Focus() {
	m.focused = true
}

func (m *Model) Blur() {
	m.focused = false
}
