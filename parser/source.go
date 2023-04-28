package parser

import (
	"fmt"

	"golang.org/x/exp/utf8string"
)

type Source struct {
	text   utf8string.String
	tokens []Token
}

func (s Source) lexeme(t Token) string {
	return s.text.Slice(int(t.start), int(t.end+1))
}

func (s Source) token(index int) Token {
	return s.tokens[index]
}

func (s Source) trace(tag Tag, lexeme string, line int, col int) string {
	str := fmt.Sprintf("\ttag = %d\n", tag)
	if lexeme != "" {
		str += "\tlexeme = " + lexeme + "\n"
	}
	if line != -1 && col != -1 {
		str += "\tloc = " + fmt.Sprintf("%d", line) + ":" + fmt.Sprintf("%d", col) + "\n"
	}
	return str
}
