package main

import "math"

var BoundVectors = []Vector{
	{1, 0},
	{-1, 0},
	{0, 1},
	{0, -1},
	{1, 1},
	{-1, -1},
	{1, -1},
	{-1, 1},
}

type Hull struct {
	Bounds [8]float64
}

func NewHullFromVector(v Vector) *Hull {
	h := new(Hull)

	for i := 0; i < 8; i++ {
		h.Bounds[i] = v.Dot(BoundVectors[i])
	}

	return h
}

func (h Hull) Size() float64 {
	return 0
}

func (h Hull) Add(h2 Hull) Hull {

	for i := 0; i < 8; i++ {
		h.Bounds[i] = math.Max(h.Bounds[i], h2.Bounds[i])
	}
	return h
}
