package util

import (
	"reflect"
	"testing"
)

func TestDisjointSet(t *testing.T) {
	expected := [][]int{
		{1, 2, 3, 4, 5},
		{1, 2, 3, 3, 5},
		{1, 1, 3, 3, 5},
		{3, 3, 3, 3, 5},
	}

	check := func(step int, items []int, s DisjointSet) (actual []int, eq bool) {
		actual = make([]int, 0, 64)
		for _, n := range items {
			x := s.Find(uint(n))
			actual = append(actual, int(x))
		}
		eq = reflect.DeepEqual(actual, expected[step])
		return
	}

	items := []int{1, 2, 3, 4, 5}
	s := NewDisjointSet()
	step := 0

	for _, i := range items {
		s.MakeSet(uint(i))
	}
	if actual, ok := check(step, items, s); !ok {
		t.Errorf("Sets are not equal %v %v", actual, expected[step])
	}
	step++

	s.Union(4, 3)
	if actual, ok := check(step, items, s); !ok {
		t.Errorf("Sets are not equal %v %v", actual, expected[step])
	}
	step++

	s.Union(2, 1)
	if actual, ok := check(step, items, s); !ok {
		t.Errorf("Sets are not equal %v %v", actual, expected[step])
	}
	step++

	s.Union(1, 3)
	if actual, ok := check(step, items, s); !ok {
		t.Errorf("Sets are not equal %v %v", actual, expected[step])
	}
	step++
}
