package store

// Vector - struct holding X Y values of a 2D vector
type Vector struct {
	X, Y float64
}

func VectorFromPoint(a Point) Vector {
	return Vector{a.X, a.Y}
}

func VectorFromPoints(a, b Point) Vector {
	return Vector{b.X - a.X, b.Y - a.Y}
}

func (a Vector) Add(b Vector) Vector {
	a.X += b.X
	a.Y += b.Y
	return a
}

func (a Vector) Sub(b Vector) Vector {
	a.X -= b.X
	a.Y -= b.Y
	return a
}

func (a Vector) Dot(b Vector) float64 {
	return a.X*b.X + a.Y*b.Y
}

func (a Vector) Scalar(v float64) Vector {
	a.X *= v
	a.Y *= v
	return a
}
