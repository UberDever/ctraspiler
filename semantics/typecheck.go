package semantics

import (
	s "some/syntax"
	u "some/util"
)

type typeKind int

const (
	KindId typeKind = iota
	KindPtr
	KindFunction
)

type typeIndex int

type nodeType struct {
	node     s.NodeIndex
	kind     typeKind
	lhs, rhs typeIndex // TODO: range
}

type typeCheckContext struct {
	nodeTypes      []nodeType
	extra          []typeIndex
	unificationSet u.DisjointSet
}

func TypeCheckPass(src *s.Source, ast *s.AST, handler *u.ErrorHandler) {
	ctx := typeCheckContext{
		nodeTypes:      make([]nodeType, 0, 64),
		extra:          make([]typeIndex, 0, 64),
		unificationSet: u.NewDisjointSet(),
	}
	_ = ctx

	onEnter := func(ast *s.AST, i s.NodeIndex) (shouldStop bool) {
		n := ast.GetNode(i)
		switch n.Tag() {

		}
		return
	}

	onExit := func(ast *s.AST, i s.NodeIndex) (shouldStop bool) {
		return
	}

	ast.Traverse(onEnter, onExit)
}
