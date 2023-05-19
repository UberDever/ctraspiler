package util

import "strings"

// I love golang for this stuff
func Min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func Max(x, y int) int {
	if x > y {
		return x
	}
	return y
}

type DisjointSet struct {
	parent map[uint]uint
	rank   map[uint]uint
}

func NewDisjointSet() DisjointSet {
	return DisjointSet{
		parent: make(map[uint]uint),
		rank:   make(map[uint]uint),
	}
}

func (s *DisjointSet) MakeSet(x uint) {
	s.parent[x] = x
	s.rank[x] = 0
}

func (s *DisjointSet) Find(x uint) uint {
	_, ok := s.parent[x]
	if !ok {
		s.MakeSet(x)
	}
	if s.parent[x] != x {
		s.parent[x] = s.Find(s.parent[x])
	}
	return s.parent[x]
}

func (s *DisjointSet) Union(a, b uint) {
	x := s.Find(a)
	y := s.Find(b)

	if x == y {
		return
	}

	if s.rank[x] > s.rank[y] {
		s.parent[y] = x
	} else if s.rank[x] < s.rank[y] {
		s.parent[x] = y
	} else {
		s.parent[x] = y
		s.rank[y] += 1
	}
}

func FormatSExpr(sexpr string) string {
	formatted := strings.Builder{}
	depth := -1
	for i := range sexpr {
		if sexpr[i] == '(' {
			depth++
			formatted.WriteByte('\n')
			for j := 0; j < depth; j++ {
				formatted.WriteString("    ")
			}
			formatted.WriteByte('(')
		} else if sexpr[i] == ')' {
			depth--
			formatted.WriteByte(')')
		} else {
			formatted.WriteByte(sexpr[i])
		}
	}
	return formatted.String()
}

func MinifySExpr(s string) string {
	formatted := strings.Builder{}
	skipWS := func(i int) (int, bool) {
		wasSpace := false
		for s[i] == ' ' || s[i] == '\n' || s[i] == '\t' {
			wasSpace = true
			i++
			if i >= len(s) {
				break
			}
		}
		return i, wasSpace
	}

	for i := 0; i < len(s); i++ {
		j, wasSpace := skipWS(i)
		if j >= len(s) {
			break
		}
		i = j
		if wasSpace {
			if s[i] != '(' && s[i] != ')' {
				formatted.WriteByte(' ')
			}
		}
		formatted.WriteByte(s[i])
	}
	return formatted.String()
}
