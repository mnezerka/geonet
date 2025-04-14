package mongostore

import (
	"context"
	"fmt"

	"mnezerka/geonet/log"

	geojson "github.com/paulmach/go.geojson"

	"go.mongodb.org/mongo-driver/bson"
)

func (ms *MongoStore) ToGeoJson() (*geojson.FeatureCollection, error) {

	collection := geojson.NewFeatureCollection()

	// render points and build dict of points for searching when
	// rendering edges

	cursor, err := ms.points.Find(context.TODO(), bson.M{})
	if err != nil {
		return nil, err
	}

	var points []DbPoint

	err = cursor.All(context.TODO(), &points)
	if err != nil {
		return nil, err
	}

	if ms.cfg.ShowPoints {

		for i := 0; i < len(points); i++ {
			point := points[i]

			pnt := geojson.NewPointFeature(point.Loc.Coordinates)

			pnt.SetProperty("id", point.Id)
			pnt.SetProperty("tracks", point.Tracks)
			pnt.SetProperty("begin", point.Begin)
			pnt.SetProperty("end", point.End)
			pnt.SetProperty("count", point.Count)
			pnt.SetProperty("crossing", point.Crossing)

			collection.AddFeature(pnt)
		}

		log.Infof("points rendered: %d", len(points))
	}

	if ms.cfg.ShowEdges {

		// render edges
		cursor, err = ms.edges.Find(context.TODO(), bson.M{})
		if err != nil {
			return nil, err
		}

		var edges []DbEdge

		err = cursor.All(context.TODO(), &edges)
		if err != nil {
			return nil, err
		}

		for i := 0; i < len(edges); i++ {
			edge := edges[i]
			p1 := findPoint(edge.Points[0], &points)
			p2 := findPoint(edge.Points[1], &points)

			// of some of the points were not found -> database is inconsistent
			if p1 == nil || p2 == nil {
				return nil, fmt.Errorf("inconsistent database, some edge points where not found %d %d", edge.Points[0], edge.Points[1])
			}

			log.Debugf("edge: %v, p1: %v p2: %v ", edge, p1, p2)

			edgeCoordinates := [][]float64{p1.Loc.Coordinates, p2.Loc.Coordinates}

			line := geojson.NewLineStringFeature(edgeCoordinates)
			line.SetProperty("id", edge.Id)
			line.SetProperty("tracks", edge.Tracks)
			line.SetProperty("count", edge.Count)

			collection.AddFeature(line)
		}

		log.Infof("edges rendered: %d", len(edges))
	}

	return collection, nil
}
