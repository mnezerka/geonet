package main

import (
	"fmt"
	"io/ioutil"
	"math"
	"os"

	"mnezerka/geonet/cmd"

	"github.com/ptrv/go-gpx"
)

/*
package go-gpx - https://pkg.go.dev/github.com/ptrv/go-gpx#Gpx
*/

// minimal distance between point in meters
const TRESHOLD = 10

const radius = 6371000 // Earth's mean radius in metets

func deg2rad(degrees float64) float64 {
	return degrees * math.Pi / 180
}

/*
Distance between two points on sphere - Haversine formula in METERS
https://community.esri.com/t5/coordinate-reference-systems-blog/distance-on-a-sphere-the-haversine-formula/ba-p/902128
*/
func distance(origin, destination gpx.Wpt) float64 {
	phi_1 := deg2rad(origin.Lat)
	phi_2 := deg2rad(destination.Lat)

	delta_phi := deg2rad(destination.Lat - origin.Lat)
	delta_lambda := deg2rad(destination.Lon - origin.Lon)

	a := math.Pow(math.Sin(delta_phi/2), 2) + math.Cos(phi_1)*math.Cos(phi_2)*math.Pow(math.Sin(delta_lambda/2), 2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	d := radius * c

	// meters
	return d
}

func add_points_to_store(store *gpx.Waypoints, waypoints *gpx.Waypoints) {

	// loop through all waypoints that should be added to the store
	for i := 0; i < len(*waypoints); i++ {
		wp := (*waypoints)[i]
		wp.Name = "added"
		//fmt.Printf("Adding point to store: %v\n", wp)
		//*store = append(*store, wp)
		// loop through store to see if point should be added or no
		// find closest point in store
		var dist float64 = -1
		for j := 0; j < len(*store); j++ {

			wpStore := (*store)[j]
			wpDist := distance(wp, wpStore)
			//fmt.Printf("dist: %f wpDist: %f\n", dist, wpDist)

			if dist == -1 {
				dist = wpDist
			} else if dist > wpDist {
				dist = wpDist
			}
		}
		if dist >= TRESHOLD {
			fmt.Printf("Adding point to store (distance is %fm)\n", dist)
			*store = append(*store, wp)
		} else {
			fmt.Printf("Ignoring point (distance is %fm)\n", dist)
		}

	}
	//*store = append(*store, (*gpx.Waypoints)...)
}

func store_to_gpx(store *gpx.Waypoints, filename string) {
	gpx_all := gpx.NewGpx()

	gpx_all.Waypoints = append(gpx_all.Waypoints, (*store)...)

	err := ioutil.WriteFile(filename, gpx_all.ToXML(), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed: %v\n", err)
		os.Exit(1)
	}

}

func maino() {

	// test case
	//p1 := gpx.Wpt{Lat: -0.116773, Lon: 51.510357}
	//p2 := gpx.Wpt{Lat: -77.009003, Lon: 38.889931}
	//p1 := gpx.Wpt{Lon: -0.116773, Lat: 51.510357}
	//p2 := gpx.Wpt{Lon: -77.009003, Lat: 38.889931}

	///fmt.Printf("Distance: %f\n", distance(p1, p2))

	//os.Exit(0)

	store := gpx.Waypoints{}

	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "No file specified\n")
		fmt.Fprintf(os.Stderr, "Use: %s file.gpx [file2.gpx] ...\n", os.Args[0])
		os.Exit(1)
	}

	for file_ix := 1; file_ix < len(os.Args); file_ix++ {

		fmt.Printf("Processing file %d: %s\n", file_ix, os.Args[file_ix])

		gpx, err := gpx.ParseFile(os.Args[file_ix])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed: %s\n", err)
			os.Exit(1)
		}

		fmt.Printf("  length 2D: %f\n", gpx.Length2D())
		fmt.Printf("  tracks: %d\n", len(gpx.Tracks))

		for i := 0; i < len(gpx.Tracks); i++ {

			track := gpx.Tracks[i]

			fmt.Printf("  track: %s\n", track.Name)

			for j := 0; j < len(track.Segments); j++ {
				segment := track.Segments[j]
				fmt.Printf("    segment with %d points\n", len(segment.Waypoints))

				// first file - store everything
				if file_ix == 1 {
					fmt.Printf("      adding %d points from first file\n", len(segment.Waypoints))
					store = append(store, segment.Waypoints...)
				} else {
					// 2nd file - loop through points
					add_points_to_store(&store, &segment.Waypoints)
				}
			}
		}

		/*
			for i := 0; i < len(gpx.Tracks); i++ {

				fmt.Printf("  track: %s\n", gpx.Tracks[i].Name)

				for j := 0; j < len(gpx.Tracks[i].Segments); j++ {
					fmt.Printf("    segment with %d points\n", len(gpx.Tracks[i].Segments[j].Waypoints))
				}
			}
		*/
	}

	store_to_gpx(&store, "all.gpx")
}

func main() {
	cmd.Execute()
}
