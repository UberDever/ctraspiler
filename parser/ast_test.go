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

func TestParseConstDecl(t *testing.T) {
	lhs := `
		fn main() {
			const _ = null
			const a, b = 8, 2
			const c, d, e = 8 * 3, -16, "E"
		}
	`
	rhs := `
	(Source
		(FunctionDecl main 
			(Signature ())
				(Block 
					(ConstDecl 
						(_) 
						(ExpressionList (null)))
					(ConstDecl 
						(a b) 
						(ExpressionList (8) (2)))
					(ConstDecl 
						(c d e) 
						(ExpressionList 
							(* (8) (3))
							(- (16))
							("E")
						))
	)))`
	result, expected := testAST(lhs, rhs)
	if result != expected {
		t.Errorf("AST are not equal\n%s\n\n%s", formatSExpr(result), formatSExpr(expected))
	}
}

func TestParseSomeExpressions(t *testing.T) {
	lhs := `
		fn main() {
			const x = 8
			x * 8 + 3
			x + 3 / 4
			f(x, x.y)
			x, x.y = x.y, x
		}
	`
	rhs := `
	(Source
		(FunctionDecl main 
			(Signature ())
				(Block 
					(ConstDecl (x) (ExpressionList (8)))
					(+ (* (x) (8)) (3))
					(+ (x) (/ (3) (4)))
					(Call (f) 
						(ExpressionList
							(x)
							(Selector (x) (y))))
					(Assign
						(ExpressionList (x) (Selector (x) (y)))
						(ExpressionList (Selector (x) (y)) (x))
					)
	)))`
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
