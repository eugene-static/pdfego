package template

import (
	"strings"
)

func (core *Core) splitText(text, font string, size, width float64) []string {
	lines := make([]string, 0)

	for seg := range strings.Lines(text) {
		lines = append(lines, core.splitSegment(seg, font, size, width)...)
	}

	return lines
}

func (core *Core) splitSegment(text, font string, size, width float64) []string {
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
