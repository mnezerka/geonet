package cmd

import (
	"fmt"

	"github.com/ptrv/go-gpx"
)

func GpxToPolyline(gpx *gpx.Gpx) string {

	result := ""

	// tracks
	for i := 0; i < len(gpx.Tracks); i++ {

		fmt.Printf("  track: %s\n", gpx.Tracks[i].Name)
		result += "[\n"

		// track segments
		for j := 0; j < len(gpx.Tracks[i].Segments); j++ {
			fmt.Printf("    segment with %d points\n", len(gpx.Tracks[i].Segments[j].Waypoints))

			for k := 0; k < len(gpx.Tracks[i].Segments[j].Waypoints); k++ {
				result += fmt.Sprintf("[%f, %f]", gpx.Tracks[i].Segments[j].Waypoints[k].Lat, gpx.Tracks[i].Segments[j].Waypoints[k].Lon)
				result += ",\n"
			}
		}
		result += "]\n"
	}

	if len(result) > 0 {
		result = fmt.Sprintf("var polyline = L.polyline(%s, {color: 'blue'}).addTo(map);\n", result)
		result += "if (bounds === null) { bounds = polyline.getBounds() } else { bounds.extend(polyline.getBounds()); };\n"
	}
	return result
}
