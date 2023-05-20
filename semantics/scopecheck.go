package semantics

import (
	"encoding/binary"
	"fmt"
	a "some/ast"
	ID "some/domain"
	s "some/syntax"
	u "some/util"
)

type scopeID int
type UniqueName []byte

const scopeTop scopeID = -1

type scopeDecl struct {
	node         ID.Node
	parent       scopeID
	line, col    int
	isIdentifier bool
	level        int
}

func (d scopeDecl) String(src *s.Source, ast *a.AST) string {
	n := ast.GetNode(d.node)
	if n.Tag() == ID.NodeSource {
		return "SRC"
	}
	if n.Tag() == ID.NodeFunctionDecl {
		return "FN"
	}
	if n.Tag() == ID.NodeBlock {
		return "BLK"
	}
	t := n.Token()
	lexeme := src.Lexeme(t)
	return fmt.Sprintf("%s[%d]", lexeme, t)
}

type scopeEnv struct {
	ast       *a.AST
	allDecls  []scopeDecl
	seenDecls []scopeDecl
	scopeBase int
	scopeTop  int
	curLevel  int
}

func newScopeEnv(ast *a.AST) scopeEnv {
	return scopeEnv{
		ast:       ast,
		allDecls:  make([]scopeDecl, 0, 4),
		seenDecls: make([]scopeDecl, 0, 4),
		scopeBase: 0,
		scopeTop:  0,
		curLevel:  -1,
	}
}

func (e *scopeEnv) enterScope() {
	e.scopeBase = e.scopeTop
	e.curLevel++
}

func (e *scopeEnv) exitScope() {
	e.scopeTop = e.scopeBase
	e.curLevel--
}

func (e *scopeEnv) add(d scopeDecl) {
	e.allDecls = append(e.allDecls, d)
	if e.scopeTop < len(e.seenDecls) {
		e.seenDecls[e.scopeTop] = d
	} else {
		e.seenDecls = append(e.seenDecls, d)
	}
	e.scopeTop++
}

func (e scopeEnv) get(i scopeID) scopeDecl {
	return e.allDecls[i]
}

func (e scopeEnv) lookup(node ID.Node) bool {
	for i := e.scopeTop - 1; i >= 0; i-- {
		d := e.seenDecls[i]
		lhs := a.Identifier_String(*e.ast, node)
		rhs := a.Identifier_String(*e.ast, d.node)
		if lhs == rhs {
			return true
		}
	}
	return false
}

func (e scopeEnv) uniqueName(i scopeID) UniqueName {
	buffer := UniqueName{}
	d := e.allDecls[i]
	p := d.parent
	for p != scopeTop {
		buffer = binary.LittleEndian.AppendUint64(buffer, uint64(d.node))
		buffer = binary.LittleEndian.AppendUint64(buffer, uint64(d.parent))
		d = e.allDecls[p]
		p = d.parent
	}
	return buffer
}

type scopecheckContext struct {
	env scopeEnv

	curParent      scopeID
	inUsageContext bool
}

type ScopeCheckResult struct {
	Ast         *a.AST
	UniqueNames map[ID.Node]UniqueName
}

func ScopecheckPass(src *s.Source, ast *a.AST, handler *u.ErrorHandler) ScopeCheckResult {
	ctx := scopecheckContext{
		env:       newScopeEnv(ast),
		curParent: scopeTop,
	}

	addDecl := func(i ID.Node, isIdentifier bool) scopeID {
		line, col := src.Location(ast.GetNode(i).Token())
		ctx.env.add(scopeDecl{
			parent:       ctx.curParent,
			node:         i,
			line:         line,
			col:          col,
			isIdentifier: isIdentifier,
			level:        ctx.env.curLevel,
		})
		return scopeID(len(ctx.env.allDecls) - 1)
	}

	dump := func(decls []scopeDecl, delimiter string) {
		for i := range decls {
			d := decls[i]
			p := d.parent
			name := ""
			for p != scopeTop {
				name = "." + d.String(src, ast) + fmt.Sprintf("(%d)", p) + name
				d = decls[p]
				p = d.parent
			}
			name = "src" + name
			fmt.Printf("%d: %s%s", i, name, delimiter)
		}
		fmt.Println("")
	}

	onEnter := func(ast *a.AST, i ID.Node) (shouldStop bool) {
		n := ast.GetNode(i)
		switch n.Tag() {
		case ID.NodeSource:
			ctx.env.enterScope()
			ctx.curParent = addDecl(i, false)
			dump(ctx.env.allDecls, " | ")

		case ID.NodeFunctionDecl:
			ctx.env.enterScope()
			ctx.curParent = addDecl(i, false)
			dump(ctx.env.allDecls, " | ")

		case ID.NodeBlock:
			ctx.env.enterScope()
			ctx.curParent = addDecl(i, false)
			dump(ctx.env.allDecls, " | ")

		case ID.NodeExpression:
			ctx.inUsageContext = true

		case ID.NodeIdentifier:
			if ctx.inUsageContext {
				seen := ctx.env.lookup(i)
				if !seen {
					id := ast.Identifier(ast.GetNode(i)).Token
					name := src.Lexeme(id)
					line, col := src.Location(id)
					handler.Add(u.NewError(
						u.Semantic,
						u.ES_ScopecheckFailed,
						line,
						col,
						src.Filename(),
						name))
				}
			} else {
				addDecl(i, true)
			}
		}
		return
	}

	onExit := func(ast *a.AST, i ID.Node) (shouldStop bool) {
		n := ast.GetNode(i)
		switch n.Tag() {
		case ID.NodeSource:
			dump(ctx.env.allDecls, " | ")
			ctx.env.exitScope()

		case ID.NodeFunctionDecl:
			dump(ctx.env.allDecls, " | ")
			ctx.curParent = ctx.env.get(ctx.curParent).parent
			ctx.env.exitScope()

		case ID.NodeBlock:
			dump(ctx.env.allDecls, " | ")
			ctx.curParent = ctx.env.get(ctx.curParent).parent
			ctx.env.exitScope()

		case ID.NodeExpression:
			ctx.inUsageContext = false
		}
		return
	}

	ast.TraversePreorder(onEnter, onExit)

	uniqueNames := make(map[ID.Node]UniqueName)
	for i, d := range ctx.env.allDecls {
		if d.isIdentifier {
			uniqueNames[d.node] = ctx.env.uniqueName(scopeID(i))
		}
	}
	return ScopeCheckResult{
		Ast:         ast,
		UniqueNames: uniqueNames,
	}
}
