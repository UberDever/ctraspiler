package parser

import (
	antlr_parser "some/antlr"

	antlr "github.com/antlr/antlr4/runtime/Go/antlr/v4"
	"golang.org/x/exp/utf8string"
)

const (
	TokenEOF          = -1
	TokenUndefined    = 0
	TokenKeyword      = antlr_parser.SomeKEYWORD
	TokenIdentifier   = antlr_parser.SomeIDENTIFIER
	TokenBinaryOp     = antlr_parser.SomeBINARY_OP
	TokenUnaryOp      = antlr_parser.SomeUNARY_OP
	TokenPunctuation  = antlr_parser.SomeOTHER_OP
	TokenIntLit       = antlr_parser.SomeINT_LIT
	TokenFloatLit     = antlr_parser.SomeFLOAT_LIT
	TokenImaginaryLit = antlr_parser.SomeIMAGINARY_LIT
	TokenRuneLit      = antlr_parser.SomeRUNE_LIT
	TokenLittleUValue = antlr_parser.SomeLITTLE_U_VALUE
	TokenBigUValue    = antlr_parser.SomeBIG_U_VALUE
	TokenStringLit    = antlr_parser.SomeSTRING_LIT
	TokenWS           = antlr_parser.SomeWS
	TokenTerminator   = antlr_parser.SomeTERMINATOR
	TokenLineComment  = antlr_parser.SomeLINE_COMMENT
)

type Token struct {
	tag   Tag
	start uint
	end   uint
	line  int
	col   int
}

func tryInsertSemicolon(source []byte, terminator antlr.Token, tokens []Token) []Token {
	text := utf8string.NewString(string(source))
	semicolon := Token{
		tag:   TokenTerminator,
		start: uint(terminator.GetStart()),
		end:   uint(terminator.GetStop()),
		line:  terminator.GetLine(),
		col:   terminator.GetColumn(),
	}

	if len(tokens) > 0 {
		last := tokens[len(tokens)-1]
		lexeme := text.Slice(int(last.start), int(last.end+1))

		switch last.tag {
		case TokenIdentifier:
			fallthrough
		case TokenIntLit:
			fallthrough
		case TokenFloatLit:
			fallthrough
		case TokenImaginaryLit:
			fallthrough
		case TokenRuneLit:
			fallthrough
		case TokenStringLit:
			tokens = append(tokens, semicolon)

		case TokenKeyword:
			if lexeme == "break" ||
				lexeme == "continue" ||
				lexeme == "fallthrough" ||
				lexeme == "return" {
				tokens = append(tokens, semicolon)

			}
		case TokenPunctuation:
			if lexeme == "++" ||
				lexeme == "--" ||
				lexeme == ")" ||
				lexeme == "]" ||
				lexeme == "}" {
				tokens = append(tokens, semicolon)
			}
		}
	} else {
		tokens = append(tokens, semicolon)
	}

	return tokens
}

func tokenize(source []byte) []Token {
	is := antlr.NewInputStream(string(source))
	lexer := antlr_parser.NewSome(is)

	antlrTokens := lexer.GetAllTokens()
	tokens := make([]Token, 0, len(antlrTokens))
	for i := range antlrTokens {
		t := antlrTokens[i]
		if t.GetChannel() == antlr.TokenHiddenChannel {
			continue
		}
		if t.GetTokenType() == antlr_parser.SomeTERMINATOR {
			tokens = tryInsertSemicolon(source, t, tokens)
			continue
		}
		tokens = append(tokens, Token{
			tag:   t.GetTokenType(),
			start: uint(t.GetStart()),
			end:   uint(t.GetStop()),
			line:  t.GetLine(),
			col:   t.GetColumn(),
		})
	}
	tokens = append(tokens, Token{tag: TokenEOF})

	return tokens
}
