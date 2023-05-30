package analysis

import (
	"fmt"
	"math"
	a "some/ast"
	ID "some/domain"
	s "some/syntax"
	u "some/util"
	"strings"
)

type declID int

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
	buffer := strings.Builder{}
	d := e.declarations[i]
	p := d.parent
	for p != declTop {
		buffer.WriteString(fmt.Sprint(d.node))
		buffer.WriteByte('\'')
		buffer.WriteString(fmt.Sprint(d.parent))
		buffer.WriteByte('\'')
		d = e.declarations[p]
		p = d.parent
	}
	return QualifiedName(buffer.String())
}

type QualifiedName string
type QualifiedNames struct {
	names     []QualifiedName
	declNodes []ID.Node
	nodeNames map[ID.Node]uint
}

func NewQualifiedNames(ctx *scopecheckContext) QualifiedNames {
	names := QualifiedNames{
		names:     make([]QualifiedName, 0, 16),
		declNodes: make([]ID.Node, 0, 16),
		nodeNames: make(map[ID.Node]uint),
	}
	declToName := map[declID]uint{}
	for i, d := range ctx.env.declarations {
		if d.isIdentifier {
			name := ctx.env.qualifiedName(declID(i))
			names.names = append(names.names, name)
			last := uint(len(names.names) - 1)
			names.declNodes = append(names.declNodes, d.node)
			names.nodeNames[d.node] = last
			declToName[declID(i)] = last
		}
	}
	for _, u := range ctx.env.declUsages {
		names.nodeNames[u.user] = declToName[u.decl]
	}
	return names
}

func (n QualifiedNames) GetNodeName(id ID.Node) (QualifiedName, bool) {
	i, has := n.nodeNames[id]
	return n.names[i], has
}

func (n QualifiedNames) GetDeclarationNode(name QualifiedName) ID.Node {
	for _, id := range n.declNodes {
		nameID := n.nodeNames[id]
		if n.names[nameID] == name {
			return id
		}
	}
	return ID.NodeInvalid
}

func (n QualifiedNames) GetDeclarations() []ID.Node {
	return n.declNodes
}

type scopecheckContext struct {
	env scopeEnv

	curParent      declID
	inUsageContext bool
}

type ScopeCheckResult struct {
	Ast            *a.AST
	QualifiedNames QualifiedNames
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

	dump := func(decls []decl, usages []usage, delimiter string) {
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
			isUsage := " used by"
			for _, u := range usages {
				if u.decl == declID(i) {
					isUsage += fmt.Sprintf(" %d", u.user)
				}
			}
			fmt.Printf("%d: %s%s%s", i, name, isUsage, delimiter)
		}
		fmt.Println("")
	}
	_ = dump

	onEnter := func(ast *a.AST, i ID.Node) (shouldStop bool) {
		n := ast.GetNode(i)
		switch n.Tag() {
		case ID.NodeSource:
			ctx.env.enterScope()
			ctx.curParent = addDecl(i, false)

		case ID.NodeFunctionDecl:
			ctx.env.enterScope()
			ctx.curParent = addDecl(i, false)

		case ID.NodeBlock:
			ctx.env.enterScope()
			ctx.curParent = addDecl(i, false)

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
			ctx.env.exitScope()

		case ID.NodeFunctionDecl:
			ctx.curParent = ctx.env.get(ctx.curParent).parent
			ctx.env.exitScope()

		case ID.NodeBlock:
			ctx.curParent = ctx.env.get(ctx.curParent).parent
			ctx.env.exitScope()

		case ID.NodeExpression:
			ctx.inUsageContext = false
		}
		return
	}

	ast.TraversePreorder(onEnter, onExit)

	// dump(ctx.env.declarations, ctx.env.declUsages, "\n")
	return ScopeCheckResult{
		Ast:            ast,
		QualifiedNames: NewQualifiedNames(&ctx),
	}
}
