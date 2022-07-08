
func TestAdd(t *testing.T) {
	result := Vector{1., 1., 1.}.Add(Vector{2., 2., 2.})
	assert.Equal(t, Vector{3., 3., 3.}, result, "should add correctly")
}
