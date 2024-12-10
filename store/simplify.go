package store

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"mnezerka/geonet/log"
)

func (ms *MongoStore) Simplify() error {

	// 1. get all free points (free = not crossing, not start nor end of track)

	pointsAll, err := ms.getPoints(bson.M{
		"$and": bson.A{
			bson.M{"begin": bson.M{"$ne": true}}, // point cannot be first in some track
			bson.M{"end": bson.M{"$ne": true}},   // point cannot be last in some track
			//bson.M{"tracks.1": bson.M{"$exists": false}}, // second track cannot exist
		},
	})

	if err != nil {
		return err
	}

	if len(pointsAll) == 0 {
		return nil
	}

	edges, err := ms.getEdges(bson.M{})
	if err != nil {
		return err
	}

	log.Debugf("free points: %v", pointsAll)
	log.Debugf("edges: %v", *edges)

	// 2. find free segment (not processed points) from the free point - both sides

	var processed []int64

	// pick the first free point and start processing
	p := pointsAll[0]
	processed = append(processed, p.Id)

	// find edge - one side
	n, err := ms.findNeighbours(p)
	if err != nil {
		return err
	}

	log.Debugf("n=%v", n)

	// 3. simplify

	// 4. adapt edges

	return nil
}

func (ms *MongoStore) findNeighbours(p DbPoint) ([]int64, error) {

	log.Debugf("looking for neighbour of point %d", p.Id)

	var edges []DbEdge
	var result []int64

	filter := bson.M{
		"$or": bson.A{
			bson.M{"points.0": p.Id},
			bson.M{"points.1": p.Id},
		},
	}

	cursor, err := ms.edges.Find(context.TODO(), filter)
	if err != nil {
		return result, err
	}

	if err = cursor.All(context.TODO(), &edges); err != nil {
		return result, err
	}

	// no edges found
	if len(edges) == 0 {
		return result, nil
	}

	// check if edge point is free
	for i := 0; i < len(edges); i++ {
		var newId int64
		if edges[i].Points[0] == p.Id {
			newId = edges[i].Points[1]
		} else {
			newId = edges[i].Points[0]
		}

		result = append(result, newId)
	}

	log.Debugf("neighbours found: %v", result)

	return result, nil
}

func (ms *MongoStore) getPoints(filter bson.M) ([]DbPoint, error) {

	cursor, err := ms.points.Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}
	var results []DbPoint

	if err = cursor.All(context.TODO(), &results); err != nil {
		return nil, err
	}

	return results, nil
}

func (ms *MongoStore) getEdges(filter bson.M) (*[]DbEdge, error) {

	cursor, err := ms.edges.Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}
	var results []DbEdge

	if err = cursor.All(context.TODO(), &results); err != nil {
		return nil, err
	}

	return &results, nil
}
