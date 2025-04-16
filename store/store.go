package store

import (
	"encoding/json"
	"fmt"
	"io"
	"mnezerka/geonet/config"
	"mnezerka/geonet/log"
	"mnezerka/geonet/svg"
	"mnezerka/geonet/tracks"
	"strings"

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
	ToGeoJson(func(feature *geojson.Feature)) *geojson.FeatureCollection
	GetMeta() Meta
}

func Export(store Store) []byte {

	toExport := struct {
		GeoJson *geojson.FeatureCollection `json:"geojson"`
		Meta    Meta                       `json:"meta"`
	}{
		GeoJson: store.ToGeoJson(nil),
		Meta:    store.GetMeta(),
	}

	bytesJson, err := json.MarshalIndent(toExport, "", " ")
	if err != nil {
		log.ExitWithError(err)
	}

	return bytesJson
}

func ExportSvg(store Store) string {

	gs := store.ToGeoJson(func(feature *geojson.Feature) {

		if feature.Geometry.Type == "LineString" {
			feature.SetProperty("style", "stroke: black; stroke-width: 2; fill: none")
		}

		if feature.Geometry.Type == "Point" {
			fill := "black"
			radius := "5px"
			style := "stroke: none;"
			if val, exists := feature.Properties["begin"]; exists && val == true {
				fill = "#4dd091"
				radius = "8px"
			}
			if val, exists := feature.Properties["end"]; exists && val == true {
				fill = "#ff5c77"
				radius = "8px"
			}
			if val, exists := feature.Properties["crossing"]; exists && val == true {
				style += "fill: #1BA1E2; r: 8px;"
				fill = "#1BA1E2"
				radius = "8px"
			}

			feature.SetProperty("style", fmt.Sprintf("stroke: none; fill: %s; r: %s", fill, radius))
		}

	})

	s := svg.NewSVG()
	s.AddFeatureCollection(gs)

	got := s.Draw(
		float64(config.Cfg.SvgWidth),
		float64(config.Cfg.SvgHeight),
		svg.WithAttribute("xmlns", "http://www.w3.org/2000/svg"),
		svg.UseProperties([]string{"style"}),
		svg.WithPadding(svg.Padding{
			Top:    10,
			Right:  10,
			Bottom: 10,
			Left:   10,
		}),
		svg.WithCustomDecorator(func(w io.Writer, sf svg.ScaleFunc, feature *geojson.Feature) {
			if feature.Geometry.Type == "LineString" {

				tracks, exists := feature.Properties["tracks"]
				if !exists {
					return
				}

				tracksStr := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(tracks)), ","), "[]")

				// add text
				ps := feature.Geometry.LineString

				// text in the  middle of the first line segment
				if len(ps) > 1 {
					x1, y1 := sf(ps[0][0], ps[0][1])
					x2, y2 := sf(ps[1][0], ps[1][1])

					midX := (x1 + x2) / 2
					midY := (y1 + y2) / 2
					fmt.Fprintf(w, `<text x="%f" y="%f">%s</text>`, midX, midY, tracksStr)
				}

			}
		}),
	)
	return got
}

func ExportGeoJson(store Store) []byte {

	gs := store.ToGeoJson(nil)

	bytesJson, err := json.MarshalIndent(gs, "", " ")
	if err != nil {
		log.ExitWithError(err)
	}

	return bytesJson
}
