package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDistance(t *testing.T) {

	d, p := Distance(Point{0, 0}, Point{1, 0}, Point{0, 1})
	assert.Equal(t, 1., d, "should add correctly")
	assert.Equal(t, true, p, "should add correctly")

	// edge case - right point
	d, p = Distance(Point{0, 0}, Point{1, 0}, Point{1, 0})
	assert.Equal(t, 0., d, "should add correctly")
	assert.Equal(t, true, p, "should add correctly")

	// edge case - left point
	d, p = Distance(Point{0, 0}, Point{1, 0}, Point{0, 0})
	assert.Equal(t, 0., d, "should add correctly")
	assert.Equal(t, true, p, "should add correctly")

	// not perpendicular - pythagoras with whole numbers 3, 4, 5
	d, p = Distance(Point{0, 0}, Point{1, 0}, Point{4, 4})
	assert.Equal(t, 5., d, "should add correctly")
	assert.Equal(t, false, p, "should add correctly")
}
