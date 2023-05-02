package parser

import (
	"fmt"
	"testing"

	"golang.org/x/exp/utf8string"
)

func isASTValid(nodes []Node) (Node, int, bool) {
	for i, n := range nodes {
		if n.tokenIdx == TokenIndexInvalid ||
			n.lhs == NodeIndexInvalid ||
			n.rhs == NodeIndexInvalid {
			return n, i, false
		}
	}
	return Node{}, -1, true
}

func runTest(lhs string, rhs string) error {
	source := utf8string.NewString(lhs)
	bytes := []byte(source.String())
	src := tokenize(bytes)
	ast := Parse(&src)
	expected := unformatSExpr(utf8string.NewString(rhs).String())
	result := ast.dump(false)

	if node, index, ok := isASTValid(ast.nodes); !ok {
		return fmt.Errorf("AST nodes failed on validity test at %d => %v", index, node)
	}
	if result != expected {
		return fmt.Errorf("AST are not equal\n%s\n\n%s", formatSExpr(result), formatSExpr(expected))
	}
	return nil
}

func TestParseFunctionDecl(t *testing.T) {
	lhs := `
		fn main()
		fn some(a, b) // some function
	`
	rhs := `
		(Source
			(FunctionDecl (main)
				(Signature ()))
			(FunctionDecl (some)
				(Signature (a b)
				)))
	`
	if e := runTest(lhs, rhs); e != nil {
		t.Error(e)
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
		(FunctionDecl (main) 
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
	if e := runTest(lhs, rhs); e != nil {
		t.Error(e)
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
		(FunctionDecl (main)
			(Signature ())
				(Block 
					(ConstDecl (x) (ExpressionList (8)))
					(+ (* (x) (8)) (3))
					(+ (x) (/ (3) (4)))
					(Call (f) 
						(ExpressionList (x) (Get (x) (y))))
					(Assign
						(ExpressionList (x) (Get (x) (y)))
						(ExpressionList (Get (x) (y)) (x)))
	)))`
	if e := runTest(lhs, rhs); e != nil {
		t.Error(e)
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
