package analysis

import (
	"encoding/binary"
	"fmt"
	"math"
	a "some/ast"
	ID "some/domain"
	s "some/syntax"
	u "some/util"
)

type declID int
type QualifiedName []byte

const declTop declID = -1
const declInvalid declID = math.MinInt

type decl struct {
	node         ID.Node
	parent       declID
	line, col    int
	isIdentifier bool
	level        int
}

func (d decl) String(src *s.Source, ast *a.AST) string {
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

type usage struct {
	user ID.Node
	decl declID
}

type usages []usage

func (u *usages) Add(user ID.Node, decl declID) {
	*u = append(*u, usage{user: user, decl: decl})
}

type scopeEnv struct {
	ast          *a.AST
	declarations []decl
	declStack    []decl
	declUsages   usages
	scopeBase    int
	scopeTop     int
	curLevel     int
}

func newScopeEnv(ast *a.AST) scopeEnv {
	return scopeEnv{
		ast:          ast,
		declarations: make([]decl, 0, 4),
		declStack:    make([]decl, 0, 4),
		declUsages:   make([]usage, 0, 4),
		scopeBase:    0,
		scopeTop:     0,
		curLevel:     -1,
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

func (e *scopeEnv) add(d decl) {
	e.declarations = append(e.declarations, d)
	if e.scopeTop < len(e.declStack) {
		e.declStack[e.scopeTop] = d
	} else {
		e.declStack = append(e.declStack, d)
	}
	e.scopeTop++
}

func (e scopeEnv) get(i declID) decl {
	return e.declarations[i]
}

func (e scopeEnv) lookup(node ID.Node) declID {
	for i := e.scopeTop - 1; i >= 0; i-- {
		d := e.declStack[i]
		lhs := a.Identifier_String(*e.ast, node)
		rhs := a.Identifier_String(*e.ast, d.node)
		if lhs == rhs {
			return declID(i)
		}
	}
	return declInvalid
}

func (e scopeEnv) qualifiedName(i declID) QualifiedName {
	buffer := QualifiedName{}
	d := e.declarations[i]
	p := d.parent
	for p != declTop {
		buffer = binary.LittleEndian.AppendUint64(buffer, uint64(d.node))
		buffer = binary.LittleEndian.AppendUint64(buffer, uint64(d.parent))
		d = e.declarations[p]
		p = d.parent
	}
	return buffer
}

type scopecheckContext struct {
	env scopeEnv

	curParent      declID
	inUsageContext bool
}

type ScopeCheckResult struct {
	Ast            *a.AST
	QualifiedNames map[ID.Node]QualifiedName
}

func ScopecheckPass(src *s.Source, ast *a.AST, handler *u.ErrorHandler) ScopeCheckResult {
	ctx := scopecheckContext{
		env:       newScopeEnv(ast),
		curParent: declTop,
	}

	addDecl := func(i ID.Node, isIdentifier bool) declID {
		line, col := src.Location(ast.GetNode(i).Token())
		ctx.env.add(decl{
			parent:       ctx.curParent,
			node:         i,
			line:         line,
			col:          col,
			isIdentifier: isIdentifier,
			level:        ctx.env.curLevel,
		})
		return declID(len(ctx.env.declarations) - 1)
	}

	dump := func(decls []decl, delimiter string) {
		for i := range decls {
			d := decls[i]
			p := d.parent
			name := ""
			for p != declTop {
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
			dump(ctx.env.declarations, " | ")

		case ID.NodeFunctionDecl:
			ctx.env.enterScope()
			ctx.curParent = addDecl(i, false)
			dump(ctx.env.declarations, " | ")

		case ID.NodeBlock:
			ctx.env.enterScope()
			ctx.curParent = addDecl(i, false)
			dump(ctx.env.declarations, " | ")

		case ID.NodeExpression:
			ctx.inUsageContext = true

		case ID.NodeIdentifier:
			if ctx.inUsageContext {
				index := ctx.env.lookup(i)
				if index == declInvalid {
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
				} else {
					ctx.env.declUsages.Add(i, index)
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
			dump(ctx.env.declarations, " | ")
			ctx.env.exitScope()

		case ID.NodeFunctionDecl:
			dump(ctx.env.declarations, " | ")
			ctx.curParent = ctx.env.get(ctx.curParent).parent
			ctx.env.exitScope()

		case ID.NodeBlock:
			dump(ctx.env.declarations, " | ")
			ctx.curParent = ctx.env.get(ctx.curParent).parent
			ctx.env.exitScope()

		case ID.NodeExpression:
			ctx.inUsageContext = false
		}
		return
	}

	ast.TraversePreorder(onEnter, onExit)

	qualifiedNames := make(map[ID.Node]QualifiedName)
	for i, d := range ctx.env.declarations {
		if d.isIdentifier {
			qualifiedNames[d.node] = ctx.env.qualifiedName(declID(i))
		}
	}
	for _, u := range ctx.env.declUsages {
		qualifiedNames[u.user] = ctx.env.qualifiedName(u.decl)
	}
	return ScopeCheckResult{
		Ast:            ast,
		QualifiedNames: qualifiedNames,
	}
}
