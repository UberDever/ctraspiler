package parser

import (
	"testing"

	"golang.org/x/exp/utf8string"
)

func testAST(lhs string, rhs string) (result string, expected string) {
	source := utf8string.NewString(lhs)
	expected = unformatSExpr(utf8string.NewString(rhs).String())
	bytes := []byte(source.String())
	src := tokenize(bytes)
	ast := Parse(&src)
	result = ast.dump(false)
	return
}

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
	src := tokenize([]byte(source.String()))
	tokens := src.tokens
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
	lhs := `
		fn main()
		fn some(a, b) // some function
	`
	rhs := `
		(Source
			(FunctionDecl main 
				(Signature ()))
			(FunctionDecl some 
				(Signature (a b)
				)))
	`
	result, expected := testAST(lhs, rhs)
	if result != expected {
		t.Errorf("AST are not equal\n%s\n\n%s", formatSExpr(result), formatSExpr(expected))
	}
}

func TestParseFunctionWithBody(t *testing.T) {
	lhs := `
		fn main() {
			const a, b = 8, 2
		}
	`
	rhs := `
	(Source
		(FunctionDecl main 
			(Signature ())
				(Block 
					(ConstDecl 
						(a b) 
						(ExpressionList (8) (2))
	))))`
	result, expected := testAST(lhs, rhs)
	if result != expected {
		t.Errorf("AST are not equal\n%s\n\n%s", formatSExpr(result), formatSExpr(expected))
	}
}

func TestSExprFormatting(t *testing.T) {
	source := utf8string.NewString(`
		fn main()
		fn some(a, b) // some function
	`)
	bytes := []byte(source.String())
	src := tokenize(bytes)
	ast := Parse(&src)
	dump := ast.dump(false)
	formatted := formatSExpr(dump)
	unformatted := unformatSExpr(formatted)
	if dump != unformatted {
		t.Errorf("SExpr are not equal {%#v} {%#v}", dump, unformatted)
	}
}
