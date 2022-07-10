package store

import "math"

/*
Function to return the minimum distance between a line segment AB and a point E

https://www.geeksforgeeks.org/minimum-distance-from-a-point-to-the-line-segment-using-vectors/

Given the coordinates of two endpoints A(x1, y1), B(x2, y2) of the line segment
and coordinates of a point E(x, y); the task is to find the minimum distance
from the point to line segment formed with the given coordinates.

Approach: The idea is to use the concept of vectors to solve the problem since
the nearest point always lies on the line segment. Assuming that the direction
of vector AB is A to B, there are three cases that arise:

1. The nearest point from the point E on the line segment AB is point B itself
if the dot product of vector AB(A to B) and vector BE(B to E) is positive where
E is the given point. Since AB . BE > 0, the given point lies in the same
direction as the vector AB is and the nearest point must be B itself because the
nearest point lies on the line segment.

                           E
                          /
                         /
                        /
   A ----------------- B

2. The nearest point from the point E on the line segment AB is point A itself
if the dot product of vector AB(A to B) and vector AE(A to E) is negative where
E is the given point. Since AB . AE < 0, the given point lies in the opposite
direction of the line segment AB and the nearest point must be A itself because
the nearest point lies on the line segment.

    E
     \
      \
       \
        A ----------------- B


3. Otherwise, if the dot product is 0, then the point E is perpendicular to the
line segment AB and the perpendicular distance to the given point E from the
line segment AB is the shortest distance. If some arbitrary point F is the point
on the line segment which is perpendicular to E, then the perpendicular distance
can be calculated as |EF| = |(AB X AE)/|AB||

                   E
                   |
                   |
                   |
   A ----------------- B
*/
func Distance(a, b, e Point) float64 {

	// vector AB
	ab := Vector{b.X - a.X, b.Y - a.Y}

	// vector BP
	be := Vector{e.X - b.X, e.Y - b.Y}

	// vector AP
	ae := Vector{e.X - a.X, e.Y - a.Y}

	// Variables to store dot product
	//double AB_BE, AB_AE;

	// Calculating the dot product
	ab_dot_be := ab.Dot(be)
	ab_dot_ae := ab.Dot(ae)

	// Minimum distance from
	// point E to the line segment
	var reqAns float64 = 0

	// Case 1
	if ab_dot_be > 0 {

		// Finding the magnitude
		var y float64 = e.Y - b.Y
		var x float64 = e.X - b.X
		reqAns = math.Sqrt(x*x + y*y)
	} else if ab_dot_ae < 0 {
		// Case 2
		var y float64 = e.Y - a.Y
		var x float64 = e.X - a.X
		reqAns = math.Sqrt(x*x + y*y)
	} else {
		// Case 3
		// Finding the perpendicular distance
		var x1 float64 = ab.X
		var y1 float64 = ab.Y
		var x2 float64 = ae.X
		var y2 float64 = ae.Y
		var mod float64 = math.Sqrt(x1*x1 + y1*y1)
		reqAns = math.Abs(x1*y2-y1*x2) / mod
	}
	return reqAns
}
