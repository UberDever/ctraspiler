package syntax

import (
	"math"
	antlr_parser "some/antlr"
	"some/util"

	antlr "github.com/antlr/antlr4/runtime/Go/antlr/v4"
)

type TokenTag int
type TokenIndex int

const (
	tokenIndexInvalid TokenIndex = math.MinInt
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

type token struct {
	tag   TokenTag
	start int
	end   int
	line  int
	col   int
}

type tokenizer struct {
	handler *util.ErrorHandler
}

func NewTokenizer(handler *util.ErrorHandler) tokenizer {
	return tokenizer{handler}
}

func (tok *tokenizer) Tokenize(src *Source) {
	is := antlr.NewInputStream(src.text.String())
	lexer := antlr_parser.NewSome(is)

	antlrTokens := lexer.GetAllTokens()
	src.tokens = make([]token, 0, len(antlrTokens))
	for i := range antlrTokens {
		t := antlrTokens[i]
		if t.GetChannel() == antlr.TokenHiddenChannel {
			continue
		}
		if t.GetTokenType() == antlr_parser.SomeTERMINATOR ||
			t.GetTokenType() == antlr_parser.SomeLINE_COMMENT {
			src.tokens = tok.tryInsertSemicolon(src, t)
			continue
		}
		src.tokens = append(src.tokens, token{
			tag:   TokenTag(t.GetTokenType()),
			start: t.GetStart(),
			end:   t.GetStop(),
			line:  t.GetLine(),
			col:   t.GetColumn(),
		})
	}
	src.tokens = append(src.tokens, token{TokenEOF, -1, -1, -1, -1})
}

func (tok *tokenizer) tryInsertSemicolon(s *Source, terminator antlr.Token) []token {
	semicolon := token{
		tag:   TokenTerminator,
		start: terminator.GetStart(),
		end:   terminator.GetStop(),
		line:  terminator.GetLine(),
		col:   terminator.GetColumn(),
	}

	if len(s.tokens) > 0 {
		i := len(s.tokens) - 1
		last := s.token(TokenIndex(i))

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
			s.tokens = append(s.tokens, semicolon)

		case TokenKeyword:
			lexeme := s.Lexeme(TokenIndex(i))
			if lexeme == "break" ||
				lexeme == "continue" ||
				lexeme == "fallthrough" ||
				lexeme == "return" {
				s.tokens = append(s.tokens, semicolon)

			}
		case TokenPunctuation:
			lexeme := s.Lexeme(TokenIndex(i))
			if lexeme == "++" ||
				lexeme == "--" ||
				lexeme == ")" ||
				lexeme == "]" ||
				lexeme == "}" {
				s.tokens = append(s.tokens, semicolon)
			}
		}
	}

	return s.tokens
}
