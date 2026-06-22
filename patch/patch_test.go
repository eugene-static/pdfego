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
			name: "Kids_1",
			body: []byte("<</Type/Pages/Count 2/Kids_1[ 3 0 R 28 0 R] >>"),
			want: parameterReference{
				start:         21,
				end:           43,
				objectNumbers: []int{3, 28},
			},
		},
		{
			name: "Kids_2",
			body: []byte("<</Type/Pages/Kids_2[ 21 0 R ]/Count 1>>"),
			want: parameterReference{
				start:         13,
				end:           30,
				objectNumbers: []int{21},
			},
		},
		{
			name: "Kids_3",
			body: []byte("<</Type /Pages/Count 2/Kids_3[5 0 R 16 0 R ]/Resources<</Font<</FAAAAJ 9 0 R/FAAABD 13 0 R>>/XObject<</X1 7 0 R/X2 18 0 R>>>>>>"),
			want: parameterReference{
				start:         22,
				end:           44,
				objectNumbers: []int{5, 16},
			},
		},
		{
			name: "Resources",
			body: []byte("<<\n/Contents [ 26 0 R ]\n/CropBox [ 0.0 0.0 595.32001 841.92004 ]\n/MediaBox [ 0.0 0.0 595.32001 841.92004 ]\n/Parent 2 0 R\n/Resources 27 0 R\n/Rotate 0\n/Type /Page\n>>\n"),
			want: parameterReference{
				start:         121,
				end:           139,
				objectNumbers: []int{27},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := parseParametersReference(test.body, test.name)
			if err != nil {
				t.Error(err)

				return
			}

			if !slices.Equal(test.want.objectNumbers, got.objectNumbers) {
				t.Errorf("want: %v, got: %v", test.want, got)
			}

			if test.want.start != got.start {
				t.Errorf("want: %v, got: %v", test.want, got)
			}

			if test.want.end != got.end {
				t.Errorf("want: %v, got: %v", test.want, got)
			}

			t.Logf("Parameter: %v", got)
			t.Logf("Concat: %s", test.body[got.start:got.end])
		})
	}
}
