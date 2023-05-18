package semantics

import (
	"math"
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

const typeIndexInvalid typeIndex = math.MinInt

type nodeType struct {
	node     s.NodeIndex
	kind     typeKind
	lhs, rhs typeIndex // TODO: range
}

func (t nodeType) isTypeVariable(repo *typeRepo) bool {
	switch t.kind {
	case KindId:
		return t.lhs >= 0
	case KindPtr:
		ptrType := repo.get(t.lhs)
		return ptrType.isTypeVariable(repo)
	case KindFunction:
		funcType := repo.get(t.lhs)
		isTypeVariable := true
		for i := funcType.lhs; i < funcType.rhs; i++ {
			paramType := repo.get(repo.getExtra(i))
			isTypeVariable = isTypeVariable && paramType.isTypeVariable(repo)
		}
	default:
		panic("Something went wrong")
	}
	panic("Something went wrong")
}

func (t nodeType) isProperType(repo *typeRepo) bool {
	return !t.isTypeVariable(repo)
}

type typeRepo struct {
	nodeTypes []nodeType
	extra     []typeIndex
}

func (c typeRepo) get(i typeIndex) nodeType {
	return c.nodeTypes[i]
}

type typeIterator struct {
	nodeType
	curExtra typeIndex
}

func newTypeIterator(t nodeType) typeIterator {
	i := typeIterator{t, typeIndexInvalid}
	switch t.kind {
	case KindPtr:
	case KindFunction:
		i.curExtra = i.lhs
	}
	return i
}

func (i typeIterator) next() typeIndex {
	e := i.curExtra
	switch i.kind {
	case KindPtr:
		i.curExtra = typeIndexInvalid
	case KindFunction:
		if i.curExtra >= i.rhs {
			return typeIndexInvalid
		}
		i.curExtra++
		return e
	}
	return i.curExtra
}

func (c typeRepo) getExtra(i typeIndex) typeIndex {
	return c.extra[i]
}

type typeCheckContext struct {
	repo           typeRepo
	unificationSet u.DisjointSet
}

// type Id struct {
// 	node s.NodeIndex
// 	t    typeIndex
// }

// func (c typeCheckContext) Id(t nodeType) Id {
// 	return Id{
// 		node: t.node,
// 		t:    t.lhs,
// 	}
// }

func (c *typeCheckContext) unify(index1, index2 typeIndex) bool {
	i1 := typeIndex(c.unificationSet.Find(uint(index1)))
	i2 := typeIndex(c.unificationSet.Find(uint(index2)))
	if i1 != i2 {
		type1 := c.repo.get(i1)
		type2 := c.repo.get(i2)
		isTypeVar1 := type1.isTypeVariable(&c.repo)
		isTypeVar2 := type2.isTypeVariable(&c.repo)
		if isTypeVar1 && isTypeVar2 {
			c.unificationSet.Union(uint(i1), uint(i2))
		} else if isTypeVar1 && !isTypeVar2 {
			c.unificationSet.Union(uint(i1), uint(i2))
		} else if !isTypeVar1 && isTypeVar2 {
			c.unificationSet.Union(uint(i2), uint(i1))
		} else if type1.kind == type2.kind {
			c.unificationSet.Union(uint(i1), uint(i2))

		}

	}
	return true
}

type TypeCheckResult struct {
}

func TypeCheckPass(result ScopeCheckResult, src *s.Source, ast *s.AST, handler *u.ErrorHandler) {
	ctx := typeCheckContext{
		repo: typeRepo{
			nodeTypes: make([]nodeType, 0, 64),
			extra:     make([]typeIndex, 0, 64),
		},
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
