package main

import (
	"testing"

	"golang.org/x/exp/utf8string"
)

var example1 = `
	type Some struct {
		result f32
	}

	fn func1(a i8, b f32, c f64) i64 {
		return a * b + c
	}

	fn func2(a Some, b f32, c f64) {
		a.result = b + c
	}

	s := Some{}
	s.func2(5.3, 4.8)
`

func TestTokenizer(t *testing.T) {
	// source := "fn main() {}"
	source := utf8string.NewString(`fn identifier
	break
	&& == + - * / 
	!
	129389512754912957199521
	3.63252e-24
	"some string"
	Идентификатор
	`)
	tokens := tokenize([]byte(source.String()))
	expected := []struct {
		int
		string
	}{
		{TokenKeyword, "fn"},
		{TokenIdentifier, "identifier"},
		{TokenTerminator, "\n"},
		{TokenKeyword, "break"},
		{TokenTerminator, "\n"},
		{TokenBinaryOp, "&&"},
		{TokenBinaryOp, "=="},
		{TokenBinaryOp, "+"},
		{TokenBinaryOp, "-"},
		{TokenBinaryOp, "*"},
		{TokenBinaryOp, "/"},
		{TokenTerminator, "\n"},
		{TokenUnaryOp, "!"},
		{TokenTerminator, "\n"},
		{TokenIntLit, "129389512754912957199521"},
		{TokenTerminator, "\n"},
		{TokenFloatLit, "3.63252e-24"},
		{TokenTerminator, "\n"},
		{TokenStringLit, "\"some string\""},
		{TokenTerminator, "\n"},
		{TokenIdentifier, "Идентификатор"},
		{TokenTerminator, "\n"},
	}
	if len(tokens) != len(expected) {
		t.Errorf("Same tokens arrays expected, got tokens=%d and expected=%d", len(tokens), len(expected))
	}
	tokensLen := len(tokens)
	for i := 0; i < tokensLen; i++ {
		lhs := tokens[i]
		rhs := expected[i]
		asStr := source.Slice(int(lhs.start), int(lhs.end)+1)
		if asStr != rhs.string {
			t.Errorf("[%d] Strings %s != %s", i, asStr, rhs.string)
		}
		if lhs.tag != rhs.int {
			t.Errorf("[%d] Types %d != %d", i, lhs.tag, rhs.int)
		}
	}
}

// func matchTokens(str string, tokens []Token) error {
// 	config, err := NewDefaultConfig("config.json")
// 	if err != nil {
// 		return err
// 	}

// 	p := NewParser(&config)
// 	p.Tokenize([]byte(str))

// 	for i := range tokens {
// 		got := p.NextToken()
// 		want := tokens[i]

// 		if got.isEOF() {
// 			return fmt.Errorf("Unexpected EOF at %v", want)
// 		}

// 		if got.key != want.key ||
// 			got.value != want.value ||
// 			got.line != want.line {
// 			return fmt.Errorf("%v =/> %v", got, want)
// 		}

// 	}

// 	return nil
// }

// func TestTokenizerIdentifiers(t *testing.T) {
// 	{
// 		got := "one two three"
// 		expected := []Token{
// 			{key: TokenType(TokenIdentifier), value: "one", line: 1},
// 			{key: TokenType(TokenIdentifier), value: "two", line: 1},
// 			{key: TokenType(TokenIdentifier), value: "three", line: 1},
// 		}

// 		err := matchTokens(got, expected)
// 		if err != nil {
// 			t.Errorf(err.Error())
// 		}
// 	}

// 	{
// 		got := ` one
// 			two
// 			three
// 		`
// 		expected := []Token{
// 			{key: TokenType(TokenIdentifier), value: "one", line: 1},
// 			{key: TokenType(TokenIdentifier), value: "two", line: 2},
// 			{key: TokenType(TokenIdentifier), value: "three", line: 3},
// 		}

// 		err := matchTokens(got, expected)
// 		if err != nil {
// 			t.Errorf(err.Error())
// 		}
// 	}

// 	{
// 		got := ` var
// 			type
// 			struct
// 			return
// 			if
// 			func
// 			for
// 			else
// 			break
// 			defer
// 			match
// 		`
// 		expected := []Token{
// 			{key: TokenType(TokenKeyword), value: "var", line: 1},
// 			{key: TokenType(TokenKeyword), value: "type", line: 2},
// 			{key: TokenType(TokenKeyword), value: "struct", line: 3},
// 			{key: TokenType(TokenKeyword), value: "return", line: 4},
// 			{key: TokenType(TokenKeyword), value: "if", line: 5},
// 			{key: TokenType(TokenKeyword), value: "func", line: 6},
// 			{key: TokenType(TokenKeyword), value: "for", line: 7},
// 			{key: TokenType(TokenKeyword), value: "else", line: 8},
// 			{key: TokenType(TokenKeyword), value: "break", line: 9},
// 			{key: TokenType(TokenKeyword), value: "defer", line: 10},
// 			{key: TokenType(TokenIdentifier), value: "match", line: 11},
// 		}

// 		err := matchTokens(got, expected)
// 		if err != nil {
// 			t.Errorf(err.Error())
// 		}
// 	}
// }

// func TestTokenizerOperators(t *testing.T) {
// 	{
// 		got := "2 + 2"
// 		expected := []Token{
// 			{key: TokenType(TokenInteger), value: "2", line: 1},
// 			{key: TokenType(TokenOperator), value: "+", line: 1},
// 			{key: TokenType(TokenInteger), value: "2", line: 1},
// 		}

// 		err := matchTokens(got, expected)
// 		if err != nil {
// 			t.Errorf(err.Error())
// 		}
// 	}

// 	{
// 		got := ` []{} + :=--
// 			++
// 		`
// 		expected := []Token{
// 			{key: TokenType(TokenOperator), value: "[]", line: 1},
// 			{key: TokenType(TokenOperator), value: "{}", line: 1},
// 			{key: TokenType(TokenOperator), value: "+", line: 1},
// 			{key: TokenType(TokenOperator), value: ":=", line: 1},
// 			{key: TokenType(TokenOperator), value: "--", line: 1},
// 			{key: TokenType(TokenOperator), value: "++", line: 2},
// 		}

// 		err := matchTokens(got, expected)
// 		if err != nil {
// 			t.Errorf(err.Error())
// 		}
// 	}
// }

// func TestTokenizerComments(t *testing.T) {
// 	{
// 		got := `
// 			// some comment
// 			non_comment
// 			//this is very//nasty comment
// 		`
// 		expected := []Token{
// 			{key: TokenType(TokenComment), value: " some comment", line: 2},
// 			{key: TokenType(TokenIdentifier), value: "non_comment", line: 3},
// 			{key: TokenType(TokenComment), value: "this is very//nasty comment", line: 4},
// 		}

// 		err := matchTokens(got, expected)
// 		if err != nil {
// 			t.Errorf(err.Error())
// 		}
// 	}
// }
