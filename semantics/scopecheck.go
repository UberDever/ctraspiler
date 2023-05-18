package semantics

import (
	"encoding/binary"
	"fmt"
	s "some/syntax"
	u "some/util"
)

type scopeIndex int
type UniqueName []byte

const scopeTop scopeIndex = -1

type scopeDecl struct {
	node         s.NodeIndex
	parent       scopeIndex
	line, col    int
	isIdentifier bool
	level        int
}

func (d scopeDecl) String(src *s.Source, ast *s.AST) string {
	n := ast.GetNode(d.node)
	if n.Tag() == s.NodeSource {
		return "SRC"
	}
	if n.Tag() == s.NodeFunctionDecl {
		return "FN"
	}
	if n.Tag() == s.NodeBlock {
		return "BLK"
	}
	t := n.Token()
	lexeme := src.Lexeme(t)
	return fmt.Sprintf("%s[%d]", lexeme, t)
}

type scopeEnv struct {
	ast       *s.AST
	allDecls  []scopeDecl
	seenDecls []scopeDecl
	scopeBase int
	scopeTop  int
	curLevel  int
}

func newScopeEnv(ast *s.AST) scopeEnv {
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

func (e scopeEnv) get(i scopeIndex) scopeDecl {
	return e.allDecls[i]
}

func (e scopeEnv) lookup(node s.NodeIndex) bool {
	for i := e.scopeTop - 1; i >= 0; i-- {
		d := e.seenDecls[i]
		lhs := s.Identifier_String(*e.ast, node)
		rhs := s.Identifier_String(*e.ast, d.node)
		if lhs == rhs {
			return true
		}
	}
	return false
}

func (e scopeEnv) uniqueName(i scopeIndex) UniqueName {
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

	curParent      scopeIndex
	inUsageContext bool
}

type ScopeCheckResult struct {
	Ast         *s.AST
	UniqueNames map[s.NodeIndex]UniqueName
}

func ScopecheckPass(src *s.Source, ast *s.AST, handler *u.ErrorHandler) ScopeCheckResult {
	ctx := scopecheckContext{
		env:       newScopeEnv(ast),
		curParent: scopeTop,
	}

	addDecl := func(i s.NodeIndex, isIdentifier bool) scopeIndex {
		line, col := src.Location(ast.GetNode(i).Token())
		ctx.env.add(scopeDecl{
			parent:       ctx.curParent,
			node:         i,
			line:         line,
			col:          col,
			isIdentifier: isIdentifier,
			level:        ctx.env.curLevel,
		})
		return scopeIndex(len(ctx.env.allDecls) - 1)
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

	onEnter := func(ast *s.AST, i s.NodeIndex) (shouldStop bool) {
		n := ast.GetNode(i)
		switch n.Tag() {
		case s.NodeSource:
			ctx.env.enterScope()
			ctx.curParent = addDecl(i, false)
			dump(ctx.env.allDecls, " | ")

		case s.NodeFunctionDecl:
			ctx.env.enterScope()
			ctx.curParent = addDecl(i, false)
			dump(ctx.env.allDecls, " | ")

		case s.NodeBlock:
			ctx.env.enterScope()
			ctx.curParent = addDecl(i, false)
			dump(ctx.env.allDecls, " | ")

		case s.NodeExpression:
			ctx.inUsageContext = true

		case s.NodeIdentifier:
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

	onExit := func(ast *s.AST, i s.NodeIndex) (shouldStop bool) {
		n := ast.GetNode(i)
		switch n.Tag() {
		case s.NodeSource:
			dump(ctx.env.allDecls, " | ")
			ctx.env.exitScope()

		case s.NodeFunctionDecl:
			dump(ctx.env.allDecls, " | ")
			ctx.curParent = ctx.env.get(ctx.curParent).parent
			ctx.env.exitScope()

		case s.NodeBlock:
			dump(ctx.env.allDecls, " | ")
			ctx.curParent = ctx.env.get(ctx.curParent).parent
			ctx.env.exitScope()

		case s.NodeExpression:
			ctx.inUsageContext = false
		}
		return
	}

	ast.Traverse(onEnter, onExit)

	uniqueNames := make(map[s.NodeIndex]UniqueName)
	for i, d := range ctx.env.allDecls {
		if d.isIdentifier {
			uniqueNames[d.node] = ctx.env.uniqueName(scopeIndex(i))
		}
	}
	return ScopeCheckResult{
		Ast:         ast,
		UniqueNames: uniqueNames,
	}
}
