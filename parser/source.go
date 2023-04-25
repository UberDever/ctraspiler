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

func (s Source) trace(t Token) string {
	lexeme := s.lexeme(t)
	return fmt.Sprintf("%d/%s/%d:%d", t.tag, lexeme, t.line, t.col)
}
