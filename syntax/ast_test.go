package syntax

import (
	"errors"
	"fmt"
	u "some/util"
	"strings"
	"testing"
	"text/tabwriter"

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

func concatVertically(lhs, rhs string) string {
	builder := strings.Builder{}
	table := tabwriter.NewWriter(&builder, 1, 4, 1, ' ', 0)
	lhs_split := strings.Split(lhs, "\n")
	rhs_split := strings.Split(rhs, "\n")
	maxlen := u.Max(len(lhs_split), len(rhs_split))

	for i := 0; i < maxlen; i++ {
		if i >= len(lhs_split) {
			fmt.Fprint(table, "...")
		} else {
			fmt.Fprint(table, lhs_split[i])
		}
		fmt.Fprint(table, "\t")
		if i >= len(rhs_split) {
			fmt.Fprint(table, "...")
		} else {
			fmt.Fprint(table, rhs_split[i])
		}
		fmt.Fprintln(table, "")
	}
	table.Flush()
	return builder.String()
}

func runTest(lhs string, rhs string) error {
	text := utf8string.NewString(lhs)
	src := NewSource("ast_test", *text)

	handler := u.NewHandler()
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
		lhs = formatSExpr(result)
		rhs = formatSExpr(expected)
		trace := concatVertically(lhs, rhs)
		return fmt.Errorf("AST are not equal\n%s", trace)
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
					(Expr[] (Expr (null)
				)))
				(ConstDecl 
					(a b) 
					(Expr[] (Expr (8)) (Expr (2))
				))
				(ConstDecl 
					(c d e) 
					(Expr[] 
						(Expr (* (8) (3)))
						(Expr (- (16)))
						(Expr ("E"))
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
				(ConstDecl (x) (Expr[] (Expr (8))))
				(Expr (+ (* (x) (8)) (3)))
				(Expr (+ (x) (/ (3) (4))))
				(Expr (Call (f) 
					(Expr[] (Expr (x)) (Expr (Get (x) (y))))))
				(Assign
					(Expr[] (Expr (x)) (Expr (Get (x) (y))))
					(Expr[] (Expr (Get (x) (y))) (Expr (x))))
				
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
				(Return (Expr[] (Expr (0)
	))))))`
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

	handler := u.NewHandler()
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
