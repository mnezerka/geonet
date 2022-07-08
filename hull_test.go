package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewHull(t *testing.T) {
	result := NewHullFromVector(Vector{2, 7})
	assert.Equal(t, Hull{[8]float64{2, -2, 7, -7, 9, -9, -5, 5}}, *result, "should add correctly")
}

func TestHullAdd(t *testing.T) {
	h1 := NewHullFromVector(Vector{2, 7})
	h2 := NewHullFromVector(Vector{5, 4})

	h := h1.Add(*h2)

	assert.Equal(t, Hull{[8]float64{5, -2, 7, -4, 9, -9, 1, 5}}, h, "should add correctly")

}
