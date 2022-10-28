package view

import (
	"fmt"
	"strings"

	"flightcrew.io/cli/internal/style"
)

type HorizontalSelector struct {
	options      []string
	currentIndex int
}

func NewHorizontalSelector(opts []string) *HorizontalSelector {
	return &HorizontalSelector{
		options:      opts,
		currentIndex: 0,
	}
}

func (s *HorizontalSelector) Value() string {
	return s.options[s.currentIndex]
}

func (s *HorizontalSelector) MoveLeft() {
	s.currentIndex--
	if s.currentIndex < 0 {
		s.currentIndex = len(s.options) - 1
	}
}

func (s *HorizontalSelector) MoveRight() {
	s.currentIndex++
	if s.currentIndex >= len(s.options) {
		s.currentIndex = 0
	}
}

func (s *HorizontalSelector) SetValue(val string) {
	for i := 0; i < len(s.options); i++ {
		if s.options[i] == val {
			s.currentIndex = i
			return
		}
	}
	panic(fmt.Sprintf("Invalid default value passed in. Got '%s', but only have '%s'", val, strings.Join(s.options, "', '")))
}

func (s *HorizontalSelector) View(focused bool) string {
	var b strings.Builder
	if focused {
		b.WriteString(style.Focused.Render("> "))
	} else {
		b.WriteString("> ")
	}

	numOpts := len(s.options)
	for i, opt := range s.options {
		if i == s.currentIndex {
			if focused {
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
