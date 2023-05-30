package analysis

import (
	"errors"
	"fmt"
	"regexp"
	a "some/ast"
	s "some/syntax"
	u "some/util"
	"strings"
	"testing"

	"golang.org/x/exp/utf8string"
)

// TODO: This `runSomething` helper should be devided into several procedurals
// for better reusability

// NOTE: more proper way to write typecheck tests would be
// matching result of type inference with actually typed (by hand) code
// but I don't implemented it in grammar and furher(

func runTypecheck(lhs string, pattern string) error {
	text := utf8string.NewString(lhs)
	src := s.NewSource("typing_test", *text)

	handler := u.NewHandler()
	tokenizer := s.NewTokenizer(&handler)
	tokenizer.Tokenize(&src)
	if !handler.IsEmpty() {
		errs := handler.AllErrors()
		return errors.New(strings.Join(errs, ""))
	}

	parser := a.NewParser(&handler)
	ast := parser.Parse(&src)
	if !handler.IsEmpty() {
		errs := handler.AllErrors()
		return errors.New(strings.Join(errs, ""))
	}

	scopeCheck := ScopecheckPass(&src, &ast, &handler)
	if !handler.IsEmpty() {
		errs := handler.AllErrors()
		return errors.New(strings.Join(errs, ""))
	}

	tAst := TypeCheckPass(scopeCheck, &src, &ast, &handler)
	if !handler.IsEmpty() {
		fmt.Println(u.FormatSExpr(ast.Dump(0)))
		fmt.Println(u.FormatSExpr(tAst.Dump()))
		errs := handler.AllErrors()
		return errors.New(strings.Join(errs, ""))
	}

	dump := tAst.Dump()
	matched, err := regexp.Match(pattern, []byte(dump))
	if err != nil {
		fmt.Println(u.FormatSExpr(ast.Dump(0)))
		fmt.Println(u.FormatSExpr(tAst.Dump()))
		return errors.New("Failed to do regexp match")
	}
	if !matched {
		fmt.Println(u.FormatSExpr(ast.Dump(0)))
		fmt.Println(u.FormatSExpr(tAst.Dump()))
		return fmt.Errorf("Can't match pattern %s, typecheck failed", pattern)
	}

	return nil
}

func TestTypecheckFail(t *testing.T) {
	code := `
		fn main() {
			const a = true + 5
		}
	`
	text := utf8string.NewString(code)
	src := s.NewSource("typing_test", *text)

	handler := u.NewHandler()
	tokenizer := s.NewTokenizer(&handler)
	tokenizer.Tokenize(&src)
	if !handler.IsEmpty() {
		t.Fatalf(strings.Join(handler.AllErrors(), ""))
	}

	parser := a.NewParser(&handler)
	ast := parser.Parse(&src)
	if !handler.IsEmpty() {
		t.Fatalf(strings.Join(handler.AllErrors(), ""))
	}

	scopeCheck := ScopecheckPass(&src, &ast, &handler)
	if !handler.IsEmpty() {
		t.Fatalf(strings.Join(handler.AllErrors(), ""))
	}

	tAst := TypeCheckPass(scopeCheck, &src, &ast, &handler)
	if !handler.IsEmpty() {
		messages := handler.AllErrors()
		if len(messages) != 1 {
			t.Fatalf("Expected only one error")
		}
	} else {
		fmt.Println(u.FormatSExpr(ast.Dump(a.DumpShowNodeID)))
		fmt.Println(u.FormatSExpr(tAst.Dump()))
		t.Fatalf("Expected fail on the typecheck")
	}
}

func TestSimpleTypecheck(t *testing.T) {
	code := `
		fn main() {
			const a = 5 + 2
		}
	`
	pattern := "a.*`int`"
	if e := runTypecheck(code, pattern); e != nil {
		t.Error(e)
	}
}

func TestConstantsTypecheck(t *testing.T) {
	code := `
		fn main() {
			const a = 5 + 2
			const b = a
			const c = 4
		}
	`
	patterns := []string{"a.*`int`", "b.*`int`", "c.*`int`"}
	for _, p := range patterns {
		if e := runTypecheck(code, p); e != nil {
			t.Error(e)
			break
		}
	}
}

func TestLogicalOpsTypecheck(t *testing.T) {
	code := `
		fn main() {
			const a = true || false
			const b = !a
			const c = b && !b
		}
	`
	patterns := []string{"a.*`bool`", "b.*`bool`", "c.*`bool`"}
	for _, p := range patterns {
		if e := runTypecheck(code, p); e != nil {
			t.Error(e)
			break
		}
	}
}

func TestVariableTypecheck(t *testing.T) {
	code := `
		fn main() {
			var a = true
			var b = a
			const c = a && b
			var d = 5.2
		}
	`
	patterns := []string{"a.*`bool`", "b.*`bool`", "c.*`bool`", "d.*`float`"}
	for _, p := range patterns {
		if e := runTypecheck(code, p); e != nil {
			t.Error(e)
			break
		}
	}
}
