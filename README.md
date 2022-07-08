# GeoNet



## FDH

FDH = Fixed Directional Hull, alternative, and I think more frequent term is
k-DOP. (Not sure for what that stands for anymore)

A suggestion: If you collect the fixed directions into a matrix, say, A; and if
you make your points into column-vectors, say x^i; then the box coordinates
come from the standard matrix multiplication of vectors A : x —> z by

z^i = Ax^i

In our case, z lives in R^8 (it is an 8-D column-vector), A is the 8x2 matrix
of fixed directions, and x comes from R^2 (it is a 2-D column-vector).


