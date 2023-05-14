package semantics

import (
	"fmt"
	s "some/syntax"
	u "some/util"
)

type scopeIndex int

const scopeTop scopeIndex = -1

type scopedDecl struct {
	node      s.NodeIndex
	parent    scopeIndex
	line, col int
	level     int
}

func (d scopedDecl) String() string {
	return fmt.Sprintf("%d-%d:%d(%d)", d.node, d.line, d.col, d.level)
}

type scopeEnv struct {
	ast          *s.AST
	seenDecls    []scopedDecl
	prevScopeEnd int
	curLevel     int
}

func newScopeEnv(ast *s.AST) scopeEnv {
	return scopeEnv{
		ast:          ast,
		seenDecls:    make([]scopedDecl, 0, 128),
		prevScopeEnd: 0,
		curLevel:     -1,
	}
}

func (e *scopeEnv) enterScope() {
	e.prevScopeEnd = len(e.seenDecls)
	e.curLevel++
}

func (e *scopeEnv) exitScope() {
	e.seenDecls = e.seenDecls[:e.prevScopeEnd]
	e.curLevel--
}

func (e *scopeEnv) add(d scopedDecl) {
	// TODO: add shadowing avoidance and redeclaration handling
	e.seenDecls = append(e.seenDecls, d)
}

func (e *scopeEnv) get(i scopeIndex) scopedDecl {
	return e.seenDecls[i]
}

func (e *scopeEnv) lookup(node s.NodeIndex) (scopedDecl, bool) {
	for i := len(e.seenDecls) - 1; i >= 0; i-- {
		d := e.seenDecls[i]
		lhs := s.Identifier_String(*e.ast, node)
		rhs := s.Identifier_String(*e.ast, d.node)
		if lhs == rhs {
			return d, true
		}
	}
	return scopedDecl{}, false
}

type scopecheckContext struct {
	env scopeEnv

	curParent    scopeIndex
	usageContext bool
}

func ScopecheckPass(src *s.Source, ast *s.AST, handler *u.ErrorHandler) {
	ctx := scopecheckContext{
		env:       newScopeEnv(ast),
		curParent: scopeTop,
	}

	addDecl := func(i s.NodeIndex) scopeIndex {
		line, col := src.Location(ast.GetNode(i).Token())
		ctx.env.add(scopedDecl{
			parent: ctx.curParent,
			node:   i,
			line:   line,
			col:    col,
			level:  ctx.env.curLevel,
		})
		return scopeIndex(len(ctx.env.seenDecls) - 1)
	}

	dump := func() {
		for i := range ctx.env.seenDecls {
			d := ctx.env.seenDecls[i]
			name := d.String()
			p := d.parent
			for p != scopeTop {
				d := ctx.env.seenDecls[p]
				name = d.String() + "/" + name
				p = d.parent
			}
			fmt.Printf("%s ", name)
		}
		fmt.Println("")
	}

	// shouldStop used here to stop processing declarations
	// and let `case NodeIdentifier` process all usages
	onEnter := func(ast *s.AST, i s.NodeIndex) (shouldStop bool) {
		n := ast.GetNode(i)
		switch n.Tag() {
		case s.NodeSource:
			dump()
			ctx.curParent = addDecl(i)
			ctx.env.enterScope()
			dump()

		case s.NodeFunctionDecl:
			dump()
			ctx.curParent = addDecl(i)
			ctx.env.enterScope()
			dump()

		case s.NodeBlock:
			dump()
			ctx.curParent = addDecl(i)
			ctx.env.enterScope()
			dump()

		case s.NodeExpression:
			ctx.usageContext = true

		case s.NodeIdentifier:
			// id := ast.Identifier(ast.GetNode(i))
			// decl, ok := ctx.env.lookup(id.Token)
			// if !ok {
			// 	name := s.Identifier_String(*ast, i)
			// 	handler.Add(u.NewError(
			// 		u.Semantic,
			// 		u.ES_ScopecheckFailed,
			// 		decl.line,
			// 		decl.col,
			// 		src.Filename(),
			// 		name))
			// }
		}
		return
	}

	onExit := func(ast *s.AST, i s.NodeIndex) (shouldStop bool) {
		n := ast.GetNode(i)
		switch n.Tag() {
		case s.NodeSource:
			dump()
			ctx.env.exitScope()
			dump()

		case s.NodeFunctionDecl:
			dump()
			ctx.env.exitScope()
			ctx.curParent = ctx.env.get(ctx.curParent).parent
			dump()

		case s.NodeBlock:
			dump()
			ctx.env.exitScope()
			ctx.curParent = ctx.env.get(ctx.curParent).parent
			dump()

		case s.NodeExpression:
			ctx.usageContext = false
		}
		return
	}

	ast.Traverse(onEnter, onExit)
}
