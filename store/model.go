package store

import "mnezerka/geonet/tracks"

type Track struct {
	Id   int64            `json:"id" bson:"id"`
	Meta tracks.TrackMeta `json:"meta" bson:"meta"`
}

type Meta struct {
	Tracks []*Track `json:"tracks"`
}

type Stat struct {
	TracksProcessed    int64
	PointsProcessed    int64
	PointsCreated      int64
	PointsSimplified   int64
	PointsReused       int64
	PointsFinal        int64
	PointsRendered     int64
	EdgesCreated       int64
	EdgesReused        int64
	EdgesSimplified    int64
	EdgesFinal         int64
	EdgesRendered      int64
	SegmentsProcessed  int64
	SegmentsSimplified int64
	SegmentsRendered   int64
}
