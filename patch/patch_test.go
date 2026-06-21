package patch

import (
	"slices"
	"testing"
)

func Test_ExtractParametersReferenceObjectNumbers(t *testing.T) {
	tests := []struct {
		name string
		body []byte
		want parameterReference
	}{
		{
			name: "/Kids_1",
			body: []byte("<</Type/Pages/Count 2/Kids_1[ 3 0 R 28 0 R] >>"),
			want: parameterReference{
				startIndex:    21,
				endIndex:      43,
				objectNumbers: []int{3, 28},
			},
		},
		{
			name: "/Kids_2",
			body: []byte("<</Type/Pages/Kids_2[ 21 0 R ]/Count 1>>"),
			want: parameterReference{
				startIndex:    13,
				endIndex:      30,
				objectNumbers: []int{21},
			},
		},
		{
			name: "/Kids_3",
			body: []byte("<</Type /Pages/Count 2/Kids_3[5 0 R 16 0 R ]/Resources<</Font<</FAAAAJ 9 0 R/FAAABD 13 0 R>>/XObject<</X1 7 0 R/X2 18 0 R>>>>>>"),
			want: parameterReference{
				startIndex:    22,
				endIndex:      44,
				objectNumbers: []int{5, 16},
			},
		},
		{
			name: "/Resources",
			body: []byte(`
				<<
				/Contents [ 26 0 R ]
				/CropBox [ 0.0 0.0 595.32001 841.92004 ]
				/MediaBox [ 0.0 0.0 595.32001 841.92004 ]
				/Parent 2 0 R
				/Resources 27 0 R
				/Rotate 0
				/Type /Page
				>>
			`),
			want: parameterReference{
				startIndex:    146,
				endIndex:      163,
				objectNumbers: []int{27},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := extractParametersReferenceObjectNumbers(test.body, 0, test.name)
			if err != nil {
				t.Error(err)

				return
			}

			if !slices.Equal(test.want.objectNumbers, got.objectNumbers) {
				t.Errorf("want: %v, got: %v", test.want, got)
			}

			if test.want.startIndex != got.startIndex {
				t.Errorf("want: %v, got: %v", test.want, got)
			}

			if test.want.endIndex != got.endIndex {
				t.Errorf("want: %v, got: %v", test.want, got)
			}

			t.Logf("Parameter: %v", got)
			t.Logf("Concat: %s", test.body[got.startIndex:got.endIndex])
		})
	}
}
