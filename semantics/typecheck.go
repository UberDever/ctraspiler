package semantics

import (
	"math"
	s "some/syntax"
	u "some/util"
)

// NOTE: This code is an example of bad non-exaustive switches
// they are exaustive, but only at runtime - this is just garbage

type typeKind int

const (
	KindId typeKind = iota
	KindPtr
	KindFunction
)

type typeID int

const typeIDInvalid typeID = math.MinInt
const typeIDTypeVar typeID = math.MaxInt

const (
	TypeIDInt typeID = -iota - 1
	TypeIDFloat
	TypeIDString
)

type nodeType struct {
	node     s.NodeID
	kind     typeKind
	lhs, rhs typeID
}

type typeRepo struct {
	nodeTypes []nodeType
	extraData []typeID
}

func newTypeRepo() typeRepo {
	r := typeRepo{
		nodeTypes: make([]nodeType, 0, 64),
		extraData: make([]typeID, 0, 64),
	}
	return r
}

func (r *typeRepo) addSimpleType(node s.NodeID, id typeID) typeID {
	t := nodeType{
		node: node,
		kind: KindId,
		lhs:  id,
		rhs:  typeIDInvalid,
	}
	r.nodeTypes = append(r.nodeTypes, t)
	return typeID(len(r.nodeTypes) - 1)
}

func (r *typeRepo) addPtrType(node s.NodeID, id typeID) typeID {
	t := nodeType{
		node: node,
		kind: KindPtr,
		lhs:  id,
		rhs:  typeIDInvalid,
	}
	r.nodeTypes = append(r.nodeTypes, t)
	return typeID(len(r.nodeTypes) - 1)
}

func (r *typeRepo) addFunctionType(node s.NodeID, ids ...typeID) typeID {
	r.extraData = append(r.extraData, ids...)
	lhs := len(r.extraData) - len(ids)
	rhs := len(r.extraData)
	t := nodeType{
		node: node,
		kind: KindPtr,
		lhs:  typeID(lhs),
		rhs:  typeID(rhs),
	}
	r.nodeTypes = append(r.nodeTypes, t)
	return typeID(len(r.nodeTypes) - 1)
}

func (r typeRepo) getType(id typeID) nodeType {
	return r.nodeTypes[id]
}

func (r typeRepo) subtypes(id typeID) typeIterator {
	return newTypeIterator(r.getType(id))
}

func (r typeRepo) isTypeVariable(id typeID) bool {
	t := r.getType(id)
	switch t.kind {
	case KindId:
		return t.lhs >= 0
	case KindPtr:
		fallthrough
	case KindFunction:
		return false
	default:
		panic("this switch should be exaustive")
	}
}

func (r typeRepo) isProperType(id typeID) bool {
	return !r.isTypeVariable(id)
}

func (r typeRepo) sameKind(id1, id2 typeID) bool {
	t1 := r.getType(id1)
	t2 := r.getType(id2)
	if t1.kind == KindId && t2.kind == KindId {
		return true
	} else if t1.kind == KindPtr && t2.kind == KindPtr {
		return true
	} else if t1.kind == KindFunction && t2.kind == KindFunction {
		argCount1 := r.subtypes(id1).count()
		argCount2 := r.subtypes(id2).count()
		return argCount1 == argCount2
	}
	return false
}

type typeIterator struct {
	nodeType
	curExtra typeID
}

func newTypeIterator(t nodeType) typeIterator {
	it := typeIterator{t, typeIDInvalid}
	switch t.kind {
	case KindId:
		fallthrough
	case KindPtr:
		fallthrough
	case KindFunction:
		it.curExtra = it.lhs
	default:
		panic("this switch should be exaustive")
	}
	return it
}

func (i typeIterator) done() bool {
	return i.curExtra == typeIDInvalid
}

func (i typeIterator) count() int {
	switch i.kind {
	case KindId:
		return 1
	case KindPtr:
		return 1
	case KindFunction:
		return int(i.rhs) - int(i.lhs) + 1
	default:
		panic("this switch should be exaustive")
	}
}

func (i *typeIterator) next() typeID {
	e := i.curExtra
	switch i.kind {
	case KindId:
		fallthrough
	case KindPtr:
		i.curExtra = typeIDInvalid
	case KindFunction:
		if i.curExtra >= i.rhs {
			return typeIDInvalid
		}
		i.curExtra++
	default:
		panic("this switch should be exaustive")
	}
	return e
}

type typeCheckContext struct {
	repo           typeRepo
	seenNodeTypes  map[s.NodeID]typeID
	unificationSet u.DisjointSet
}

func newTypeCheckContext() typeCheckContext {
	return typeCheckContext{
		repo:           newTypeRepo(),
		seenNodeTypes:  make(map[s.NodeID]typeID),
		unificationSet: u.NewDisjointSet(),
	}
}

func (c *typeCheckContext) makeSet(id typeID) {
	if id < 0 {
		panic("Something went horribly wrong")
	}
	c.unificationSet.MakeSet(uint(id))
}

func (c *typeCheckContext) unify(id1, id2 typeID) bool {
	i1 := typeID(c.unificationSet.Find(uint(id1)))
	i2 := typeID(c.unificationSet.Find(uint(id2)))
	if i1 != i2 {
		isTypeVar1 := c.repo.isTypeVariable(i1)
		isTypeVar2 := c.repo.isTypeVariable(i2)
		if isTypeVar1 && isTypeVar2 {
			c.unificationSet.Union(uint(i1), uint(i2))
		} else if isTypeVar1 && !isTypeVar2 {
			c.unificationSet.Union(uint(i1), uint(i2))
		} else if !isTypeVar1 && isTypeVar2 {
			c.unificationSet.Union(uint(i2), uint(i1))
		} else if c.repo.sameKind(i1, i2) {
			c.unificationSet.Union(uint(i1), uint(i2))
			types1 := c.repo.subtypes(i1)
			types2 := c.repo.subtypes(i2)
			for {
				sub1 := types1.next()
				sub2 := types2.next()
				if (types1.done() && !types2.done()) ||
					(!types1.done() && types2.done()) {
					return false
				}
				if types1.done() || types2.done() {
					break
				}
				c.unify(sub1, sub2)
			}
		}
	}
	return true
}

type TypeCheckResult struct {
}

func TypeCheckPass(result ScopeCheckResult, src *s.Source, ast *s.AST, handler *u.ErrorHandler) {
	ctx := newTypeCheckContext()

	onEnter := func(ast *s.AST, i s.NodeID) (shouldStop bool) {
		n := ast.GetNode(i)
		var v, t typeID
		switch n.Tag() {
		case s.NodeIntLiteral:
			ctx.makeSet(t)
			ctx.makeSet(v)
			v = ctx.repo.addSimpleType(i, typeIDTypeVar)
			t = ctx.repo.addSimpleType(s.NodeIDInvalid, TypeIDInt)
		case s.NodeBinaryPlus:
			// TODO: here we have situations like this:
			// a = 2 + 3
			// a = b + 3
			// first case => do all `addSimpleType` and stuff
			// second case => find type for b that already have been present
			// possible solutions:
			// 1. Rewrite it in recursive descend manner
			// 2. Reread Engineering a compiler on attribute grammar (type inference example)
		}
		ctx.unify(t, v)
		return
	}

	onExit := func(ast *s.AST, i s.NodeID) (shouldStop bool) {
		return
	}

	ast.Traverse(onEnter, onExit)
}
