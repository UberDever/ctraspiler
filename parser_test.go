package main

import (
	"fmt"
	"testing"
)

func matchTokens(str string, tokens []Token) error {
	config, err := NewDefaultConfig("config.json")
	if err != nil {
		return err
	}

	p := NewParser(&config)
	p.Tokenize([]byte(str))

	for i := range tokens {
		got := p.NextToken()
		want := tokens[i]

		if got.isEOF() {
			return fmt.Errorf("Unexpected EOF at %v", want)
		}

		if got.key != want.key {
			return fmt.Errorf("Key %d =/> %d", got.key, want.key)
		}
		if got.value != want.value {
			return fmt.Errorf("Value %s =/> %s", got.value, want.value)
		}

	}

	return nil
}

func TestTokenizerIdentifiers(t *testing.T) {
	{
		got := "one two three"
		expected := []Token{
			{key: int(TokenIdentifier), value: "one", line: 0},
			{key: int(TokenIdentifier), value: "two", line: 0},
			{key: int(TokenIdentifier), value: "three", line: 0},
		}

		err := matchTokens(got, expected)
		if err != nil {
			t.Errorf(err.Error())
		}
	}

	{
		got := `
			one
			two
			three
		`
		expected := []Token{
			{key: int(TokenIdentifier), value: "one", line: 1},
			{key: int(TokenIdentifier), value: "two", line: 2},
			{key: int(TokenIdentifier), value: "three", line: 3},
		}

		err := matchTokens(got, expected)
		if err != nil {
			t.Errorf(err.Error())
		}
	}

	{
		got := `
			var
			type
			struct
			return
			if
			func
			for
			else
			break
			defer
			match
		`
		expected := []Token{
			{key: int(TokenKeyword), value: "var", line: 1},
			{key: int(TokenKeyword), value: "type", line: 2},
			{key: int(TokenKeyword), value: "struct", line: 3},
			{key: int(TokenKeyword), value: "return", line: 4},
			{key: int(TokenKeyword), value: "if", line: 5},
			{key: int(TokenKeyword), value: "func", line: 6},
			{key: int(TokenKeyword), value: "for", line: 7},
			{key: int(TokenKeyword), value: "else", line: 8},
			{key: int(TokenKeyword), value: "break", line: 9},
			{key: int(TokenKeyword), value: "defer", line: 10},
			{key: int(TokenIdentifier), value: "match", line: 11},
		}

		err := matchTokens(got, expected)
		if err != nil {
			t.Errorf(err.Error())
		}
	}
}

func TestTokenizerOperators(t *testing.T) {
	{
		got := "2 + 2"
		expected := []Token{
			{key: int(TokenInteger), value: "2", line: 0},
			{key: int(TokenOperator), value: "+", line: 0},
			{key: int(TokenInteger), value: "2", line: 0},
		}

		err := matchTokens(got, expected)
		if err != nil {
			t.Errorf(err.Error())
		}
	}
}
