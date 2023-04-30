package parser

import (
	"fmt"
	"strings"

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

func (s Source) trace(tag TokenTag, lexeme string, line int, col int) string {
	str := fmt.Sprintf("\ttag = %d\n", tag)
	if lexeme != "" {
		str += fmt.Sprintf("\tlexeme = %#v\n", lexeme)
	}
	if line != -1 && col != -1 {
		str += "\tloc = " + fmt.Sprintf("%d", line) + ":" + fmt.Sprintf("%d", col) + "\n"
	}
	return str
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func max(x, y int) int {
	if x > y {
		return x
	}
	return y
}

func (s Source) near(i Index) string {
	margin := 3
	start := max(0, i-margin)
	end := min(len(s.tokens)-1, i+margin)
	tokens := s.tokens[start:end]
	ss := strings.Builder{}
	for _, t := range tokens {
		if t.tag == TokenTerminator {
			ss.WriteByte(';')
			continue
		}
		ss.WriteString(s.lexeme(t) + " ")
	}
	return ss.String()
}
