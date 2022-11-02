package button

import (
	"fmt"
	"strings"

	"flightcrew.io/cli/internal/style"
)

type Button struct {
	text string
}

func New(text string, width int) (*Button, error) {
	numPadding := width - len(text)
	if numPadding < 0 {
		return nil, fmt.Errorf("text length is less than width: %d vs %d", len(text), width)
	}

	leftPadding := (numPadding / 2)
	rightPadding := numPadding - leftPadding

	// Make space for the brackets.
	leftPadding--
	rightPadding--

	return &Button{
		text: fmt.Sprintf("[%s%s%s]",
			strings.Repeat(" ", leftPadding),
			text,
			strings.Repeat(" ", rightPadding)),
	}, nil
}

func (b *Button) View(focused bool) string {
	if focused {
		return style.Focused.Render(b.text)
	}

	return style.Blurred.Render(b.text)
}
