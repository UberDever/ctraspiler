package syntax

import (
	"fmt"
	ID "some/domain"

	"golang.org/x/exp/utf8string"
)

type Source struct {
	filename string
	text     utf8string.String
	tokens   []token
}

func NewSource(filename string, text utf8string.String) Source {
	return Source{filename: filename, text: text}
}

func (s Source) Location(id ID.Token) (line, col int) {
	t := s.Token(id)
	line = t.Line
	col = t.Col
	return
}

func (s Source) Lexeme(id ID.Token) string {
	t := s.Token(id)
	return s.text.Slice(int(t.Start), int(t.End+1))
}

func (s Source) Filename() string {
	return s.filename
}

func (s Source) Token(id ID.Token) token {
	return s.tokens[id]
}

func (s Source) TraceToken(tag ID.Token, lexeme string, line int, col int) string {
	str := fmt.Sprintf("\ttag = %d\n", tag)
	if lexeme != "" {
		str += fmt.Sprintf("\tlexeme = %#v\n", lexeme)
	}
	if line != -1 && col != -1 {
		str += "\tloc = " + fmt.Sprintf("%d", line) + ":" + fmt.Sprintf("%d", col) + "\n"
	}
	return str
}
