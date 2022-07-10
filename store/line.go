package store

type Line struct {
	A, B HullId

	// TODO: direction (A->B, B->A, both)
}

func (l Line) Equals(l2 *Line) bool {

	if l.A == l2.A && l.B == l2.B {
		return true
	}

	if l.A == l2.B && l.B == l2.A {
		return true
	}

	return false
}
