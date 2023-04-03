package parser

//go:generate ./generate.sh

import (
	"github.com/antlr/antlr4/runtime/Go/antlr/v4"
)

type SomeParserBase struct {
	*antlr.BaseParser
}

// Returns true if the current Token is a closing bracket (")" or "}")
func (p *SomeParserBase) ClosingBracket() bool {
	stream := p.GetTokenStream()
	prevTokenType := stream.LA(1)
	return prevTokenType == SomeParserR_PAREN || prevTokenType == SomeParserR_CURLY
}
