# GeoNet

GeoNet is a library for generating net from multiple gps tracks.

[![Go](https://github.com/mnezerka/geonet/actions/workflows/go.yml/badge.svg?branch=master)](https://github.com/mnezerka/geonet/actions/workflows/go.yml)

Current state of the algorithm could be checkd on [Cases page](https://mnezerka.github.io/geonet/)

## Algorithm

- tracks are added one by one
- for each track
  - if the track is first, store all points and continue with next track
  - go through track points one by one
    - create new hull from point (PH)
    - look for existing hull (H) which is close to the point
    - if such hull exist, try to add it (merge hulls H and PH). If size of
      joined hull is under predefined treshold, add point to H and update
      metadata of H (point, track, etc.)
    - if such hull doesn't exist, register new hull

TODO: document line splitting part of algorithm

## Data Structures

### FDH

FDH = Fixed Directional Hull, alternative, and I think more frequent term is
k-DOP. (Not sure for what that stands for anymore)

A suggestion: If you collect the fixed directions into a matrix, say, A; and if
you make your points into column-vectors, say x^i; then the box coordinates
come from the standard matrix multiplication of vectors A : x —> z by

z^i = Ax^i

In our case, z lives in R^8 (it is an 8-D column-vector), A is the 8x2 matrix
of fixed directions, and x comes from R^2 (it is a 2-D column-vector).


