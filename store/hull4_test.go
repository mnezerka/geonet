package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewHull4(t *testing.T) {
	result := NewHull4FromVector(Vector{2, 7})
	assert.Equal(t, Hull4{[4]float64{2, -2, 7, -7}}, *result, "should add correctly")
}

func TestHull4Add(t *testing.T) {
	h1 := NewHull4FromVector(Vector{2, 7})
	h2 := NewHull4FromVector(Vector{5, 4})

	h := h1.Add(*h2)

	assert.Equal(t, Hull4{[4]float64{5, -2, 7, -4}}, h, "should add correctly")
}

func TestHull4Size(t *testing.T) {
	h1 := NewHull4FromVector(Vector{2, 1})
	h2 := NewHull4FromVector(Vector{4, 6})

	h := h1.Add(*h2)

	assert.Equal(t, 7., h.Size(), "should add correctly")
}

func TestHull4BoundRect(t *testing.T) {
	h1 := NewHull4FromVector(Vector{2, 1})
	h2 := NewHull4FromVector(Vector{4, 6})

	h := h1.Add(*h2)

	assert.Equal(t, Rect{Point{4, 6}, Point{2, 1}}, h.BoundRect(), "should add correctly")
}
