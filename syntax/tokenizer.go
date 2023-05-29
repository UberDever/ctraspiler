package syntax

import (
	ID "some/domain"
	"some/util"

	antlr_parser "some/antlr"

	antlr "github.com/antlr/antlr4/runtime/Go/antlr/v4"
)

type token struct {
	Tag   ID.Token
	Start int
	End   int
	Line  int
	Col   int
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
			Tag:   ID.Token(t.GetTokenType()),
			Start: t.GetStart(),
			End:   t.GetStop(),
			Line:  t.GetLine(),
			Col:   t.GetColumn(),
		})
	}
	src.tokens = append(src.tokens, token{ID.TokenEOF, -1, -1, -1, -1})
}

func (tok *tokenizer) tryInsertSemicolon(s *Source, terminator antlr.Token) []token {
	semicolon := token{
		Tag:   ID.TokenTerminator,
		Start: terminator.GetStart(),
		End:   terminator.GetStop(),
		Line:  terminator.GetLine(),
		Col:   terminator.GetColumn(),
	}

	if len(s.tokens) > 0 {
		i := len(s.tokens) - 1
		last := s.Token(ID.Token(i))

		switch last.Tag {
		case ID.TokenIdentifier:
			fallthrough
		case ID.TokenIntLit:
			fallthrough
		case ID.TokenFloatLit:
			fallthrough
		case ID.TokenImaginaryLit:
			fallthrough
		case ID.TokenRuneLit:
			fallthrough
		case ID.TokenStringLit:
			fallthrough
		case ID.TokenBoolLit:
			s.tokens = append(s.tokens, semicolon)

		case ID.TokenKeyword:
			lexeme := s.Lexeme(ID.Token(i))
			if lexeme == "break" ||
				lexeme == "continue" ||
				lexeme == "fallthrough" ||
				lexeme == "return" {
				s.tokens = append(s.tokens, semicolon)

			}
		case ID.TokenPunctuation:
			lexeme := s.Lexeme(ID.Token(i))
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
