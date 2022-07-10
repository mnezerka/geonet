package store

type Rect struct {
	Corner1 Point
	Corner2 Point
}

func (r Rect) Center() Point {
	return Point{
		r.Corner1.X + (r.Corner2.X-r.Corner1.X)/2,
		r.Corner1.Y + (r.Corner2.Y-r.Corner1.Y)/2,
	}
}
