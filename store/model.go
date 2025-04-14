package store

import "mnezerka/geonet/tracks"

type Track struct {
	Id   int64            `json:"id" bson:"id"`
	Meta tracks.TrackMeta `json:"meta" bson:"meta"`
}

type Meta struct {
	Tracks []*Track `json:"tracks"`
}
