package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRectCenter(t *testing.T) {
	r := Rect{Point{1, 1}, Point{7, 7}}
	assert.Equal(t, Point{4, 4}, r.Center(), "should add correctly")
}
