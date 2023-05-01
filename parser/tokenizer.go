package parser

import (
	"math"
	antlr_parser "some/antlr"

	antlr "github.com/antlr/antlr4/runtime/Go/antlr/v4"
	"golang.org/x/exp/utf8string"
)

type TokenTag int
type TokenIndex int

const (
	TokenIndexInvalid = math.MinInt
)

// NOTE: This was a bad idea - full mapping of tokens is better solution
// i.e. increase granularity
const (
	TokenEOF          TokenTag = -1
	TokenKeyword               = antlr_parser.SomeKEYWORD
	TokenIdentifier            = antlr_parser.SomeIDENTIFIER
	TokenPunctuation           = antlr_parser.SomeOTHER_OP
	TokenUnaryOp               = antlr_parser.SomeUNARY_OP
	TokenBinaryOp              = antlr_parser.SomeBINARY_OP
	TokenIntLit                = antlr_parser.SomeINT_LIT
	TokenFloatLit              = antlr_parser.SomeFLOAT_LIT
	TokenImaginaryLit          = antlr_parser.SomeIMAGINARY_LIT
	TokenRuneLit               = antlr_parser.SomeRUNE_LIT
	TokenLittleUValue          = antlr_parser.SomeLITTLE_U_VALUE
	TokenBigUValue             = antlr_parser.SomeBIG_U_VALUE
	TokenStringLit             = antlr_parser.SomeSTRING_LIT
	TokenWS                    = antlr_parser.SomeWS
	TokenTerminator            = antlr_parser.SomeTERMINATOR
	TokenLineComment           = antlr_parser.SomeLINE_COMMENT
)

type Token struct {
	tag   TokenTag
	start int
	end   int
	line  int
	col   int
}

var EOF = Token{
	tag:   TokenEOF,
	start: -1,
	end:   -1,
	line:  -1,
	col:   -1,
}

func tryInsertSemicolon(src Source, terminator antlr.Token, tokens []Token) []Token {
	semicolon := Token{
		tag:   TokenTerminator,
		start: terminator.GetStart(),
		end:   terminator.GetStop(),
		line:  terminator.GetLine(),
		col:   terminator.GetColumn(),
	}

	if len(tokens) > 0 {
		i := len(tokens) - 1
		last := src.token(TokenIndex(i))

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
			lexeme := src.lexeme(src.token(TokenIndex(i)))
			if lexeme == "break" ||
				lexeme == "continue" ||
				lexeme == "fallthrough" ||
				lexeme == "return" {
				tokens = append(tokens, semicolon)

			}
		case TokenPunctuation:
			lexeme := src.lexeme(src.token(TokenIndex(i)))
			if lexeme == "++" ||
				lexeme == "--" ||
				lexeme == ")" ||
				lexeme == "]" ||
				lexeme == "}" {
				tokens = append(tokens, semicolon)
			}
		}
	}

	return tokens
}

func tokenize(source []byte) Source {
	src := Source{text: *utf8string.NewString(string(source))}

	is := antlr.NewInputStream(string(source))
	lexer := antlr_parser.NewSome(is)

	antlrTokens := lexer.GetAllTokens()
	src.tokens = make([]Token, 0, len(antlrTokens))
	for i := range antlrTokens {
		t := antlrTokens[i]
		if t.GetChannel() == antlr.TokenHiddenChannel {
			continue
		}
		if t.GetTokenType() == antlr_parser.SomeTERMINATOR ||
			t.GetTokenType() == antlr_parser.SomeLINE_COMMENT {
			src.tokens = tryInsertSemicolon(src, t, src.tokens)
			continue
		}
		src.tokens = append(src.tokens, Token{
			tag:   TokenTag(t.GetTokenType()),
			start: t.GetStart(),
			end:   t.GetStop(),
			line:  t.GetLine(),
			col:   t.GetColumn(),
		})
	}
	src.tokens = append(src.tokens, EOF)

	return src
}
