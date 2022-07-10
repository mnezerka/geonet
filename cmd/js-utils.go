package cmd

import (
	"fmt"

	"mnezerka/geonet/store"

	"github.com/ptrv/go-gpx"
)

func GpxToPolyline(gpx *gpx.Gpx, index int) string {

	result := ""

	var order int64 = 1

	// tracks
	var points string
	for i := 0; i < len(gpx.Tracks); i++ {

		//fmt.Printf("  track: %s\n", gpx.Tracks[i].Name)
		points = "[\n"

		// track segments
		for j := 0; j < len(gpx.Tracks[i].Segments); j++ {
			//fmt.Printf("    segment with %d points\n", len(gpx.Tracks[i].Segments[j].Waypoints))

			for k := 0; k < len(gpx.Tracks[i].Segments[j].Waypoints); k++ {
				points += fmt.Sprintf("[%f, %f]", gpx.Tracks[i].Segments[j].Waypoints[k].Lat, gpx.Tracks[i].Segments[j].Waypoints[k].Lon)
				points += ",\n"

				result += fmt.Sprintf("var circle = L.circle([%f, %f], {radius:1}).addTo(map).bindTooltip('%d / %d');\n", gpx.Tracks[i].Segments[j].Waypoints[k].Lat, gpx.Tracks[i].Segments[j].Waypoints[k].Lon, index, order)
				order++
			}
		}
		points += "]\n"
	}

	if len(points) > 0 {
		result += fmt.Sprintf("var polyline = L.polyline(%s, {color: 'blue', weight: '2', opacity: '0.3'}).addTo(map);\n", points)
		result += "if (bounds === null) { bounds = polyline.getBounds() } else { bounds.extend(polyline.getBounds()); };\n"
	}
	return result
}

func StoreHullsToPolygons(s *store.Store) string {
	result := ""

	for id, h := range s.Hulls {
		r := h.BoundRect()
		result += fmt.Sprintf("var rect%d = L.rectangle([[%f, %f], [%f, %f]], {color: 'red', weight: '1'}).addTo(map);\n", id, r.Corner1.X, r.Corner1.Y, r.Corner2.X, r.Corner2.Y)
	}

	return result
}

func StoreLinesToPolylines(s *store.Store) string {
	result := ""

	for i := 0; i < len(s.Lines); i++ {
		ha := s.Hulls[s.Lines[i].A]
		hb := s.Hulls[s.Lines[i].B]
		//log.Infof("ha %v hb %v", )
		ca := ha.BoundRect().Center()
		cb := hb.BoundRect().Center()
		result += fmt.Sprintf("var polyline%d = L.polyline([[%f, %f], [%f, %f]], {color: 'green', dashArray: '4', weight: '1'}).addTo(map);\n", i, ca.X, ca.Y, cb.X, cb.Y)
	}

	return result
}
