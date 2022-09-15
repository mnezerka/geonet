package cmd

import (
	"fmt"

	"mnezerka/geonet/store"

	"github.com/ptrv/go-gpx"
)

func GpxToPolyline(gpx *gpx.Gpx, index int, mapId int) string {

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

				//result += fmt.Sprintf("var circle = L.circle([%f, %f], {radius:1}).addTo(map).bindTooltip('%d / %d');\n", gpx.Tracks[i].Segments[j].Waypoints[k].Lat, gpx.Tracks[i].Segments[j].Waypoints[k].Lon, index, order)
				result += fmt.Sprintf("var circle = L.circle([%f, %f], {radius:1}).addTo(map%d);\n", gpx.Tracks[i].Segments[j].Waypoints[k].Lat, gpx.Tracks[i].Segments[j].Waypoints[k].Lon, mapId)
				order++
			}
		}
		points += "]\n"
	}

	if len(points) > 0 {
		result += fmt.Sprintf("var polyline = L.polyline(%s, {color: 'blue', weight: '2',dashArray: '2 4', opacity: '1'}).addTo(map%d);\n", points, mapId)
		result += fmt.Sprintf("if (bounds%d === null) { bounds%d = polyline.getBounds() } else { bounds%d.extend(polyline.getBounds()); };\n", mapId, mapId, mapId)
	}
	return result
}

func StoreHullsToPolygons(s *store.Store, mapId int) string {
	result := ""

	for id, h := range s.Hulls {
		r := h.BoundRect()
		rc := r.Center()
		if h.Size() == 0 {
			result += fmt.Sprintf("var hullcircle%d = L.circle([%f, %f], {radius: 1.5, color: 'red', weight: '3'}).addTo(map%d).bindTooltip('hull %d');\n", id, rc.X, rc.Y, mapId, id)
		} else {
			result += fmt.Sprintf("var rect%d = L.rectangle([[%f, %f], [%f, %f]], {color: 'red', weight: '5'}).addTo(map%d).bindTooltip('hull %d');\n", id, r.Corner1.X, r.Corner1.Y, r.Corner2.X, r.Corner2.Y, mapId, id)
		}
	}

	return result
}

func StoreLinesToPolylines(s *store.Store, mapId int) string {
	result := ""

	for _, line := range s.Lines {
		ha := s.Hulls[line.A]
		hb := s.Hulls[line.B]
		//log.Infof("ha %v hb %v", )
		ca := ha.BoundRect().Center()
		cb := hb.BoundRect().Center()
		result += fmt.Sprintf("var polyline%d = L.polyline([[%f, %f], [%f, %f]], {color: 'green', weight: '2'}).addTo(map%d);\n", line.Id, ca.X, ca.Y, cb.X, cb.Y, mapId)
	}

	return result
}
