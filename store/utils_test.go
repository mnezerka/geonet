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

func TestProjection(t *testing.T) {

	p := Projection(Point{0, 0}, Point{1, 0}, Point{0, 1})
	assert.Equal(t, Point{0, 0}, p, "should add correctly")

	p = Projection(Point{0, 0}, Point{8, 0}, Point{3, 3})
	assert.Equal(t, Point{3, 0}, p, "should add correctly")

	// projection is not on line segment
	p = Projection(Point{3, 0}, Point{8, 0}, Point{2, 2})
	assert.Equal(t, Point{2, 0}, p, "should add correctly")

	p = Projection(Point{0, 0}, Point{1, 1}, Point{0, 2})
	assert.Equal(t, Point{1, 1}, p, "should add correctly")

}
