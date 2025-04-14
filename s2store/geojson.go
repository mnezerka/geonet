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

	s.stat.PointsFinal = 0
	s.stat.EdgesFinal = 0

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

		s.stat.PointsRendered = int64(len(points))
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

				s.stat.SegmentsRendered++
			}
		} else {
			for _, edge := range s.edges {

				p1, p1ok := points[edge.Id.P1]
				p2, p2ok := points[edge.Id.P2]

				// of some of the points were not found -> inconsistent data
				if !p1ok || !p2ok {
					log.Exitf("inconsistent data, some edge points where not found %d %d", edge.Id.P1, edge.Id.P2)
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
				s.stat.EdgesRendered++
			}
		}

		s.stat.EdgesFinal = int64(len(s.edges))
	}

	return collection
}
