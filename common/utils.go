package common

import (
	"math/big"
	"sort"
)

type Float64Sort struct {
	name  string
	value float64
}

func Float64SorterProcess(planets []Float64Sort) {
	bi := &Float64Sorter{
		data: planets,
	}
	sort.Sort(bi)
}

type Float64Sorter struct {
	data []Float64Sort
	by   func(p1, p2 *Float64Sort) bool
}

func (s *Float64Sorter) Len() int {
	return len(s.data)
}

func (s *Float64Sorter) Swap(i, j int) {
	s.data[i], s.data[j] = s.data[j], s.data[i]
}

func (s *Float64Sorter) Less(i, j int) bool {
	return s.data[i].value > s.data[j].value
}

// uint64
type Uint64Sort struct {
	Name  string
	Value uint64
}

func Uint64SorterProcess(planets []Uint64Sort) {
	bi := &Uint64Sorter{
		data: planets,
	}
	sort.Sort(bi)
}

type Uint64Sorter struct {
	data []Uint64Sort
	by   func(p1, p2 *Uint64Sort) bool
}

func (s *Uint64Sorter) Len() int {
	return len(s.data)
}

func (s *Uint64Sorter) Swap(i, j int) {
	s.data[i], s.data[j] = s.data[j], s.data[i]
}

func (s *Uint64Sorter) Less(i, j int) bool {
	return s.data[i].Value > s.data[j].Value
}

type BigIntSort struct {
	Name  string
	Value *big.Int
}

// type ByBigInt func(p1, p2 *bigIntSort) bool

func BigIntSorterProcess(planets []BigIntSort) {
	bi := &BigIntSorter{
		data: planets,
	}
	sort.Sort(bi)
}

type BigIntSorter struct {
	data []BigIntSort
}

func (s *BigIntSorter) Len() int {
	return len(s.data)
}

func (s *BigIntSorter) Swap(i, j int) {
	s.data[i], s.data[j] = s.data[j], s.data[i]
}

func (s *BigIntSorter) Less(i, j int) bool {
	return s.data[i].Value.Cmp(s.data[j].Value) > 0
}
