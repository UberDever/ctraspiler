package main

import (
	"testing"
)

func TestParser(t *testing.T) {
	test := `
	;
	/* fn main() {
		2 + 3
		4 * 10 
	}

	fn someOtherFunc
	(a, b, c) {
		- 10
	} */
	`
	ast := Parse(&DefaultConfig{}, []byte(test))
	_ = ast
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
