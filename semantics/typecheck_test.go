package semantics

import (
	"errors"
	sx "some/syntax"
	"some/util"
	"strings"
	"testing"

	"golang.org/x/exp/utf8string"
)

func runTypecheck(code string) error {
	text := utf8string.NewString(code)
	src := sx.NewSource("typing_test", *text)

	handler := util.NewHandler()
	tokenizer := sx.NewTokenizer(&handler)
	tokenizer.Tokenize(&src)
	if !handler.Empty() {
		errs := handler.AllErrors()
		return errors.New(strings.Join(errs, ""))
	}

	parser := sx.NewParser(&handler)
	ast := parser.Parse(&src)
	if !handler.Empty() {
		errs := handler.AllErrors()
		return errors.New(strings.Join(errs, ""))
	}

	scopeCheck := ScopecheckPass(&src, &ast, &handler)
	if !handler.Empty() {
		errs := handler.AllErrors()
		return errors.New(strings.Join(errs, ""))
	}

	TypeCheckPass(scopeCheck, &src, &ast, &handler)
	if !handler.Empty() {
		errs := handler.AllErrors()
		return errors.New(strings.Join(errs, ""))
	}

	return nil
}

func TestSimpleTypecheck(t *testing.T) {
	code := `
		fn some(a, b) {
			return a * b
		}

		fn main() {
			const c = some(2, 3)
		}
	`
	if e := runTypecheck(code); e != nil {
		t.Error(e)
	}
}
