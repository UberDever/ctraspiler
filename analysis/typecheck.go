package analysis

import (
	a "some/ast"
	ID "some/domain"
	s "some/syntax"
	T "some/typesystem"
	u "some/util"
)

// NOTE: This code is an example of bad non-exaustive switches
// they are exaustive, but only at runtime - this is just garbage

type typeCheckContext struct {
	repo                T.TypeRepo
	evaluationStack     []ID.Type
	seenIdentifierTypes map[string]ID.Type
	unificationSet      u.DisjointSet
}

func newTypeCheckContext() typeCheckContext {
	return typeCheckContext{
		repo:                T.NewTypeRepo(),
		evaluationStack:     make([]ID.Type, 0, 64),
		seenIdentifierTypes: make(map[string]ID.Type),
		unificationSet:      u.NewDisjointSet(),
	}
}

func (c *typeCheckContext) makeSet(id ID.Type) {
	if id < 0 {
		panic("Something went horribly wrong")
	}
	c.unificationSet.MakeSet(uint(id))
}

func (c *typeCheckContext) pushType(t ID.Type) {
	c.evaluationStack = append(c.evaluationStack, t)
}

func (c *typeCheckContext) popType() ID.Type {
	t := c.evaluationStack[len(c.evaluationStack)-1]
	c.evaluationStack = c.evaluationStack[:len(c.evaluationStack)-1]
	return t
}

func (c typeCheckContext) result(ast *a.AST) a.TypedAST {
	repo := T.NewTypeRepo()
	for id := 0; id < c.repo.Count(); id++ {
		originalT := c.repo.GetType(ID.Type(id))
		// if originalT.Node == ID.NodeInvalid {
		// 	continue
		// }
		actualID := ID.Type(c.unificationSet.Find(uint(id)))
		actualT := c.repo.GetType(actualID)
		subtypes := make([]ID.Type, 0, 2)
		it := c.repo.Subtypes(actualID)

		for !it.Done() {
			subtypes = append(subtypes, it.Next())
		}
		repo.AddType(originalT.Node, actualT.Kind, subtypes[0], subtypes[1:]...)
	}

	return a.NewTypedAST(ast, repo)
}

func (c *typeCheckContext) unify(id1, id2 ID.Type) bool {
	i1 := ID.Type(c.unificationSet.Find(uint(id1)))
	i2 := ID.Type(c.unificationSet.Find(uint(id2)))
	if i1 != i2 {
		isTypeVar1 := c.repo.IsTypeVariable(i1)
		isTypeVar2 := c.repo.IsTypeVariable(i2)
		if isTypeVar1 && isTypeVar2 {
			c.unificationSet.Union(uint(i1), uint(i2))
		} else if isTypeVar1 && !isTypeVar2 {
			c.unificationSet.Union(uint(i1), uint(i2))
		} else if !isTypeVar1 && isTypeVar2 {
			c.unificationSet.Union(uint(i2), uint(i1))
		} else if c.repo.SameKind(i1, i2) {
			c.unificationSet.Union(uint(i1), uint(i2))
			types1 := c.repo.Subtypes(i1)
			types2 := c.repo.Subtypes(i2)
			for {
				sub1 := types1.Next()
				sub2 := types2.Next()
				if (types1.Done() && !types2.Done()) ||
					(!types1.Done() && types2.Done()) {
					return false
				}
				if types1.Done() || types2.Done() {
					break
				}
				if !c.unify(sub1, sub2) {
					return false
				}
			}
		} else /* Type inference failed */ {
			return false
		}
	}
	return true
}

