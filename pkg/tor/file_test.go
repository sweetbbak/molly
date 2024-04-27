package tor

import (
	"log"
	"slices"
	"testing"
)

func TestParseIndices(t *testing.T) {
	var tests = []struct {
		input string
		want  []int
	}{
		{"0,3,4-5,10-15", []int{0, 3, 4, 5, 10, 11, 12, 13, 14, 15}},
		{"0,100,20-23", []int{0, 20, 21, 22, 23, 100}},
		{"1,0,3", []int{0, 1, 3}},
		{"0-5", []int{0, 1, 2, 3, 4, 5}},
		{"0,0,0,1,0", []int{0, 1}},
	}

	var fails = []struct {
		input string
		want  []int
	}{
		{"2,3,100-1", []int{}},
		{"", []int{}},
		{"1,0,3,", []int{0, 1, 3}},
		{"100-1", []int{}},
		{"x,y,z", []int{}},
	}

	for _, tt := range tests {
		t.Run("Parse Indices", func(t *testing.T) {
			i, err := ParseIndices(tt.input)
			if err != nil {
				log.Println(err)
			}

			if slices.Equal(i, tt.want) {
				log.Printf("Slices are equal: %v %v\n", i, tt.want)
			} else {
				t.Errorf("Slices are NOT equal %v %v\n", i, tt.want)
			}
		})
	}

	for _, tt := range fails {
		t.Run("Parse Indices fails", func(t *testing.T) {
			i, err := ParseIndices(tt.input)
			if err != nil {
				log.Printf("successful error: %v %v\n", err, i)

			} else {
				t.Errorf("test should have failed but did not [%v]\n", i)
			}
		})
	}

}
