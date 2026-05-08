package template

import (
	"bytes"
	"fmt"
	"testing"
	"unicode/utf16"
)

func Test_WriteInt64d10(t *testing.T) {
	b := new(bytes.Buffer)

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
			writeInt64D10(b, tc.value)

			got := b.String()

			if got != tc.expected {
				t.Errorf("AppendInt10(%d): got %q, want %q", tc.value, got, tc.expected)
			}
		})
	}
}

func Test_Rowspan(t *testing.T) {
	b := new(bytes.Buffer)

	cases := map[string]struct {
		value    string
		expected string
	}{
		"привет": {
			value:    "П",
			expected: fmt.Sprintf("%04X", "П"),
		},
	}

	for k, tc := range cases {
		t.Run(k, func(t *testing.T) {
			idxs := utf16.Encode([]rune(tc.value))
			writeUint16D4(b, idxs[0])

			got := b.String()

			if got != tc.expected {
				t.Errorf("AppendUint16D4(%s): got %q, want %q", tc.value, got, tc.expected)
			}
		})
	}

}
