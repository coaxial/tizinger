package helpers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUniq(t *testing.T) {
	tests := []struct {
		input []int
		want  []int
		msg   string
	}{
		{
			[]int{1, 2, 3, 4, 5, 6},
			[]int{1, 2, 3, 4, 5, 6},
			"should not remove any element",
		},
		{
			[]int{1, 2, 2, 3},
			[]int{1, 2, 3},
			"should remove duplicate element",
		},
		{
			[]int{1, 4, 2, 3, 2},
			[]int{1, 4, 2, 3},
			"should remove duplicate element",
		},
	}

	for _, test := range tests {
		got := Uniq(test.input)
		assert.Equal(t, test.want, got, test.msg)
	}
}
