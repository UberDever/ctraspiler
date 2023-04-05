package parser

//go:generate ./generate.sh

import (
	"github.com/antlr/antlr4/runtime/Go/antlr/v4"
)

type SomeParserBase struct {
	*antlr.BaseParser
}
