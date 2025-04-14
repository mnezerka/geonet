package store

import (
	"encoding/json"
	"mnezerka/geonet/log"
	"mnezerka/geonet/tracks"

	geojson "github.com/paulmach/go.geojson"
)

type Track struct {
	Id   int64            `json:"id" bson:"id"`
	Meta tracks.TrackMeta `json:"meta" bson:"meta"`
}

type Meta struct {
	Tracks []*Track `json:"tracks"`
}

type Store interface {
	ToGeoJson() *geojson.FeatureCollection
	GetMeta() Meta
}

func Export(store Store) []byte {

	toExport := struct {
		GeoJson *geojson.FeatureCollection `json:"geojson"`
		Meta    Meta                       `json:"meta"`
	}{
		GeoJson: store.ToGeoJson(),
		Meta:    store.GetMeta(),
	}

	bytesJson, err := json.MarshalIndent(toExport, "", " ")
	if err != nil {
		log.ExitWithError(err)
	}

	return bytesJson
}
