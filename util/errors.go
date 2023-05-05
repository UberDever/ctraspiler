package util

import "fmt"

type errorKind int
type errorCode int

const (
	Lexer errorKind = iota
	Parser
	Ast
	Semantic
)

var sources = [...]string{
	Lexer:    "Lexer-%04d at %s:%d:%d %s",
	Parser:   "Parser-%04d at %s:%d:%d %s",
	Ast:      "AST-%04d at %s:%d:%d %s",
	Semantic: "Sema-%04d at %s:%d:%d %s",
}

const (
	EP_ExpectedTag errorCode = iota
	EP_ExpectedToken
	EP_ExpectedFunctionDecl
)

var templates = [...][]string{
	Lexer: {
		"No errors here",
	},
	Parser: {
		EP_ExpectedTag:          "Expected tag\n%s\n but got\n%s\n",
		EP_ExpectedFunctionDecl: "At %s expected function declaration",
	},
	Ast:      {},
	Semantic: {},
}

type Error struct {
	kind      errorKind
	code      errorCode
	line, col int
	filename  string
	message   string
}

func NewError(kind errorKind, code errorCode, line, col int, filename string, args ...any) Error {
	// this will panic if kind or code invalid, this is fine
	template := templates[kind][code]
	message := fmt.Sprintf(template, args...)

	return Error{
		kind:     kind,
		code:     code,
		line:     line,
		col:      col,
		filename: filename,
		message:  message,
	}
}

func (e Error) String() string {
	return fmt.Sprintf(sources[e.kind], e.code, e.filename, e.line, e.col, e.message)
}

func (e Error) Kind() errorKind { return e.kind }

func (e Error) Code() errorCode { return e.code }

func (e Error) Position() (line int, col int, filename string) {
	line = e.line
	col = e.col
	filename = e.filename
	return
}

type Handler struct {
	errors []Error
}

func NewHandler() Handler {
	return Handler{errors: make([]Error, 0)}
}

func (h Handler) Empty() bool {
	return len(h.errors) == 0
}

func (h *Handler) Clear() {
	h.errors = make([]Error, 0)
}

func (h *Handler) Add(e Error) {
	h.errors = append(h.errors, e)
}

const Threshold = 10

func (h Handler) AllErrors() []string {
	s := make([]string, 0, len(h.errors))
	n := Max(len(h.errors), Threshold)
	for i := 0; i < n; i++ {
		s = append(s, h.errors[i].message)
	}
	return s
}

var ErrorHandler Handler

func init() { ErrorHandler = NewHandler() }
