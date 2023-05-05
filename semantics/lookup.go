package semantics

import "some/syntax"

type lookupStatus int

const (
	NotFound lookupStatus = iota
	Found
)

type context struct {
	ast   *syntax.AST
	decls map[syntax.Node]bool
}

func Pass(ast *syntax.AST) {
	// ctx := context{
	// 	ast:   ast,
	// 	decls: make(map[syntax.Node]bool),
	// }

}
