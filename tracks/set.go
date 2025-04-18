package tracks

import (
	"encoding/json"
	"fmt"
	"io"
	"mnezerka/geonet/config"
	"mnezerka/geonet/log"
	"mnezerka/geonet/svg"
	"mnezerka/geonet/utils"

	geojson "github.com/paulmach/go.geojson"
)

// default attributes ()
type Set struct {
	tracks []*Track
}

func NewSet(filePaths []string) *Set {
	s := &Set{}

	if len(filePaths) > 0 {
		s.LoadFromFiles(filePaths)
	}

	return s
}

func (s *Set) InterpolateDistance(distance int64) {
	for _, t := range s.tracks {
		t.InterpolateDistance(distance)
	}
}

func (s *Set) LoadFromFiles(filePaths []string) {

	for i := 0; i < len(filePaths); i++ {

		t := NewTrack(filePaths[i])
		log.Infof("  %d/%d %s (%d points)", i+1, len(filePaths), filePaths[i], len(t.Points))
		s.tracks = append(s.tracks, t)
	}
}

func (s *Set) ToGeoJson(each func(feature *geojson.Feature)) *geojson.FeatureCollection {

	collection := geojson.NewFeatureCollection()

	for i, t := range s.tracks {

		lineCoordinates := [][]float64{}

		for _, p := range t.Points {
			lineCoordinates = append(lineCoordinates, []float64{p.Longitude, p.Latitude})
		}
		line := geojson.NewLineStringFeature(lineCoordinates)

		line.SetProperty("id", i)

		if each != nil {
			each(line)
		}

		collection.AddFeature(line)
	}

	if config.Cfg.ShowPoints {
		for tracki, t := range s.tracks {
			for i, p := range t.Points {
				pnt := geojson.NewPointFeature([]float64{p.Longitude, p.Latitude})
				pnt.SetProperty("track", tracki)

				if i == 0 {
					pnt.SetProperty("begin", true)
				}
				if i == len(t.Points)-1 {
					pnt.SetProperty("end", true)
				}
				if each != nil {
					each(pnt)
				}

				collection.AddFeature(pnt)
			}
		}
	}

	return collection
}

func (s *Set) ExportGeoJson() []byte {
	gs := s.ToGeoJson(nil)

	bytesJson, err := json.MarshalIndent(gs, "", " ")
	if err != nil {
		log.ExitWithError(err)
	}

	return bytesJson
}

func (s *Set) ExportSvg() string {

	gs := s.ToGeoJson(func(feature *geojson.Feature) {

		if feature.Geometry.Type == "LineString" {
			color := "black"
			if val, exists := feature.Properties["id"]; exists {
				color = utils.GetDarkPastelColor(val.(int))
			}

			style := fmt.Sprintf("stroke: %s; stroke-width: 2; fill: none", color)
			feature.SetProperty("style", style)
		}
		if feature.Geometry.Type == "Point" {
			color := "black"
			radius := "5px"
			if val, exists := feature.Properties["track"]; exists {
				color = utils.GetDarkPastelColor(val.(int))
			}

			style := fmt.Sprintf("stroke: none; fill: %s; r: %s", color, radius)

			feature.SetProperty("style", style)
		}

	})

	svgOut := svg.NewSVG()
	svgOut.AddFeatureCollection(gs)

	got := svgOut.Draw(
		float64(config.Cfg.SvgWidth),
		float64(config.Cfg.SvgHeight),
		svg.WithAttribute("xmlns", "http://www.w3.org/2000/svg"),
		svg.UseProperties([]string{"style"}),
		svg.WithPadding(svg.Padding{
			Top:    float64(config.Cfg.SvgPadding),
			Right:  float64(config.Cfg.SvgPadding),
			Bottom: float64(config.Cfg.SvgPadding),
			Left:   float64(config.Cfg.SvgPadding),
		}),
		svg.WithCustomDecorator(func(w io.Writer, sf svg.ScaleFunc, feature *geojson.Feature) {
			if feature.Geometry.Type == "LineString" {

				if config.Cfg.SvgEdgeLabels {

					id, exists := feature.Properties["id"]
					if !exists {
						return
					}

					// add text
					ps := feature.Geometry.LineString

					// text in the  middle of the first line segment
					if len(ps) > 1 {
						x1, y1 := sf(ps[0][0], ps[0][1])
						x2, y2 := sf(ps[1][0], ps[1][1])

						midX := (x1 + x2) / 2
						midY := (y1 + y2) / 2
						fmt.Fprintf(w, `<text x="%f" y="%f">Track %d</text>`, midX, midY, id.(int)+1)
					}
				}
			}
		}),
	)
	return got
}
