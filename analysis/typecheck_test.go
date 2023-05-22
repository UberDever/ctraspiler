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

// NOTE: more proper way to write typecheck tests would be
// matching result of type inference with actually typed (by hand) code
// but I don't implemented it in grammar and furher(

func runTypecheck(lhs string, pattern string) error {
	text := utf8string.NewString(lhs)
	src := s.NewSource("typing_test", *text)

	handler := u.NewHandler()
	tokenizer := s.NewTokenizer(&handler)
	tokenizer.Tokenize(&src)
	if !handler.Empty() {
		errs := handler.AllErrors()
		return errors.New(strings.Join(errs, ""))
	}

	parser := a.NewParser(&handler)
	ast := parser.Parse(&src)
	if !handler.Empty() {
		errs := handler.AllErrors()
		return errors.New(strings.Join(errs, ""))
	}
	fmt.Println(u.FormatSExpr(ast.Dump()))

	scopeCheck := ScopecheckPass(&src, &ast, &handler)
	if !handler.Empty() {
		errs := handler.AllErrors()
		return errors.New(strings.Join(errs, ""))
	}

	tAst := TypeCheckPass(scopeCheck, &src, &ast, &handler)
	if !handler.Empty() {
		errs := handler.AllErrors()
		return errors.New(strings.Join(errs, ""))
	}

	dump := tAst.Dump()
	matched, err := regexp.Match(pattern, []byte(dump))
	if err != nil {
		return errors.New("Failed to do regexp match")
	}
	if !matched {
		return fmt.Errorf("Can't match pattern %s, typecheck failed", pattern)
	}

	fmt.Println(u.FormatSExpr(tAst.Dump()))

	return nil
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

func TestVariableTypecheck(t *testing.T) {
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
		}
	}
}
