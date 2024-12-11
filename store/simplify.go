package store

import (
	"context"

	"mnezerka/geonet/log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var freeNotProcessedPointsFilter bson.M = bson.M{
	"begin":     false, // point cannot be first in some track
	"end":       false, // point cannot be last in some track
	"crossing":  false, // point cannot be crossing
	"processed": false, // point cannot be processed
}

// closing point of the path segment -> could be begin, end or crossing
var closingPointsFilter bson.M = bson.M{
	"$or": bson.A{
		bson.M{"begin": true},
		bson.M{"end": true},
		bson.M{"crossing": true},
	},
}

func (ms *MongoStore) Simplify() error {

	_, err := ms.points.UpdateMany(context.TODO(), bson.M{}, bson.M{"$set": bson.M{"processed": false}})
	if err != nil {
		panic(err)
	}

	// 1. get random free point (free = not crossing, not start nor end of track)
	// pick the first free point and start processing
	freePoint, err := ms.getSinglePoint(bson.M{
		"$and": bson.A{
			bson.M{"begin": bson.M{"$ne": true}},    // point cannot be first in some track
			bson.M{"end": bson.M{"$ne": true}},      // point cannot be last in some track
			bson.M{"crossing": bson.M{"$ne": true}}, // point cannot be crossing
		},
	})

	if err != nil {
		return err
	}

	if freePoint == nil {
		log.Infof("no free point found")
		return nil
	}

	log.Debugf("free point: %d, setting processed -> true", freePoint.Id)
	ms.updateSinglePoint((*freePoint).Id, bson.M{"processed": true})

	// 2. find free segment (not processed points) from the free point - both sides
	nIds := ms.findNeighbours(*freePoint)

	// filter out points that are beginnings, endings and crossings
	log.Debugf("nIds=%v", nIds)
	n := ms.getPointsByIds(nIds, bson.M{})

	path := []DbPoint{*freePoint}

	for leg := 0; leg < 2; leg++ {
		log.Debugf("starting iteration %d with path %v", leg+1, pointsToIds(path))

		if len(nIds) > 0 {
			nextPointExists := true
			for nextPointExists {
				nextPoint := ms.getNextPoint(n[0], path)
				if nextPoint != nil {
					path = append(path, *nextPoint)
					log.Debugf("mark point: %v, setting processed -> true", nextPoint.Id)
					log.Debugf("path was extended to: %v", pointsToIds(path))
					ms.updateSinglePoint(nextPoint.Id, bson.M{"processed": true})
				} else {
					nextPointExists = false
				}
			}
		}

		for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
			path[i], path[j] = path[j], path[i]
		}

		log.Debugf("path at the end of iteration %d: %v", leg+1, pointsToIds(path))
	}

	// 3. simplify

	// 4. adapt edges

	return nil
}

func (ms *MongoStore) getNextPoint(p DbPoint, path []DbPoint) *DbPoint {

	// nothing to do if path is empty
	if len(path) == 0 {
		return nil
	}

	// get last point from path
	p = path[len(path)-1]

	log.Debugf("looking for path continuation from last path point: %d", p.Id)
	//path = append(path, p)

	// stop on track ends, begins and crossings
	if p.Begin || p.End || p.Crossing {
		log.Debugf("last path point %d is begin | end | crossing -> path is complete", p.Id)
		return nil
	}

	// else we have another free point that is being processed
	nIds := ms.findNeighbours(p)
	log.Debugf("direct neighbours: %v", nIds)

	n := ms.getPointsByIds(nIds, freeNotProcessedPointsFilter)
	log.Debugf("free not processed neighbours: %v", pointsToIds(n))

	if len(n) > 0 {
		return &n[0]
	}

	// there are no free unprocessed points, try to find begin, end or crossing as closing point of the path
	c := ms.getPointsByIds(nIds, closingPointsFilter)
	log.Debugf("possible closing points : %v", pointsToIds(c))

	if len(c) > 0 {
		return &c[0]
	}

	return nil
}

func (ms *MongoStore) findNeighbours(p DbPoint) []int64 {

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
		panic(err)
	}

	if err = cursor.All(context.TODO(), &edges); err != nil {
		panic(err)
	}

	// no edges found
	if len(edges) == 0 {
		return result
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

	return result
}

func (ms *MongoStore) getSinglePoint(filter bson.M) (*DbPoint, error) {
	var point DbPoint

	// FindOne returns a single document
	err := ms.points.FindOne(context.TODO(), filter).Decode(&point)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// No document was found
			return nil, nil
		}
		return nil, err
	}

	return &point, nil
}

func (ms *MongoStore) getPointsByIds(ids []int64, filter bson.M) []DbPoint {
	var points []DbPoint

	filter["id"] = bson.M{
		"$in": ids,
	}

	log.Debugf("filter=%v", filter)

	cursor, err := ms.points.Find(context.TODO(), filter)
	if err != nil {
		panic(err)
	}

	if err = cursor.All(context.TODO(), &points); err != nil {
		panic(err)
	}

	/*
		if len(ids) != len(points) {
			panic(fmt.Errorf("inconsistent number of ids (%d) and fetched entries (%d)", len(ids), len(points)))
		}
	*/

	return points
}

func (ms *MongoStore) updateSinglePoint(id int64, update bson.M) {

	_, err := ms.points.UpdateOne(context.TODO(), bson.M{"id": id}, bson.M{"$set": update})
	if err != nil {
		panic(err)
	}
}

func pointsToIds(points []DbPoint) []int64 {
	var ids []int64
	for i := 0; i < len(points); i++ {
		ids = append(ids, points[i].Id)
	}

	return ids
}
