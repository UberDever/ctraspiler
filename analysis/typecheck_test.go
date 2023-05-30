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

type compiler struct {
	src              s.Source
	ast              a.AST
	scopecheckResult ScopeCheckResult
	tAst             a.TypedAST
	handler          u.ErrorHandler
}

func newCompiler(code string) (c compiler) {
	text := utf8string.NewString(code)
	c.src = s.NewSource("typing_test", *text)
	c.handler = u.NewHandler()
	return
}

func (c *compiler) tokenize() error {
	tokenizer := s.NewTokenizer(&c.handler)
	tokenizer.Tokenize(&c.src)
	if !c.handler.IsEmpty() {
		errs := c.handler.AllErrors()
		return errors.New(strings.Join(errs, ""))
	}
	return nil
}

func (c *compiler) parse() error {
	parser := a.NewParser(&c.handler)
	c.ast = parser.Parse(&c.src)
	if !c.handler.IsEmpty() {
		errs := c.handler.AllErrors()
		return errors.New(strings.Join(errs, ""))
	}
	return nil
}

func (c *compiler) scopecheck() error {
	c.scopecheckResult = ScopecheckPass(&c.src, &c.ast, &c.handler)
	if !c.handler.IsEmpty() {
		errs := c.handler.AllErrors()
		return errors.New(strings.Join(errs, ""))
	}
	return nil
}

func (c *compiler) typecheck() error {
	c.tAst = TypeCheckPass(c.scopecheckResult, &c.src, &c.ast, &c.handler)
	if !c.handler.IsEmpty() {
		errs := c.handler.AllErrors()
		return errors.New(strings.Join(errs, ""))
	}
	return nil
}

// NOTE: more proper way to write typecheck tests would be
// matching result of type inference with actually typed (by hand) code
// but I don't implemented it in grammar and furher(

func runTypecheck(code string, pattern string) error {
	c := newCompiler(code)
	if err := c.tokenize(); err != nil {
		return err
	}
	if err := c.parse(); err != nil {
		return err
	}
	if err := c.scopecheck(); err != nil {
		return err
	}
	if err := c.typecheck(); err != nil {
		return err
	}
	astDump := c.ast.Dump(a.DumpShowNodeID)
	tAstDump := c.tAst.Dump()
	matched, err := regexp.Match(pattern, []byte(tAstDump))
	if err != nil {
		fmt.Println(u.FormatSExpr(astDump))
		fmt.Println(u.FormatSExpr(tAstDump))
		return errors.New("Failed to do regexp match")
	}
	if !matched {
		fmt.Println(u.FormatSExpr(astDump))
		fmt.Println(u.FormatSExpr(tAstDump))
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
	c := newCompiler(code)
	if err := c.tokenize(); err != nil {
		t.Fatal(err)
	}
	if err := c.parse(); err != nil {
		t.Fatal(err)
	}
	if err := c.scopecheck(); err != nil {
		t.Fatal(err)
	}
	if err := c.typecheck(); err != nil {
		messages := c.handler.AllErrors()
		if len(messages) != 1 {
			t.Fatalf("Expected only one error")
		}
	} else {
		fmt.Println(u.FormatSExpr(c.ast.Dump(a.DumpShowNodeID)))
		fmt.Println(u.FormatSExpr(c.tAst.Dump()))
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

func TestFunctionTypecheck(t *testing.T) {
	code := `
	 	fn unary(a) {
			return -a
		}

		fn some(f, a, b) {
			if a == b {
				return f(-1)
			}
			return f(1)
		}
		
		fn main() {
			return some(unary, true, false)
		}
	`
	patterns := []string{
		"unary.*`(FN int int)`",
		"some.*`(FN (FN int int) bool bool)`",
		"main.*`(FN int)`",
	}
	for _, p := range patterns {
		if e := runTypecheck(code, p); e != nil {
			t.Error(e)
			break
		}
	}
}
