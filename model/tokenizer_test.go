package model

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTokenizer(t *testing.T) {
	cases := []struct {
		input   string
		lexemes []string
	}{
		{
			input: "quick  brown\tfox\njumps over a lazy dog!",
			lexemes: []string{
				"quick",
				"brown",
				"fox",
				"jumps",
				"over",
				"a",
				"lazy",
				"dog",
				"!",
			},
		},
		{
			input: "Летом не припасёшь, зимой не принесёшь.",
			lexemes: []string{
				"Летом",
				"не",
				"припасёшь",
				",",
				"зимой",
				"не",
				"принесёшь",
				".",
			},
		},
		{
			input: "Ei kala miestä hae, jollei mies kalaa",
			lexemes: []string{
				"Ei",
				"kala",
				"miestä",
				"hae",
				",",
				"jollei",
				"mies",
				"kalaa",
			},
		},
		{
			input: "养军千日，用军一朝。",
			lexemes: []string{
				"养军千日",
				"，",
				"用军一朝",
				"。",
			},
		},
		{
			input: "3.1415926 1.2e+34 2022-10-22 22:11:45",
			lexemes: []string{
				"3.1415926",
				"1.2e+34",
				"2022",
				"-",
				"10",
				"-",
				"22",
				"22",
				":",
				"11",
				":",
				"45",
			},
		},
		{
			input:   "   ",
			lexemes: []string{},
		},
	}
	for _, c := range cases {
		t.Run(c.input, func(t *testing.T) {
			tknzr := NewTokenizer(c.input)
			var lexemes []string
			for {
				l, eoi := tknzr.Next()
				if eoi {
					break
				} else {
					lexemes = append(lexemes, l)
				}
			}
			assert.ElementsMatch(t, c.lexemes, lexemes)
		})
	}
}
