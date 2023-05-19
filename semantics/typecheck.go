package semantics

import (
	"math"
	a "some/ast"
	s "some/syntax"
	u "some/util"
	"strings"
)

// NOTE: This code is an example of bad non-exaustive switches
// they are exaustive, but only at runtime - this is just garbage

type typeKind int

const (
	KindId typeKind = iota
	KindPtr
	KindFunction
)

type TypeID int

const typeIDInvalid TypeID = math.MinInt
const typeIDTypeVar TypeID = math.MaxInt

const (
	TypeIDInt TypeID = -iota - 1
	TypeIDFloat
	TypeIDString
)

type nodeType struct {
	node     a.NodeID
	kind     typeKind
	lhs, rhs TypeID
}

type typeRepo struct {
	nodeTypes []nodeType
	extraData []TypeID
}

func newTypeRepo() typeRepo {
	r := typeRepo{
		nodeTypes: make([]nodeType, 0, 64),
		extraData: make([]TypeID, 0, 64),
	}
	return r
}

func (r *typeRepo) addType(node a.NodeID, kind typeKind, subtype TypeID, rest ...TypeID) TypeID {
	lhs := subtype
	rhs := typeIDInvalid
	if len(rest) == 1 {
		rhs = rest[0]
	} else if len(rest) > 1 {
		r.extraData = append(r.extraData, rest...)
		lhs = TypeID(len(r.extraData) - len(rest))
		rhs = TypeID(len(r.extraData))
	}
	t := nodeType{
		node: node,
		kind: kind,
		lhs:  lhs,
		rhs:  rhs,
	}
	r.nodeTypes = append(r.nodeTypes, t)
	return TypeID(len(r.nodeTypes) - 1)
}

func (r typeRepo) getType(id TypeID) nodeType {
	return r.nodeTypes[id]
}

func (r typeRepo) subtypes(id TypeID) typeIterator {
	return newTypeIterator(r.getType(id))
}

