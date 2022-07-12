package store

import "fmt"

type Line struct {
	A, B StoreId
	Id   StoreId
	// TODO: direction (A->B, B->A, both)
}

func (l Line) Str() string {
	return fmt.Sprintf("%d (%d <--> %d)", l.Id, l.A, l.B)
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
