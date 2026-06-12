package buffer

import (
	"fmt"
	"testing"
)

func Test_WriteInt64d10(t *testing.T) {
	b := New(0)

	cases := map[string]struct {
		value    int64
		expected string
	}{
		"65": {
			value:    65,
			expected: fmt.Sprintf("%010d", 65),
		},
		"100": {
			value:    100,
			expected: fmt.Sprintf("%010d", 100),
		},
		"0": {
			value:    0,
			expected: fmt.Sprintf("%010d", 0),
		},
	}

	for k, tc := range cases {
		t.Run(k, func(t *testing.T) {
			b.writeInt64D10(tc.value)

			got := b.content.String()

			if got != tc.expected {
				t.Errorf("AppendInt10(%d): got %q, want %q", tc.value, got, tc.expected)
			}
		})
	}
}
