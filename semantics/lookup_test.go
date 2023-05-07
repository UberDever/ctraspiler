package semantics

import (
	"errors"
	sx "some/syntax"
	"some/util"
	"strings"
	"testing"

	"golang.org/x/exp/utf8string"
)

func runTest(code string) error {
	text := utf8string.NewString(code)
	src := sx.NewSource("lookup_test", *text)

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

	LookupPass(src, ast, &handler)
	if !handler.Empty() {
		errs := handler.AllErrors()
		return errors.New(strings.Join(errs, ""))
	}
	return nil
}

func TestLookupDeclarations(t *testing.T) {
	code := `
		fn main()
		fn some(a, b) { }
		fn complex(a, b, d) {
			const v1, v2 = 3, 4
			const v3 = v2
			some(a, v3)
		}
	`
	if e := runTest(code); e != nil {
		t.Error(e)
	}
}

func TestLookupFailed(t *testing.T) {
	code := `
		fn main() {
			some(2, 3)
		}

		fn some(a, b) { }

		fn complex(a, b, d) {
			const a = c + d
		}
	`
	e := runTest(code)
	if e == nil {
		t.Error("Exprected failed lookup")
	}
	failed := []string{
		"identifier some failed",
		"identifier c failed",
	}
	for _, fail := range failed {
		if !strings.Contains(e.Error(), fail) {
			t.Error("Lookup fail error message malformed")
		}
	}
}
