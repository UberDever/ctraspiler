package semantics

import (
	"errors"
	s "some/syntax"
	"some/util"
	"strings"
	"testing"

	"golang.org/x/exp/utf8string"
)

func runScopecheck(code string) error {
	text := utf8string.NewString(code)
	src := s.NewSource("lookup_test", *text)

	handler := util.NewHandler()
	tokenizer := s.NewTokenizer(&handler)
	tokenizer.Tokenize(&src)
	if !handler.Empty() {
		errs := handler.AllErrors()
		return errors.New(strings.Join(errs, ""))
	}

	parser := s.NewParser(&handler)
	ast := parser.Parse(&src)
	if !handler.Empty() {
		errs := handler.AllErrors()
		return errors.New(strings.Join(errs, ""))
	}

	ScopecheckPass(&src, &ast, &handler)
	if !handler.Empty() {
		errs := handler.AllErrors()
		return errors.New(strings.Join(errs, ""))
	}
	return nil
}

func TestScopecheckDeclarations(t *testing.T) {
	code := `
		fn main()
		fn some(a, b) { }
		fn complex(a, b, d) {
			const v1, v2 = 3, 4
			const v3 = v2
			some(a, v3)
		}
	`
	if e := runScopecheck(code); e != nil {
		t.Error(e)
	}
}

func TestScopecheckFailed(t *testing.T) {
	code := `
		fn main() {
			some(2, 3)
		}

		fn some(a, b) { }

		fn complex(a, b, d) {
			const a = c + d
		}
	`
	e := runScopecheck(code)
	if e == nil {
		t.Error("Exprected failed lookup")
	}
	failed := []string{
		"identifier some failed",
		"identifier c failed",
	}
	for _, fail := range failed {
		if !strings.Contains(e.Error(), fail) {
			t.Error("Scopecheck fail error message malformed")
		}
	}
}
