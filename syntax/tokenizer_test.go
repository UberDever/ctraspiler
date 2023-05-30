package syntax

import (
	ID "some/domain"
	"some/util"
	"strings"
	"testing"

	"golang.org/x/exp/utf8string"
)

func TestTokenizerTokens(t *testing.T) {
	text := utf8string.NewString(`fn identifier()
	break
	&& == + - * / 
	!
	129389512754912957199521
	3.63252e-24
	"some string"
	Идентификатор
	`)

	handler := util.NewHandler()
	src := NewSource("tokenizer_test", *text)
	tokenizer := NewTokenizer(&handler)
	tokenizer.Tokenize(&src)

	tokens := src.tokens
	expected := [...]struct {
		ID.Token
		string
	}{
		{ID.TokenKeyword, "fn"},
		{ID.TokenIdentifier, "identifier"},
		{ID.TokenPunctuation, "("},
		{ID.TokenPunctuation, ")"},
		{ID.TokenTerminator, "\n"},
		{ID.TokenKeyword, "break"},
		{ID.TokenTerminator, "\n"},
		{ID.TokenBinaryOp, "&&"},
		{ID.TokenBinaryOp, "=="},
		{ID.TokenBinaryOp, "+"},
		{ID.TokenBinaryOp, "-"},
		{ID.TokenUnaryOp, "*"},
		{ID.TokenBinaryOp, "/"},
		{ID.TokenUnaryOp, "!"},
		{ID.TokenIntLit, "129389512754912957199521"},
		{ID.TokenTerminator, "\n"},
		{ID.TokenFloatLit, "3.63252e-24"},
		{ID.TokenTerminator, "\n"},
		{ID.TokenStringLit, "\"some string\""},
		{ID.TokenTerminator, "\n"},
		{ID.TokenIdentifier, "Идентификатор"},
		{ID.TokenTerminator, "\n"},
	}

	if !handler.IsEmpty() {
		errs := handler.AllErrors()
		t.Error(strings.Join(errs, " "))
	}

	if tokens[len(tokens)-1].Tag != ID.TokenEOF {
		t.Errorf("Missed EOF at the end of token stream")
	}
	tokens = tokens[:len(tokens)-1]

	// for i := range tokens {
	// 	t := tokens[i]
	// 	if t.tag == ID.TokenTerminator {
	// 		fmt.Print(";")
	// 	} else {
	// 		fmt.Print(source.Slice(int(t.start), int(t.end)+1))
	// 	}
	// 	fmt.Print(" ")
	// }

	if len(tokens) != len(expected) {
		t.Errorf("Same tokens arrays expected, got tokens=%d and expected=%d", len(tokens), len(expected))
	}

	tokensLen := len(tokens)
	for i := 0; i < tokensLen; i++ {
		lhs := tokens[i]
		rhs := expected[i]
		asStr := text.Slice(int(lhs.Start), int(lhs.End)+1)
		if asStr != rhs.string {
			t.Errorf("[%d] Strings %s != %s", i, asStr, rhs.string)
		}
		if lhs.Tag != rhs.Token {
			t.Errorf("[%d] Types %d != %d", i, lhs.Tag, rhs.Token)
		}
	}
}

func TestTokenizerProgram(t *testing.T) {
	text := utf8string.NewString(`
	fn main() {

	}

	fn some(a, b) {

	}
	`)

	handler := util.NewHandler()
	src := NewSource("tokenizer_test", *text)
	tokenizer := NewTokenizer(&handler)
	tokenizer.Tokenize(&src)

	tokens := src.tokens
	expected := [...]struct {
		ID.Token
		string
	}{
		{ID.TokenKeyword, "fn"},
		{ID.TokenIdentifier, "main"},
		{ID.TokenPunctuation, "("},
		{ID.TokenPunctuation, ")"},
		{ID.TokenPunctuation, "{"},
		{ID.TokenPunctuation, "}"},
		{ID.TokenTerminator, "\n"},
		{ID.TokenKeyword, "fn"},
		{ID.TokenIdentifier, "some"},
		{ID.TokenPunctuation, "("},
		{ID.TokenIdentifier, "a"},
		{ID.TokenPunctuation, ","},
		{ID.TokenIdentifier, "b"},
		{ID.TokenPunctuation, ")"},
		{ID.TokenPunctuation, "{"},
		{ID.TokenPunctuation, "}"},
		{ID.TokenTerminator, "\n"},
	}

	if !handler.IsEmpty() {
		errs := handler.AllErrors()
		t.Error(strings.Join(errs, " "))
	}

	if tokens[len(tokens)-1].Tag != ID.TokenEOF {
		t.Errorf("Missed EOF at the end of token stream")
	}
	tokens = tokens[:len(tokens)-1]

	if len(tokens) != len(expected) {
		t.Errorf("Same tokens arrays expected, got tokens=%d and expected=%d", len(tokens), len(expected))
	}

	tokensLen := len(tokens)
	for i := 0; i < tokensLen; i++ {
		lhs := tokens[i]
		rhs := expected[i]
		asStr := text.Slice(int(lhs.Start), int(lhs.End)+1)
		if asStr != rhs.string {
			t.Errorf("[%d] Strings %s != %s", i, asStr, rhs.string)
		}
		if lhs.Tag != rhs.Token {
			t.Errorf("[%d] Types %d != %d", i, lhs.Tag, rhs.Token)
		}
	}
}
