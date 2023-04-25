package parser

import (
	"fmt"
	"testing"

	"golang.org/x/exp/utf8string"
)

// type Some struct {
// 	result f32
// }

// fn func1(a i8, b f32, c f64) i64 {
// 	return a * b + c
// }

// fn func2(a Some, b f32, c f64) {
// 	a.result = b + c
// }

// s := Some{}
// s.func2(5.3, 4.8)

func TestTokenizer(t *testing.T) {
	source := utf8string.NewString(`fn identifier()
	break
	&& == + - * / 
	!
	129389512754912957199521
	3.63252e-24
	"some string"
	Идентификатор
	`)
	tokens := tokenize([]byte(source.String()))
	expected := [...]struct {
		int
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
		{TokenBinaryOp, "*"},
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
		asStr := source.Slice(int(lhs.start), int(lhs.end)+1)
		if asStr != rhs.string {
			t.Errorf("[%d] Strings %s != %s", i, asStr, rhs.string)
		}
		if lhs.tag != rhs.int {
			t.Errorf("[%d] Types %d != %d", i, lhs.tag, rhs.int)
		}
	}
}

func TestParseFunctionDecl(t *testing.T) {
	source := utf8string.NewString(`
		fn main()
		fn some(a, b)
	`)
	bytes := []byte(source.String())
	tokens := tokenize(bytes)
	ast := Parse(bytes, tokens)
	ast.Traverse(func(node *Node) {
		fmt.Println(ast.GetNodeString(node))
	})
}

// func TestParseLiterals(t *testing.T) {
// 	source := utf8string.NewString(`
// 		5.6
// 		 14.8
// 		  1235419
// 		   "some"
// 		"string"
// 	`)
// 	bytes := []byte(source.String())
// 	tokens := tokenize(bytes)
// 	ast := Parse(bytes, tokens)
// 	ast.Traverse(func(node *Node) {
// 		fmt.Println(ast.GetNodeString(node))
// 	})
// }
