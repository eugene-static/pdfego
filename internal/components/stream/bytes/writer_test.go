package bytes

import (
	"fmt"
	"testing"
)

func Test_WriteUint64FixedSize(t *testing.T) {
	cases := map[string]struct {
		value    uint
		size     int
		expected string
	}{
		"65": {
			value:    65,
			size:     10,
			expected: fmt.Sprintf("%010d", 65),
		},
		"100": {
			value:    100,
			size:     5,
			expected: fmt.Sprintf("%05d", 100),
		},
		"0": {
			value:    0,
			size:     10,
			expected: fmt.Sprintf("%010d", 0),
		},
		"15": {
			value:    15,
			size:     0,
			expected: "",
		},
	}

	for k, tc := range cases {
		dst := make([]byte, 0, 50)
		bw := NewWriter(dst)

		t.Run(k, func(t *testing.T) {

			bw.WriteUintFixed(tc.value, tc.size)

			got := bw.Bytes()

			if string(got) != tc.expected {
				t.Errorf("writeUint64FixedSize(%d): got %q, want %q", tc.value, got, tc.expected)
			}
		})
	}
}
