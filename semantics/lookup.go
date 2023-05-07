package semantics

import (
	"fmt"
	sx "some/syntax"
	"some/util"
)

type decl struct {
	name      string
	line, col int
	level     int
}

func (d decl) String() string {
	return fmt.Sprintf("%s:%d:%d(%d)", d.name, d.line, d.col, d.level)
}

type environment struct {
	seenDecls    []decl
	prevScopeEnd int
	curLevel     int
}

func newEnvironment() environment {
	return environment{
		seenDecls:    make([]decl, 0, 128),
		prevScopeEnd: 0,
		curLevel:     -1,
	}
}

func (e *environment) enterScope() {
	e.prevScopeEnd = len(e.seenDecls)
	e.curLevel++
}

func (e *environment) exitScope() {
	e.seenDecls = e.seenDecls[:e.prevScopeEnd]
	e.curLevel--
}

func (e *environment) add(d decl) {
	e.seenDecls = append(e.seenDecls, d)
}

func (e *environment) lookup(name string) (decl, bool) {
	for i := len(e.seenDecls) - 1; i >= 0; i-- {
		d := e.seenDecls[i]
		if d.name == name {
			return d, true
		}
	}
	return decl{}, false
}

type context struct {
	ast *sx.AST
	env environment

	funcContext sx.NodeIndex
}

func LookupPass(src sx.Source, ast sx.AST, handler *util.ErrorHandler) {
	ctx := context{
		ast: &ast,
		env: newEnvironment(),
	}

	addIdentifier := func(src sx.Source, ast sx.AST, i sx.NodeIndex) {
		name := ast.Identifier(ast.GetNode(i))
		lexeme := src.Lexeme(name.Token)
		line, col := src.Location(name.Token)
		ctx.env.add(decl{
			name:  lexeme,
			line:  line,
			col:   col,
			level: ctx.env.curLevel,
		})
	}

	onEnter := func(ast *sx.AST, i sx.NodeIndex) (shouldStop bool) {
		n := ast.GetNode(i)
		switch n.Tag() {
		case sx.NodeSource:
			ctx.env.enterScope()
		case sx.NodeFunctionDecl:
			node := ast.FunctionDecl(n)
			addIdentifier(src, *ast, node.Name)
			ctx.funcContext = i
		case sx.NodeSignature:
			// skip signature, it handled in block
			shouldStop = true
		case sx.NodeBlock:
			ctx.env.enterScope()
			if ctx.funcContext != sx.NodeIndexInvalid {
				f := ast.FunctionDecl(ast.GetNode(ctx.funcContext))
				s := ast.Signature(ast.GetNode(f.Signature))
				params := sx.IdentifierList_Children(*ast, s.Parameters)
				for _, i := range params {
					addIdentifier(src, *ast, i)
				}
				ctx.funcContext = sx.NodeIndexInvalid
			}
		case sx.NodeIdentifierList:
			node := ast.IdentifierList(n)
			for _, i := range node.Identifiers {
				addIdentifier(src, *ast, i)
			}
			shouldStop = true
		case sx.NodeIdentifier:
			// skip function name, it handled in function decl
			if ctx.funcContext != sx.NodeIndexInvalid {
				break
			}
			name := sx.Identifier_String(*ast, i)
			decl, ok := ctx.env.lookup(name)
			if !ok {
				handler.Add(util.NewError(
					util.Semantic,
					util.ES_LookupFailed,
					decl.line,
					decl.col,
					src.Filename(),
					name))
			}
		}
		return
	}

	onExit := func(ast *sx.AST, i sx.NodeIndex) (shouldStop bool) {
		n := ast.GetNode(i)
		switch n.Tag() {
		case sx.NodeSource:
			ctx.env.exitScope()
		case sx.NodeFunctionDecl:
			ctx.funcContext = sx.NodeIndexInvalid

		case sx.NodeBlock:
			ctx.env.exitScope()
		}
		return
	}

	ast.Traverse(onEnter, onExit)
}
