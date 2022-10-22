package model

import (
	"strings"
	"text/scanner"
)

type (

	// Tokenizer splits the input string to the lexemes those may be iterated using the Next function.
	Tokenizer interface {

		// Next returns the next lexeme from the provided input, 2nd output parameter is true if no more lexemes left.
		Next() (lexeme string, endOfInput bool)
	}

	tokenizer struct {
		s scanner.Scanner
	}
)

func NewTokenizer(input string) Tokenizer {
	var s scanner.Scanner
	s.Init(strings.NewReader(input))
	return &tokenizer{
		s: s,
	}
}

func (t *tokenizer) Next() (lexeme string, endOfInput bool) {
	token := t.s.Scan()
	endOfInput = token == scanner.EOF
	if !endOfInput {
		lexeme = t.s.TokenText()
	}
	return
}
