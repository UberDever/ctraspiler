package syntax

import (
	"errors"
	"fmt"
	"some/util"
	"strings"
	"testing"

	"golang.org/x/exp/utf8string"
)

func isASTValid(nodes []Node) (Node, int, bool) {
	for i, n := range nodes {
		if n.tokenIdx == tokenIndexInvalid ||
			n.lhs == NodeIndexInvalid ||
			n.rhs == NodeIndexInvalid {
			return n, i, false
		}
	}
	return Node{}, -1, true
}

func runTest(lhs string, rhs string) error {
	text := utf8string.NewString(lhs)
	src := NewSource("ast_test", *text)

	handler := util.NewHandler()
	tokenizer := NewTokenizer(&handler)
	tokenizer.Tokenize(&src)
	if !handler.Empty() {
		errs := handler.AllErrors()
		return errors.New(strings.Join(errs, ""))
	}

	parser := NewParser(&handler)
	ast := parser.Parse(&src)
	if !handler.Empty() {
		errs := handler.AllErrors()
		return errors.New(strings.Join(errs, ""))
	}

	expected := unformatSExpr(utf8string.NewString(rhs).String())
	result := ast.Dump(false)

	if node, index, ok := isASTValid(ast.nodes); !ok {
		return fmt.Errorf("AST nodes failed on validity test at %d => %v", index, node)
	}
	if result != expected {
		return fmt.Errorf("AST are not equal\n%s\n\n%s", formatSExpr(result), formatSExpr(expected))
	}
	return nil
}

func TestErrorHandling(t *testing.T) {
	lhs := `
		fn main()
		some(a, b) // some function
	`
	rhs := ``
	e := runTest(lhs, rhs)
	if e == nil {
		t.Error("Expected error")
	}
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

func TestReturnStmt(t *testing.T) {
	lhs := `
		fn main() {
			return 0
		}
	`
	rhs := `
	(Source
		(FunctionDecl (main)
			(Signature ())
			(Block 
				(Return (ExpressionList (0)
	)))))`
	if e := runTest(lhs, rhs); e != nil {
		t.Error(e)
	}
}

func TestSExprFormatting(t *testing.T) {
	text := utf8string.NewString(`
		fn main()
		fn some(a, b) // some function
	`)
	src := NewSource("ast_test", *text)

	handler := util.NewHandler()
	tokenizer := NewTokenizer(&handler)
	tokenizer.Tokenize(&src)
	if !handler.Empty() {
		errs := handler.AllErrors()
		t.Error(strings.Join(errs, " "))
	}

	parser := NewParser(&handler)
	ast := parser.Parse(&src)
	if !handler.Empty() {
		errs := handler.AllErrors()
		t.Error(strings.Join(errs, " "))
	}
	dump := ast.Dump(false)
	formatted := formatSExpr(dump)
	unformatted := unformatSExpr(formatted)
	if dump != unformatted {
		t.Errorf("SExpr are not equal {%#v} {%#v}", dump, unformatted)
	}
}
