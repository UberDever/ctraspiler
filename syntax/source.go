package syntax

import (
	"fmt"

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

func (s Source) Location(index TokenIndex) (line, col int) {
	t := s.token(index)
	line = t.line
	col = t.col
	return
}

func (s Source) Lexeme(index TokenIndex) string {
	t := s.token(index)
	return s.text.Slice(int(t.start), int(t.end+1))
}

func (s Source) Filename() string {
	return s.filename
}

func (s Source) token(index TokenIndex) token {
	return s.tokens[index]
}

func (s Source) traceToken(tag TokenTag, lexeme string, line int, col int) string {
	str := fmt.Sprintf("\ttag = %d\n", tag)
	if lexeme != "" {
		str += fmt.Sprintf("\tlexeme = %#v\n", lexeme)
	}
	if line != -1 && col != -1 {
		str += "\tloc = " + fmt.Sprintf("%d", line) + ":" + fmt.Sprintf("%d", col) + "\n"
	}
	return str
}
