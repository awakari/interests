package model

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGroupLogic_String(t *testing.T) {
	cases := map[string]int{
		"And": 0,
		"Or":  1,
		"Xor": 2,
	}
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, k, GroupLogic(c).String())
		})
	}
}
