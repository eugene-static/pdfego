package pdfego

import (
	"slices"
	"testing"
)

func Test_ParseParameters(t *testing.T) {
	tests := []struct {
		name string
		body []byte
		want []string
	}{
		{
			name: "No Spaces",
			body: []byte(`<</Type/Pages/Kids_2[ 21 0 R ]/Count 1>>`),
			want: []string{"Type", "Kids_2", "Count"},
		},
		{
			name: "Inline",
			body: []byte(`<< /Contents [ 26 0 R ] /CropBox [ 0.0 0.0 595.32001 841.92004 ]/MediaBox [ 0.0 0.0 595.32001 841.92004 ] /Parent 2 0 R /Resources<</Font<</FAAAAJ 9 0 R/FAAABD 13 0 R>>/XObject<</X1 7 0 R/X2 18 0 R>>>>/Rotate 0 /Type /Page >>`),
			want: []string{"Type", "CropBox", "MediaBox", "Resources", "Parent", "Rotate"},
		},
		{
			name: "Large",
			body: []byte("<</Type /Page /Resources <</ProcSet [/PDF /Text /ImageB /ImageC /ImageI] /ExtGState <</G3 3 0 R /G4 4 0 R>> /Font <</F5 5 0 R /F6 6 0 R /F7 7 0 R /F8 8 0 R /F9 9 0 R /F10 10 0 R /F11 11 0 R /F14 14 0 R /F18 18 0 R /F19 19 0 R /F20 20 0 R>>>> /MediaBox [0 0 420 595.91998] /Annots [21 0 R] /Contents 22 0 R /StructParents 1 /Tabs /S /Parent 23 0 R>>"),
			want: []string{"Type", "Resources", "MediaBox", "Annots", "Contents", "StructParents", "Tabs", "Parent"},
		},
		{
			name: "Not Inline",
			body: []byte("25 0 R"),
			want: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseDictionary(tt.body)

			for _, want := range tt.want {
				value, ok := got[want]
				if !ok {
					t.Errorf("не найден параметр %s", want)

					continue
				}

				t.Logf("%s: %s", want, value)
			}
		})
	}
}

func Test_ParseParametersReference(t *testing.T) {
	tests := []struct {
		name      string
		parameter parameter
		want      parameterReferences
	}{
		{
			name: "Prefix Spaces Array",
			parameter: parameter{
				value: []byte("[ 3 0 R 28 0 R]"),
			},
			want: parameterReferences{
				objectNumbers: []int{3, 28},
			},
		},
		{
			name: "Around Spaces Array",
			parameter: parameter{
				value: []byte("[ 21 0 R ]"),
			},
			want: parameterReferences{
				objectNumbers: []int{21},
			},
		},
		{
			name: "Suffix Spaces Array",
			parameter: parameter{
				value: []byte("[5 0 R 16 0 R ]"),
			},
			want: parameterReferences{
				objectNumbers: []int{5, 16},
			},
		},
		{
			name: "Single",
			parameter: parameter{
				value: []byte("26 0 R"),
			},
			want: parameterReferences{
				objectNumbers: []int{26},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := test.parameter.parseReferences()
			if err != nil {
				t.Error(err)

				return
			}

			if !slices.Equal(test.want.objectNumbers, got.objectNumbers) {
				t.Errorf("want: %v, got: %v", test.want, got)
			}

			t.Logf("Parameter: %v", got)
		})
	}
}

func Test_ParseParameterArray(t *testing.T) {
	tests := []struct {
		name      string
		parameter parameter
		want      parameterNumberArray
	}{
		{
			name: "MediaBox",
			parameter: parameter{
				value: []byte("[0 0 420 595.91998]"),
			},
			want: parameterNumberArray{
				values: []float64{0, 0, 420, 595.91998},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := test.parameter.parseNumberArray()
			if err != nil {
				t.Error(err)
			}

			if !slices.Equal(got.values, test.want.values) {
				t.Errorf("want: %v, got: %v", test.want, got)
			}

			t.Logf("Parameter: %v", got)
		})
	}
}
