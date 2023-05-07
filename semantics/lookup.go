package semantics

import (
	"fmt"
	sx "some/syntax"
	"some/util"
)

type lookupStatus int

const (
	NotFound lookupStatus = iota
	Found
)

type decl struct {
	name string
}

type environment struct {
	seenDecls    []decl
	prevScopeEnd int
}

func newEnvironment() environment {
	return environment{
		seenDecls:    make([]decl, 0, 128),
		prevScopeEnd: 0,
	}
}

func (e *environment) enterScope() {
	e.prevScopeEnd = len(e.seenDecls)
}

func (e *environment) exitScope() {
	e.seenDecls = e.seenDecls[:e.prevScopeEnd]
}

func (e *environment) add(d decl) {
	e.seenDecls = append(e.seenDecls, d)
}

func (e *environment) lookup(name string) {

}

type context struct {
	ast *sx.AST
	env environment

	funcContext sx.NodeIndex
}

func LookupPass(ast *sx.AST, handler *util.ErrorHandler) {
	ctx := context{
		ast: ast,
		env: newEnvironment(),
	}

	onEnter := func(ast *sx.AST, i sx.NodeIndex) (shouldStop bool) {
		n := ast.GetNode(i)
		switch n.Tag() {
		case sx.NodeSource:
			ctx.env.enterScope()
			fmt.Println("enter: ", ctx.env.seenDecls)
		case sx.NodeFunctionDecl:
			node := ast.FunctionDecl(n)
			name := sx.Identifier_String(*ast, node.Name)
			ctx.env.add(decl{name})
			ctx.funcContext = i
		case sx.NodeConstDecl:
			node := ast.ConstDecl(n)
			ids := sx.IdentifierList_Children(*ast, node.IdentifierList)
			for _, i := range ids {
				ctx.env.add(decl{sx.Identifier_String(*ast, i)})
			}
		case sx.NodeBlock:
			ctx.env.enterScope()

			if ctx.funcContext != sx.NodeIndexInvalid {
				f := ast.FunctionDecl(ast.GetNode(ctx.funcContext))
				s := ast.Signature(ast.GetNode(f.Signature))
				params := sx.IdentifierList_Children(*ast, s.Parameters)
				for _, i := range params {
					ctx.env.add(decl{sx.Identifier_String(*ast, i)})
				}
				ctx.funcContext = sx.NodeIndexInvalid
			}
			fmt.Println("enter: ", ctx.env.seenDecls)
		}
		return
	}

	onExit := func(ast *sx.AST, i sx.NodeIndex) (shouldStop bool) {
		n := ast.GetNode(i)
		switch n.Tag() {
		case sx.NodeSource:
			fmt.Println("exit: ", ctx.env.seenDecls)
			ctx.env.exitScope()
		case sx.NodeFunctionDecl:
			ctx.funcContext = sx.NodeIndexInvalid

		case sx.NodeBlock:
			fmt.Println("exit: ", ctx.env.seenDecls)
			ctx.env.exitScope()
		}
		return
	}

	ast.Traverse(onEnter, onExit)
}
