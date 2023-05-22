package ast

import (
	"errors"
	"fmt"
	ID "some/domain"
	s "some/syntax"
	u "some/util"
	"strings"
	"testing"
	"text/tabwriter"

	"golang.org/x/exp/utf8string"
)

func isASTValid(nodes []Node) (Node, int, bool) {
	for i, n := range nodes {
		if n.tokenIdx == ID.TokenInvalid ||
			n.lhs == ID.NodeInvalid ||
			n.rhs == ID.NodeInvalid {
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
	src := s.NewSource("ast_test", *text)

	handler := u.NewHandler()
	tokenizer := s.NewTokenizer(&handler)
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

	expected := u.MinifySExpr(utf8string.NewString(rhs).String())
	result := ast.Dump(0)

	if node, index, ok := isASTValid(ast.nodes); !ok {
		return fmt.Errorf("AST nodes failed on validity test at %d => %v", index, node)
	}
	if result != expected {
		lhs = u.FormatSExpr(result)
		rhs = u.FormatSExpr(expected)
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
				(Signature (ID[])))
			(FunctionDecl (some)
				(Signature (ID[] (a) (b))
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
			(Signature (ID[]))
			(Block 
				(ConstDecl 
					(ID[] (_)) 
					(Expr[] (Expr (null)
				)))
				(ConstDecl 
					(ID[] (a) (b)) 
					(Expr[] (Expr (8)) (Expr (2))
				))
				(ConstDecl 
					(ID[] (c) (d) (e))
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
			(Signature (ID[]))
			(Block 
				(ConstDecl (ID[] (x)) (Expr[] (Expr (8))))
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
			(Signature (ID[]))
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
	src := s.NewSource("ast_test", *text)

	handler := u.NewHandler()
	tokenizer := s.NewTokenizer(&handler)
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
	dump := ast.Dump(0)
	formatted := u.FormatSExpr(dump)
	unformatted := u.MinifySExpr(formatted)
	if dump != unformatted {
		t.Errorf("SExpr are not equal {%#v} {%#v}", dump, unformatted)
	}
}
