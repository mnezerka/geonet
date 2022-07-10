package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAdd(t *testing.T) {
	result := Vector{1., 1.}.Add(Vector{2., 2.})
	assert.Equal(t, Vector{3., 3.}, result, "should add correctly")
}

func TestDotProductOfTwoPerpendicularVectors(t *testing.T) {
	result := Vector{1., 0.}.Dot(Vector{0., 1.})
	assert.Equal(t, 0., result, "dot product of two perpendicular vectors is 0")
}

func TestDotProductOfTwoParallelVectors(t *testing.T) {
	result := Vector{1., 0.}.Dot(Vector{1., 0.})
	assert.Equal(t, 1., result, "dot product of two parallel vectors is 1")
}

func TestScalarProduct(t *testing.T) {
	result := Vector{2., 3.}.Scalar(5)
	assert.Equal(t, Vector{10, 15}, result, "dot product of two parallel vectors is 1")
}
