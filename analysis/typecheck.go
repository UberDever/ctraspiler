package analysis

import (
	a "some/ast"
	ID "some/domain"
	s "some/syntax"
	. "some/typesystem"
	u "some/util"
)

// NOTE: This code is an example of bad non-exaustive switches
// they are exaustive, but only at runtime - this is just garbage

type typeCheckContext struct {
	repo                TypeRepo
	evaluationStack     []ID.Type
	seenIdentifierTypes map[string]ID.Type
	unificationSet      u.DisjointSet
}

func newTypeCheckContext() typeCheckContext {
	return typeCheckContext{
		repo:                NewTypeRepo(),
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
	repo := NewTypeRepo()
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
				c.unify(sub1, sub2)
			}
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

	onEnter := func(ast *a.AST, id ID.Node) (shouldStop bool) {
		n := ast.GetNode(id)

		switch n.Tag() {
		case ID.NodeIntLiteral:
			v := addSimpleType(id, ID.TypeVar)
			t := addSimpleType(ID.NodeInvalid, ID.TypeInt)
			ctx.unify(t, v)
			ctx.pushType(v)
		case ID.NodeIdentifier:
			name, has := qualifiedNames[id]
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
				ctx.unify(seenT, v)
			}
			ctx.pushType(v)
		case ID.NodeBinaryPlus:
			lhsT := ctx.popType()
			rhsT := ctx.popType()
			v := addSimpleType(id, ID.TypeVar)
			ctx.unify(lhsT, rhsT)
			ctx.unify(lhsT, v)
			ctx.pushType(v)
		case ID.NodeAssignment:
			v := ctx.popType()
			t := ctx.popType()
			ctx.unify(t, v)
			// no push type - statement
		case ID.NodeConstDecl:
			v := ctx.popType()
			t := ctx.popType()
			ctx.unify(t, v)
			// no push type - statement
		}
		return
	}

	onExit := func(ast *a.AST, i ID.Node) (shouldStop bool) {
		return
	}

	ast.TraversePostorder(onEnter, onExit)
	return ctx.result(ast)
}