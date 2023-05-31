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

// TODO: add stack for returns, since it is control structure
type typeCheckContext struct {
	repo                T.TypeRepo
	evaluationStack     u.Stack[ID.Type]
	returnStack         u.Stack[ID.Type]
	seenIdentifierTypes map[string]ID.Type
	unificationSet      u.DisjointSet
	inUsageContext      bool
}

func newTypeCheckContext() typeCheckContext {
	return typeCheckContext{
		repo:                T.NewTypeRepo(),
		evaluationStack:     u.NewStack[ID.Type](),
		seenIdentifierTypes: make(map[string]ID.Type),
		unificationSet:      u.NewDisjointSet(),
		inUsageContext:      false,
	}
}

func (c *typeCheckContext) makeSet(id ID.Type) {
	if id < 0 {
		panic("Something went horribly wrong")
	}
	c.unificationSet.MakeSet(uint(id))
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
		repo.AddType(originalT.Node, actualT.Kind, subtypes...)
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
				if types1.Done() && types2.Done() {
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

	addFunctionType := func(node ID.Node, rest ...ID.Type) ID.Type {
		t := ctx.repo.AddType(node, ID.KindFunction, rest...)
		ctx.makeSet(t)
		return t
	}
	_ = addFunctionType

	tryUnify := func(node ID.Node, t1, t2 ID.Type) bool {
		result := ctx.unify(t1, t2)
		if !result {
			assumedT1 := ID.Type(ctx.unificationSet.Find(uint(t1)))
			assumedT2 := ID.Type(ctx.unificationSet.Find(uint(t2)))
			line, col := src.Location(ast.GetNode(node).Token())
			s1 := ctx.repo.GetString(assumedT1)
			s2 := ctx.repo.GetString(assumedT2)
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
		case ID.NodeFunctionDecl:
			// TODO: we should handle returns, parameters and name here
			identifierTs := make([]ID.Type, 0, 16)
			for {
				identifierT, ok := ctx.evaluationStack.Pop()
				if !ok {
					break
				}
				node := ctx.repo.GetType(identifierT).Node
				if ast.GetNode(node).Tag() != ID.NodeIdentifier {
					ctx.evaluationStack.Push(identifierT)
					break
				}
				identifierTs = append(identifierTs, identifierT)
			}
			if len(identifierTs) == 0 {
				panic("Excepted at least function type here")
			}
			paramTs := identifierTs[1:]
			returnT := addSimpleType(ID.NodeInvalid, ID.TypeVar)
			for !ctx.returnStack.IsEmpty() {
				retT, _ := ctx.returnStack.Pop()
				if !tryUnify(id, retT, returnT) {
					return
				}
			}
			signatureT := append(paramTs, returnT)
			fnT := identifierTs[0]
			t := addFunctionType(ctx.repo.GetType(fnT).Node, signatureT...)
			ctx.unify(fnT, t)
			//TODO: remove
			// t := addSimpleType(ID.NodeInvalid, ID.TypeInt)
			// ctx.unify(firstReturnT, t)
		case ID.NodeVarDecl:
			fallthrough
		case ID.NodeConstDecl:
			fallthrough
		case ID.NodeAssignment:
			lhsT, _ := ctx.evaluationStack.Pop()
			rhsT, _ := ctx.evaluationStack.Pop()
			if !tryUnify(id, rhsT, lhsT) {
				return
			}
		case ID.NodeReturnStmt:
			exprT, _ := ctx.evaluationStack.Pop()
			returnT := addSimpleType(id, ID.TypeVar)
			ctx.unify(returnT, exprT)
			ctx.returnStack.Push(returnT)

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
			lhsT, _ := ctx.evaluationStack.Pop()
			rhsT, _ := ctx.evaluationStack.Pop()
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
			ctx.evaluationStack.Push(v)
		case ID.NodeBinaryMinus:
			fallthrough
		case ID.NodeMultiply:
			fallthrough
		case ID.NodeDivide:
			fallthrough
		// TODO: In this case binary plus `adds` all types together, however, I need only
		// int, float, string
		case ID.NodeBinaryPlus:
			lhsT, _ := ctx.evaluationStack.Pop()
			rhsT, _ := ctx.evaluationStack.Pop()
			v := addSimpleType(id, ID.TypeVar)
			if !tryUnify(id, lhsT, rhsT) {
				return
			}
			if !tryUnify(id, lhsT, v) {
				return
			}
			ctx.evaluationStack.Push(v)
		case ID.NodeUnaryPlus:
			fallthrough
		case ID.NodeUnaryMinus:
			v, _ := ctx.evaluationStack.Pop()
			t := addSimpleType(ID.NodeInvalid, ID.TypeInt)
			if !tryUnify(id, t, v) {
				return
			}
			ctx.evaluationStack.Push(v)
		case ID.NodeNot:
			v, _ := ctx.evaluationStack.Pop()
			t := addSimpleType(ID.NodeInvalid, ID.TypeBool)
			if !tryUnify(id, t, v) {
				return
			}
			ctx.evaluationStack.Push(v)
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
				v = seenT
				// v = addSimpleType(id, ID.TypeVar)
				// if !tryUnify(id, seenT, v) {
				// 	return
				// }
			}
			ctx.evaluationStack.Push(v)
		case ID.NodeIntLiteral:
			v := addSimpleType(id, ID.TypeVar)
			t := addSimpleType(ID.NodeInvalid, ID.TypeInt)
			if !tryUnify(id, t, v) {
				return
			}
			ctx.evaluationStack.Push(v)
		case ID.NodeFloatLiteral:
			v := addSimpleType(id, ID.TypeVar)
			t := addSimpleType(ID.NodeInvalid, ID.TypeFloat)
			if !tryUnify(id, t, v) {
				return
			}
			ctx.evaluationStack.Push(v)
		case ID.NodeBoolLiteral:
			v := addSimpleType(id, ID.TypeVar)
			t := addSimpleType(ID.NodeInvalid, ID.TypeBool)
			if !tryUnify(id, t, v) {
				return
			}
			ctx.evaluationStack.Push(v)
		case ID.NodeExpression:
			exprT := addSimpleType(id, ID.TypeVar)
			t, _ := ctx.evaluationStack.Pop()
			ctx.unify(exprT, t)
			ctx.evaluationStack.Push(exprT)
			ctx.inUsageContext = true
		}
		return
	}

	onExit := func(ast *a.AST, id ID.Node) (shouldStop bool) {
		n := ast.GetNode(id)
		switch n.Tag() {
		case ID.NodeExpression:
			ctx.inUsageContext = false
		}
		return
	}

	// main : () -> int
	for i := 0; i < ast.NodeCount(); i++ {
		id := ID.Node(i)
		n := ast.GetNode(id)
		if n.Tag() == ID.NodeFunctionDecl {
			nameId := ast.FunctionDecl(n).Name

			if a.Identifier_String(*ast, nameId) == "main" {
				name, has := qualifiedNames.GetNodeName(nameId)
				if !has {
					panic("Something went horribly wrong")
				}
				intT := addSimpleType(ID.NodeInvalid, ID.TypeInt)
				v := addFunctionType(nameId, intT)
				ctx.seenIdentifierTypes[string(name)] = v
				break
			}
		}
	}

	ast.TraversePostorder(onEnter, onExit)
	return ctx.result(ast)
}
