package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	p "some/ast"
	s "some/syntax"
	u "some/util"
	"strings"

	"golang.org/x/exp/utf8string"
)

func main() {
	filepath := flag.String("path", "undefined", "filepath to source")
	flag.Parse()
	contents, err := ioutil.ReadFile(*filepath)
	if err != nil {
		panic(err)
	}

	text := utf8string.NewString(string(contents))
	src := s.NewSource("ast_test", *text)

	handler := u.NewHandler()
	tokenizer := s.NewTokenizer(&handler)
	tokenizer.Tokenize(&src)
	if !handler.IsEmpty() {
		errs := handler.AllErrors()
		panic(errors.New(strings.Join(errs, "")))
	}

	parser := p.NewParser(&handler)
	ast := parser.Parse(&src)
	if !handler.IsEmpty() {
		errs := handler.AllErrors()
		panic(errors.New(strings.Join(errs, "")))
	}

	dump := ast.Dump(0)
	formatted := u.FormatSExpr(dump)

	fmt.Println(formatted)
}
