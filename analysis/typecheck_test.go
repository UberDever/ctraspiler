package analysis

import (
	"errors"
	"fmt"
	a "some/ast"
	s "some/syntax"
	u "some/util"
	"strings"
	"testing"

	"golang.org/x/exp/utf8string"
)

func runTypecheck(code string) error {
	text := utf8string.NewString(code)
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
	_ = tAst
	if !handler.Empty() {
		errs := handler.AllErrors()
		return errors.New(strings.Join(errs, ""))
	}

	fmt.Println(u.FormatSExpr(tAst.Dump()))

	return nil
}

func TestSimpleTypecheck(t *testing.T) {
	code := `
		fn main() {
			const a = 5 + 2
			{
				a = 8 + 3 + 9
			}
		}
	`
	if e := runTypecheck(code); e != nil {
		t.Error(e)
	}
}