func (r typeRepo) isTypeVariable(id TypeID) bool {
	if id == typeIDTypeVar {
		return true
	}
	if id == typeIDInvalid {
		panic("Something went horribly wrong")
	}
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

func (r typeRepo) isProperType(id TypeID) bool {
	return !r.isTypeVariable(id)
}

func (r typeRepo) sameKind(id1, id2 TypeID) bool {
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

func (r typeRepo) getString(id TypeID) (s string) {
	t := r.getType(id)

	typeString := func(id TypeID) string {
		if r.isProperType(id) {
			switch t.lhs {
			case TypeIDInt:
				return "int "
			default:
				panic("this switch should be exaustive")
			}
		} else {
			return "V "
		}
	}

	switch t.kind {
	case KindId:
		s += typeString(t.lhs)
	case KindPtr:
		s += "(^ "
		s += r.getString(t.lhs)
		s += ")"
	case KindFunction:
		s += "(FN "
		subtypes := r.subtypes(id)
		for {
			if subtypes.done() {
				break
			}
			sub := r.getString(subtypes.next())
			s += sub
		}
		s += ")"
	default:
		panic("this switch should be exaustive")
	}
	return
}

type typeIterator struct {
	nodeType
	curExtra TypeID
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

func (i *typeIterator) next() TypeID {
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
	repo                typeRepo
	evaluationStack     []TypeID
	seenIdentifierTypes map[string]TypeID
	unificationSet      u.DisjointSet
}

func newTypeCheckContext() typeCheckContext {
	return typeCheckContext{
		repo:                newTypeRepo(),
		evaluationStack:     make([]TypeID, 0, 64),
		seenIdentifierTypes: make(map[string]TypeID),
		unificationSet:      u.NewDisjointSet(),
	}
}

func (c *typeCheckContext) makeSet(id TypeID) {
	if id < 0 {
		panic("Something went horribly wrong")
	}
	c.unificationSet.MakeSet(uint(id))
}

func (c *typeCheckContext) pushType(t TypeID) {
	c.evaluationStack = append(c.evaluationStack, t)
}

func (c *typeCheckContext) popType() TypeID {
	t := c.evaluationStack[len(c.evaluationStack)-1]
	c.evaluationStack = c.evaluationStack[:len(c.evaluationStack)-1]
	return t
}

func (c *typeCheckContext) unify(id1, id2 TypeID) bool {
	i1 := TypeID(c.unificationSet.Find(uint(id1)))
	i2 := TypeID(c.unificationSet.Find(uint(id2)))
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

type TypedAST struct {
	a.AST
	nodeTypes map[a.NodeID]TypeID
	repo      typeRepo
}

func newTypedAST(ast *a.AST, ctx *typeCheckContext) TypedAST {
	tAst := TypedAST{
		AST:       *ast,
		nodeTypes: make(map[a.NodeID]TypeID),
		repo:      newTypeRepo(),
	}
	_ = tAst
	for i := range ctx.repo.nodeTypes {
		id := TypeID(ctx.unificationSet.Find(uint(i)))
		t := ctx.repo.nodeTypes[id]
		tAst.nodeTypes[t.node] = TypeID(id)

		subtypes := make([]TypeID, 0, 2)
		it := ctx.repo.subtypes(id)
		for !it.done() {
			subtypes = append(subtypes, it.next())
		}
		tAst.repo.addType(t.node, t.kind, subtypes[0], subtypes[:1]...)
	}
	return tAst
}

func (ast *TypedAST) Dump() string {
	str := strings.Builder{}
	onEnter := func(_ *a.AST, i a.NodeID) (stopTraversal bool) {
		str.WriteByte('(')
		str.WriteString(ast.GetNodeString(i))

		// filter nodes that are composite by themselves
		n := ast.AST.GetNode(i)
		stopTraversal = n.Tag() == a.NodeIdentifierList
		return
	}
	onExit := func(_ *a.AST, i a.NodeID) (stopTraversal bool) {
		str.WriteByte(')')
		return false
	}
	ast.AST.TraversePreorder(onEnter, onExit)

	return str.String()
}

// NOTE: Could have been using attributed grammar framework here
func TypeCheckPass(scopeCheckResult ScopeCheckResult, src *s.Source, ast *a.AST, handler *u.ErrorHandler) TypedAST {
	ctx := newTypeCheckContext()
	uniqueNames := scopeCheckResult.UniqueNames

	addSimpleType := func(node a.NodeID, id TypeID) TypeID {
		t := ctx.repo.addType(node, KindId, id)
		ctx.makeSet(t)
		return t
	}

	addFunctionType := func(node a.NodeID, id ...TypeID) TypeID {
		if len(id) < 1 {
			panic("Something went horribly wrong")
		}
		t := ctx.repo.addType(node, KindId, id[0], id[1:]...)
		ctx.makeSet(t)
		return t
	}
	_ = addFunctionType

	onEnter := func(ast *a.AST, i a.NodeID) (shouldStop bool) {
		n := ast.GetNode(i)

		switch n.Tag() {
		case a.NodeIntLiteral:
			v := addSimpleType(i, typeIDTypeVar)
			t := addSimpleType(a.NodeIDInvalid, TypeIDInt)
			ctx.unify(t, v)
			ctx.pushType(v)
		case a.NodeIdentifier:
			id, has := uniqueNames[i]
			if !has {
				panic("Something went horribly wrong")
			}
			seenT, seen := ctx.seenIdentifierTypes[string(id)]
			var v TypeID
			if !seen {
				v = addSimpleType(i, typeIDTypeVar)
				ctx.seenIdentifierTypes[string(id)] = v
			} else {
				v = seenT
			}
			ctx.pushType(v)
		case a.NodeBinaryPlus:
			lhsT := ctx.popType()
			rhsT := ctx.popType()
			v := addSimpleType(i, typeIDTypeVar)
			ctx.unify(lhsT, rhsT)
			ctx.unify(lhsT, v)
			ctx.pushType(v)
		case a.NodeAssignment:
			v := ctx.popType()
			t := ctx.popType()
			ctx.unify(t, v)
			// no push type - statement
		}
		return
	}

	onExit := func(ast *a.AST, i a.NodeID) (shouldStop bool) {
		return
	}

	ast.TraversePostorder(onEnter, onExit)

	return newTypedAST(ast, &ctx)
}
