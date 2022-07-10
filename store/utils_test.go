package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDistance(t *testing.T) {

	d := Distance(Point{0, 0}, Point{1, 0}, Point{0, 1})
	assert.Equal(t, 1., d, "should add correctly")

	// edge case - right point
	d = Distance(Point{0, 0}, Point{1, 0}, Point{1, 0})
	assert.Equal(t, 0., d, "should add correctly")

	// edge case - left point
	d = Distance(Point{0, 0}, Point{1, 0}, Point{0, 0})
	assert.Equal(t, 0., d, "should add correctly")
}
