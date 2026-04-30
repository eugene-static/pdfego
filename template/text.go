package template

import (
	"strings"
)

func (core *Core) splitText(font, text string, size, width float64) []string {
	segments := strings.Split(text, "\n")

	lines := make([]string, 0, len(segments))

	for _, seg := range segments {
		lines = append(lines, core.splitSegment(font, seg, size, width)...)
	}

	return lines
}

func (core *Core) splitSegment(font, text string, size, width float64) []string {
	words := strings.Fields(text)

	lines := make([]string, 0, len(words))

	line := words[0]

	for _, word := range words[1:] {
		candidate := line + " " + word

		candidateWidth := core.measureText(font, size, candidate)
		if candidateWidth > width {
			lines = append(lines, line)

			line = word

			continue
		}

		line = candidate
	}

	lines = append(lines, line)

	return lines
}
