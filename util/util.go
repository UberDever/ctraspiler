package util

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
