package main

// Vector - struct holding X Y values of a 2D vector
type Vector struct {
	X, Y float64
}

func (a Vector) Add(b Vector) Vector {
	a.X += b.X
	a.Y += b.Y
	return a
}

func (a Vector) Dot(b Vector) float64 {
	return a.X*b.X + a.Y*b.Y
}
