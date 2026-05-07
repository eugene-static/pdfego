package template

import (
	"bytes"
	"fmt"
	"testing"
)

func Test_AppendInt10(t *testing.T) {
	b := new(bytes.Buffer)

	cases := map[string]struct {
		value    int64
		expected string
	}{
		"65": {
			value:    65,
			expected: fmt.Sprintf("%10d", 65),
		},
	}

	for k, tc := range cases {
		t.Run(k, func(t *testing.T) {
			appendInt10(b, tc.value)

			got := b.String()

			if got != tc.expected {
				t.Errorf("AppendInt10(%d): got %q, want %q", tc.value, got, tc.expected)
			}
		})
	}
}
