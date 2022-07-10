package store

import "math"

var Hull8BoundVectors = []Vector{
	{1, 0},
	{-1, 0},
	{0, 1},
	{0, -1},
	{1, 1},
	{-1, -1},
	{1, -1},
	{-1, 1},
}

type Hull8 struct {
	Bounds [8]float64
}

func NewHull8FromVector(v Vector) *Hull8 {
	h := new(Hull8)

	for i := 0; i < 8; i++ {
		h.Bounds[i] = v.Dot(Hull8BoundVectors[i])
	}

	return h
}

func (h Hull8) Size() float64 {
	return 0
}

func (h Hull8) Add(h2 Hull8) Hull8 {

	for i := 0; i < 8; i++ {
		h.Bounds[i] = math.Max(h.Bounds[i], h2.Bounds[i])
	}
	return h
}
