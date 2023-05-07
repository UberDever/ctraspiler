package syntax

import (
	"fmt"

	"golang.org/x/exp/utf8string"
)

type source struct {
	file   string
	text   utf8string.String
	tokens []token
}

func NewSource(filename string, text utf8string.String) source {
	return source{file: filename, text: text}
}

func (s source) Lexeme(index tokenIndex) string {
	t := s.token(index)
	return s.text.Slice(int(t.start), int(t.end+1))
}

func (s source) token(index tokenIndex) token {
	return s.tokens[index]
}

func (s source) traceToken(tag tokenTag, lexeme string, line int, col int) string {
	str := fmt.Sprintf("\ttag = %d\n", tag)
	if lexeme != "" {
		str += fmt.Sprintf("\tlexeme = %#v\n", lexeme)
	}
	if line != -1 && col != -1 {
		str += "\tloc = " + fmt.Sprintf("%d", line) + ":" + fmt.Sprintf("%d", col) + "\n"
	}
	return str
}
