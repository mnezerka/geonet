package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewHull8(t *testing.T) {
	result := NewHull8FromVector(Vector{2, 7})
	assert.Equal(t, Hull8{[8]float64{2, -2, 7, -7, 9, -9, -5, 5}}, *result, "should add correctly")
}

func TestHull8Add(t *testing.T) {
	h1 := NewHull8FromVector(Vector{2, 7})
	h2 := NewHull8FromVector(Vector{5, 4})

	h := h1.Add(*h2)

	assert.Equal(t, Hull8{[8]float64{5, -2, 7, -4, 9, -9, 1, 5}}, h, "should add correctly")

}
