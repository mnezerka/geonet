package s2store

import (
	"fmt"
	"mnezerka/geonet/log"
	"strings"

	geojson "github.com/paulmach/go.geojson"
)

func (s *S2Store) ToGeoJson() *geojson.FeatureCollection {

	collection := geojson.NewFeatureCollection()

	points := s.index.GetLocations()

	s.Stat.PointsFinal = 0
	s.Stat.EdgesFinal = 0

	if s.cfg.ShowPoints {

		for _, point := range points {

			pnt := geojson.NewPointFeature([]float64{point.Lng, point.Lat})

			pnt.SetProperty("id", point.Id)
			pnt.SetProperty("tracks", point.Tracks)
			pnt.SetProperty("begin", point.Begin)
			pnt.SetProperty("end", point.End)
			pnt.SetProperty("crossing", point.Crossing)
			// TODO: pnt.SetProperty("count", point.Count)

			collection.AddFeature(pnt)
		}

		s.Stat.PointsRendered = int64(len(points))
	}

	if s.cfg.ShowEdges {

		if s.cfg.GeoJsonMergeEdges {

			// reset all points to not processed state to be sure we start with clean setup
			//s.index.SetLocationsNotProcessed()
			s.setEdgesNotProcessed()

			for {
				path := s.getNextFreeSegment()
				if len(path) < 2 {
					break
				}

				log.Debugf("path for exporting: %v", pointsToIds(path))

				pathCoordinates := [][]float64{}

				for i := 0; i < len(path); i++ {
					p1 := path[i]
					pathCoordinates = append(pathCoordinates, []float64{p1.Lng, p1.Lat})
				}

				// line is complete, add metadata
				line := geojson.NewLineStringFeature(pathCoordinates)
				// ids of all points => 1-5-3-6-7

				if len(path) < 10 {
					line.SetProperty("id", strings.Trim(strings.Join(strings.Fields(fmt.Sprint(pointsToIds(path))), "-"), "[]"))
				} else {
					line.SetProperty("id", fmt.Sprintf("%d-%d", path[0].Id, path[len(path)-1].Id))
				}

				line.SetProperty("tracks", path[0].Tracks)

				// TODO: line.SetProperty("count", edge.Count)

				collection.AddFeature(line)

				s.Stat.SegmentsRendered++
			}
		} else {
			for _, edge := range s.edges {

				p1, p1ok := points[edge.Points[0]]
				p2, p2ok := points[edge.Points[1]]

				// of some of the points were not found -> inconsistent data
				if !p1ok || !p2ok {
					log.Exitf("inconsistent data, some edge points where not found %d %d", edge.Points[0], edge.Points[1])
				}

				edgeCoordinates := [][]float64{
					{p1.Lng, p1.Lat},
					{p2.Lng, p2.Lat},
				}

				line := geojson.NewLineStringFeature(edgeCoordinates)
				line.SetProperty("id", edgeIdToString(edge.Id))
				line.SetProperty("tracks", edge.Tracks)
				// TODO: line.SetProperty("count", edge.Count)

				collection.AddFeature(line)
				s.Stat.EdgesRendered++
			}
		}

		s.Stat.EdgesFinal = int64(len(s.edges))
	}

	return collection
}
