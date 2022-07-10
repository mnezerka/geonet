package store

import (
	"math"

	"github.com/apex/log"
)

var Hull4BoundVectors = []Vector{
	{1, 0},
	{-1, 0},
	{0, 1},
	{0, -1},
}

type Hull4 struct {
	Bounds [4]float64
}

func NewHull4FromVector(v Vector) *Hull4 {
	h := new(Hull4)

	for i := 0; i < 4; i++ {
		h.Bounds[i] = v.Dot(Hull4BoundVectors[i])
	}

	return h
}

func (h Hull4) Size() float64 {
	v := make([]Vector, len(Hull4BoundVectors))
	copy(v, Hull4BoundVectors)

	for i := 0; i < 4; i++ {
		v[i] = v[i].Scalar(h.Bounds[i])
	}

	return v[0].X - v[1].X + v[2].Y - v[3].Y
}

func (h Hull4) Add(h2 Hull4) Hull4 {

	for i := 0; i < 4; i++ {
		h.Bounds[i] = math.Max(h.Bounds[i], h2.Bounds[i])
	}
	return h
}

func (h Hull4) BoundRect() Rect {
	v := make([]Vector, len(Hull4BoundVectors))
	copy(v, Hull4BoundVectors)

	log.Infof("hull %v", h)
	log.Infof("hull bound vectors %v", v)

	for i := 0; i < 4; i++ {
		v[i] = v[i].Scalar(h.Bounds[i])
	}

	log.Infof("hull vectors %v", v)

	return Rect{Point{v[0].X, v[2].Y}, Point{v[1].X, v[3].Y}}
}