// NOTE: Could have been using attributed grammar framework here
func TypeCheckPass(scopeCheckResult ScopeCheckResult, src *s.Source, ast *a.AST, handler *u.ErrorHandler) a.TypedAST {
	ctx := newTypeCheckContext()
	qualifiedNames := scopeCheckResult.QualifiedNames

	addSimpleType := func(node ID.Node, id ID.Type) ID.Type {
		t := ctx.repo.AddType(node, ID.KindIdentity, id)
		ctx.makeSet(t)
		return t
	}

	addFunctionType := func(node ID.Node, id ...ID.Type) ID.Type {
		if len(id) < 1 {
			panic("Something went horribly wrong")
		}
		t := ctx.repo.AddType(node, ID.KindIdentity, id[0], id[1:]...)
		ctx.makeSet(t)
		return t
	}
	_ = addFunctionType

	tryUnify := func(node ID.Node, t1, t2 ID.Type) bool {
		result := ctx.unify(t1, t2)
		if !result {
			line, col := src.Location(ast.GetNode(node).Token())
			s1 := ctx.repo.GetString(t1)
			s2 := ctx.repo.GetString(t2)
			handler.Add(
				u.NewError(u.Semantic,
					u.ES_TypeinferenceFailed,
					line,
					col,
					src.Filename(),
					s1,
					s2,
					node))
		}
		return result
	}

	onEnter := func(ast *a.AST, id ID.Node) (shouldStop bool) {
		n := ast.GetNode(id)

		switch n.Tag() {
		case ID.NodeVarDecl:
			fallthrough
		case ID.NodeConstDecl:
			fallthrough
		case ID.NodeAssignment:
			lhsT := ctx.popType()
			rhsT := ctx.popType()
			if !tryUnify(id, rhsT, lhsT) {
				return
			}
			// no push type - statement
		case ID.NodeOr:
			fallthrough
		case ID.NodeAnd:
			fallthrough
		case ID.NodeEquals:
			fallthrough
		case ID.NodeNotEquals:
			fallthrough
		case ID.NodeGreaterThan:
			fallthrough
		case ID.NodeLessThan:
			fallthrough
		case ID.NodeGreaterThanEquals:
			fallthrough
		case ID.NodeLessThanEquals:
			lhsT := ctx.popType()
			rhsT := ctx.popType()
			v := addSimpleType(id, ID.TypeVar)
			t := addSimpleType(ID.NodeInvalid, ID.TypeBool)
			if !tryUnify(id, lhsT, rhsT) {
				return
			}
			if !tryUnify(id, lhsT, v) {
				return
			}
			if !tryUnify(id, lhsT, t) {
				return
			}
			ctx.pushType(v)
		case ID.NodeBinaryMinus:
			fallthrough
		case ID.NodeMultiply:
			fallthrough
		case ID.NodeDivide:
			fallthrough
		// TODO: In this case binary plus `adds` all types together, however, I need only
		// int, float, string
		case ID.NodeBinaryPlus:
			lhsT := ctx.popType()
			rhsT := ctx.popType()
			v := addSimpleType(id, ID.TypeVar)
			if !tryUnify(id, lhsT, rhsT) {
				return
			}
			if !tryUnify(id, lhsT, v) {
				return
			}
			ctx.pushType(v)
		case ID.NodeUnaryPlus:
			fallthrough
		case ID.NodeUnaryMinus:
			v := ctx.popType()
			t := addSimpleType(ID.NodeInvalid, ID.TypeInt)
			if !tryUnify(id, t, v) {
				return
			}
			ctx.pushType(v)
		case ID.NodeNot:
			v := ctx.popType()
			t := addSimpleType(ID.NodeInvalid, ID.TypeBool)
			if !tryUnify(id, t, v) {
				return
			}
			ctx.pushType(v)
		case ID.NodeIdentifier:
			name, has := qualifiedNames.GetNodeName(id)
			if !has {
				panic("Something went horribly wrong")
			}
			seenT, seen := ctx.seenIdentifierTypes[string(name)]
			var v ID.Type
			if !seen {
				v = addSimpleType(id, ID.TypeVar)
				ctx.seenIdentifierTypes[string(name)] = v
			} else {
				v = addSimpleType(id, ID.TypeVar)
				if !tryUnify(id, seenT, v) {
					return
				}
			}
			ctx.pushType(v)
		case ID.NodeIntLiteral:
			v := addSimpleType(id, ID.TypeVar)
			t := addSimpleType(ID.NodeInvalid, ID.TypeInt)
			if !tryUnify(id, t, v) {
				return
			}
			ctx.pushType(v)
		case ID.NodeFloatLiteral:
			v := addSimpleType(id, ID.TypeVar)
			t := addSimpleType(ID.NodeInvalid, ID.TypeFloat)
			if !tryUnify(id, t, v) {
				return
			}
			ctx.pushType(v)
		case ID.NodeBoolLiteral:
			v := addSimpleType(id, ID.TypeVar)
			t := addSimpleType(ID.NodeInvalid, ID.TypeBool)
			if !tryUnify(id, t, v) {
				return
			}
			ctx.pushType(v)
		}
		return
	}

	onExit := func(ast *a.AST, i ID.Node) (shouldStop bool) {
		return
	}

	ast.TraversePostorder(onEnter, onExit)
	return ctx.result(ast)
}
