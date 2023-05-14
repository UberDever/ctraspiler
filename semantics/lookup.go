package semantics

import (
	"fmt"
	s "some/syntax"
	u "some/util"
)

type lookupDecl struct {
	name      string
	line, col int
	level     int
}

func (d lookupDecl) String() string {
	return fmt.Sprintf("%s:%d:%d(%d)", d.name, d.line, d.col, d.level)
}

type lookupEnv struct {
	seenDecls    []lookupDecl
	prevScopeEnd int
	curLevel     int
}

func newLookupEnv() lookupEnv {
	return lookupEnv{
		seenDecls:    make([]lookupDecl, 0, 128),
		prevScopeEnd: 0,
		curLevel:     -1,
	}
}

func (e *lookupEnv) enterScope() {
	e.prevScopeEnd = len(e.seenDecls)
	e.curLevel++
}

func (e *lookupEnv) exitScope() {
	e.seenDecls = e.seenDecls[:e.prevScopeEnd]
	e.curLevel--
}

func (e *lookupEnv) add(d lookupDecl) {
	e.seenDecls = append(e.seenDecls, d)
}

func (e *lookupEnv) lookup(name string) (lookupDecl, bool) {
	for i := len(e.seenDecls) - 1; i >= 0; i-- {
		d := e.seenDecls[i]
		if d.name == name {
			return d, true
		}
	}
	return lookupDecl{}, false
}

type lookupContext struct {
	ast *s.AST
	env lookupEnv

	funcContext s.NodeIndex
}

func LookupPass(src s.Source, ast s.AST, handler *u.ErrorHandler) {
	ctx := lookupContext{
		ast: &ast,
		env: newLookupEnv(),
	}

	addIdentifier := func(src s.Source, ast s.AST, i s.NodeIndex) {
		name := ast.Identifier(ast.GetNode(i))
		lexeme := src.Lexeme(name.Token)
		line, col := src.Location(name.Token)
		ctx.env.add(lookupDecl{
			name:  lexeme,
			line:  line,
			col:   col,
			level: ctx.env.curLevel,
		})
	}

	onEnter := func(ast *s.AST, i s.NodeIndex) (shouldStop bool) {
		n := ast.GetNode(i)
		switch n.Tag() {
		case s.NodeSource:
			ctx.env.enterScope()
		case s.NodeFunctionDecl:
			node := ast.FunctionDecl(n)
			addIdentifier(src, *ast, node.Name)
			ctx.funcContext = i
		case s.NodeSignature:
			// skip signature, it handled in block
			shouldStop = true
		case s.NodeBlock:
			ctx.env.enterScope()
			if ctx.funcContext != s.NodeIndexInvalid {
				f := ast.FunctionDecl(ast.GetNode(ctx.funcContext))
				sig := ast.Signature(ast.GetNode(f.Signature))
				params := s.IdentifierList_Children(*ast, sig.Parameters)
				for _, i := range params {
					addIdentifier(src, *ast, i)
				}
				ctx.funcContext = s.NodeIndexInvalid
			}
		case s.NodeIdentifierList:
			node := ast.IdentifierList(n)
			for _, i := range node.Identifiers {
				addIdentifier(src, *ast, i)
			}
			shouldStop = true
		case s.NodeIdentifier:
			// skip function name, it handled in function decl
			if ctx.funcContext != s.NodeIndexInvalid {
				break
			}
			name := s.Identifier_String(*ast, i)
			decl, ok := ctx.env.lookup(name)
			if !ok {
				handler.Add(u.NewError(
					u.Semantic,
					u.ES_LookupFailed,
					decl.line,
					decl.col,
					src.Filename(),
					name))
			}
		}
		return
	}

	onExit := func(ast *s.AST, i s.NodeIndex) (shouldStop bool) {
		n := ast.GetNode(i)
		switch n.Tag() {
		case s.NodeSource:
			ctx.env.exitScope()
		case s.NodeFunctionDecl:
			ctx.funcContext = s.NodeIndexInvalid

		case s.NodeBlock:
			ctx.env.exitScope()
		}
		return
	}

	ast.Traverse(onEnter, onExit)
}
