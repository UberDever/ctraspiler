package typesystem

import (
	ID "some/domain"
)

type nodeType struct {
	Node     ID.Node
	Kind     ID.Kind
	lhs, rhs ID.Type
}

type TypeRepo struct {
	nodeTypes []nodeType
	extraData []ID.Type
}

func NewTypeRepo() TypeRepo {
	r := TypeRepo{
		nodeTypes: make([]nodeType, 0, 64),
		extraData: make([]ID.Type, 0, 64),
	}
	return r
}

func (r *TypeRepo) AddType(node ID.Node, kind ID.Kind, subtype ID.Type, rest ...ID.Type) ID.Type {
	lhs := subtype
	rhs := ID.TypeInvalid
	if len(rest) == 1 {
		rhs = rest[0]
	} else if len(rest) > 1 {
		r.extraData = append(r.extraData, rest...)
		lhs = ID.Type(len(r.extraData) - len(rest))
		rhs = ID.Type(len(r.extraData))
	}
	t := nodeType{
		Node: node,
		Kind: kind,
		lhs:  lhs,
		rhs:  rhs,
	}
	r.nodeTypes = append(r.nodeTypes, t)
	return ID.Type(len(r.nodeTypes) - 1)
}

func (r TypeRepo) GetType(id ID.Type) nodeType {
	return r.nodeTypes[id]
}

func (r TypeRepo) Count() int {
	return len(r.nodeTypes)
}

func (r TypeRepo) Subtypes(id ID.Type) typeIterator {
	return newTypeIterator(r.GetType(id))
}

func (r TypeRepo) IsTypeVariable(id ID.Type) bool {
	if id == ID.TypeVar {
		return true
	}
	if id < 0 {
		return false
	}
	if id == ID.TypeInvalid {
		panic("Something went horribly wrong")
	}
	t := r.GetType(id)
	switch t.Kind {
	case ID.KindIdentity:
		return t.lhs >= 0
	case ID.KindPtr:
		fallthrough
	case ID.KindFunction:
		return false
	default:
		panic("this switch should be exaustive")
	}
}

func (r TypeRepo) SameKind(id1, id2 ID.Type) bool {
	t1 := r.GetType(id1)
	t2 := r.GetType(id2)
	if t1.Kind == ID.KindIdentity && t2.Kind == ID.KindIdentity {
		return true
	} else if t1.Kind == ID.KindPtr && t2.Kind == ID.KindPtr {
		return true
	} else if t1.Kind == ID.KindFunction && t2.Kind == ID.KindFunction {
		argCount1 := r.Subtypes(id1).Count()
		argCount2 := r.Subtypes(id2).Count()
		return argCount1 == argCount2
	}
	return false
}

func (r TypeRepo) GetString(id ID.Type) (s string) {
	t := r.GetType(id)

	typeString := func(id ID.Type) string {
		if !r.IsTypeVariable(id) {
			switch t.lhs {
			case ID.TypeInt:
				return "int"
			default:
				panic("this switch should be exaustive")
			}
		} else {
			return "V"
		}
	}

	switch t.Kind {
	case ID.KindIdentity:
		s += typeString(t.lhs)
	case ID.KindPtr:
		s += "(^ "
		s += r.GetString(t.lhs)
		s += ")"
	case ID.KindFunction:
		s += "(FN "
		subtypes := r.Subtypes(id)
		for {
			if subtypes.Done() {
				break
			}
			sub := r.GetString(subtypes.Next())
			s += sub + " "
		}
		s += ")"
	default:
		panic("this switch should be exaustive")
	}
	return
}

type typeIterator struct {
	nodeType
	curExtra ID.Type
}

func newTypeIterator(t nodeType) typeIterator {
	it := typeIterator{t, ID.TypeInvalid}
	switch t.Kind {
	case ID.KindIdentity:
		fallthrough
	case ID.KindPtr:
		fallthrough
	case ID.KindFunction:
		it.curExtra = it.lhs
	default:
		panic("this switch should be exaustive")
	}
	return it
}

func (i typeIterator) Done() bool {
	return i.curExtra == ID.TypeInvalid
}

func (i typeIterator) Count() int {
	switch i.Kind {
	case ID.KindIdentity:
		return 1
	case ID.KindPtr:
		return 1
	case ID.KindFunction:
		return int(i.rhs) - int(i.lhs) + 1
	default:
		panic("this switch should be exaustive")
	}
}

func (i *typeIterator) Next() ID.Type {
	e := i.curExtra
	switch i.Kind {
	case ID.KindIdentity:
		fallthrough
	case ID.KindPtr:
		i.curExtra = ID.TypeInvalid
	case ID.KindFunction:
		if i.curExtra >= i.rhs {
			return ID.TypeInvalid
		}
		i.curExtra++
	default:
		panic("this switch should be exaustive")
	}
	return e
}
