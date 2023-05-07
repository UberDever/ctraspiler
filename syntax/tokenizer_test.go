package syntax

import (
	"some/util"
	"strings"
	"testing"

	"golang.org/x/exp/utf8string"
)

func TestTokenizer(t *testing.T) {
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
		tokenTag
		string
	}{
		{TokenKeyword, "fn"},
		{TokenIdentifier, "identifier"},
		{TokenPunctuation, "("},
		{TokenPunctuation, ")"},
		{TokenTerminator, "\n"},
		{TokenKeyword, "break"},
		{TokenTerminator, "\n"},
		{TokenBinaryOp, "&&"},
		{TokenBinaryOp, "=="},
		{TokenBinaryOp, "+"},
		{TokenBinaryOp, "-"},
		{TokenUnaryOp, "*"},
		{TokenBinaryOp, "/"},
		{TokenUnaryOp, "!"},
		{TokenIntLit, "129389512754912957199521"},
		{TokenTerminator, "\n"},
		{TokenFloatLit, "3.63252e-24"},
		{TokenTerminator, "\n"},
		{TokenStringLit, "\"some string\""},
		{TokenTerminator, "\n"},
		{TokenIdentifier, "Идентификатор"},
		{TokenTerminator, "\n"},
	}

	if !handler.Empty() {
		errs := handler.AllErrors()
		t.Error(strings.Join(errs, " "))
	}

	if tokens[len(tokens)-1].tag != TokenEOF {
		t.Errorf("Missed EOF at the end of token stream")
	}
	tokens = tokens[:len(tokens)-1]

	// for i := range tokens {
	// 	t := tokens[i]
	// 	if t.tag == TokenTerminator {
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
		asStr := text.Slice(int(lhs.start), int(lhs.end)+1)
		if asStr != rhs.string {
			t.Errorf("[%d] Strings %s != %s", i, asStr, rhs.string)
		}
		if lhs.tag != rhs.tokenTag {
			t.Errorf("[%d] Types %d != %d", i, lhs.tag, rhs.tokenTag)
		}
	}
}
