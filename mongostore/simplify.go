package store

import (
	"context"
	"fmt"
	"strconv"

	"mnezerka/geonet/log"

	"github.com/mnezerka/gpxcli/gpxutils"
	"github.com/tkrajina/gpxgo/gpx"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"mnezerka/geonet/config"
)

var freeNotProcessedPointsFilter bson.M = bson.M{
	"begin":     false, // point cannot be first in some track
	"end":       false, // point cannot be last in some track
	"crossing":  false, // point cannot be crossing
	"processed": false, // point cannot be processed
}

func (ms *MongoStore) Simplify() {

	ms.Log(fmt.Sprintf("simplify cfg=(%s)",
		config.Cfg.ToString(),
	))

	// reset all points to not processed state to be sure we start with clean setup
	_, err := ms.points.UpdateMany(context.TODO(), bson.M{}, bson.M{"$set": bson.M{"processed": false}})
	if err != nil {
		panic(err)
	}

	// simplify segments until there is no more to simplify
	for ms.SimplifySegment() {
	}
}

func (ms *MongoStore) SimplifySegment() bool {

	// 1. get random free point (free = not crossing, not start nor end of track)
	// pick the first free point and start processing
	freePoint, err := ms.getSinglePoint(bson.M{
		"$and": bson.A{
			bson.M{"begin": bson.M{"$ne": true}},     // point cannot be first in some track
			bson.M{"end": bson.M{"$ne": true}},       // point cannot be last in some track
			bson.M{"crossing": bson.M{"$ne": true}},  // point cannot be crossing
			bson.M{"processed": bson.M{"$ne": true}}, // point cannot be crossing
		},
	})

	if err != nil {
		panic(err)
	}

	if freePoint == nil {
		log.Infof("no free point found")
		return false
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
	simplifiedPath := simplifyPath(path, ms.cfg.SimplifyMinDistance)
	log.Debugf("simplified path: %v", simplifiedPath)

	// 4. adapt edges
	ms.adaptEdges(path, simplifiedPath)

	return true
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
	closingPointsFilter := bson.M{
		"$and": bson.A{
			bson.M{
				"$or": bson.A{
					bson.M{"begin": true},
					bson.M{"end": true},
					bson.M{"crossing": true},
				},
			},
			bson.M{"id": bson.M{"$nin": pointsToIds(path)}},
		},
	}

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

func (ms *MongoStore) getSingleEdge(filter bson.M) *DbEdge {
	var edge DbEdge

	// FindOne returns a single document
	err := ms.edges.FindOne(context.TODO(), filter).Decode(&edge)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// No document was found
			return nil
		}
		panic(err)
	}

	return &edge
}

func pointsToIds(points []DbPoint) []int64 {
	var ids []int64
	for i := 0; i < len(points); i++ {
		ids = append(ids, points[i].Id)
	}

	return ids
}

func simplifyPath(path []DbPoint, minDistance int64) []int64 {

	gpxFile := gpx.GPX{}

	for i := 0; i < len(path); i++ {
		np := gpx.GPXPoint{
			Point: gpx.Point{
				Latitude:  path[i].Loc.Coordinates[0],
				Longitude: path[i].Loc.Coordinates[1],
			},
			Name: strconv.FormatInt(path[i].Id, 10),
		}

		gpxFile.AppendPoint(&np)
	}

	gpxFile.SimplifyTracks(float64(minDistance))

	simplifiedPath, err := gpxutils.GpxFileToPoints(&gpxFile)
	if err != nil {
		panic(err)
	}

	var result []int64
	for i := 0; i < len(simplifiedPath); i++ {
		id, err := strconv.ParseInt(simplifiedPath[i].Name, 10, 64)
		if err != nil {
			panic(err)
		}
		result = append(result, id)
	}

	return result
}

func (ms *MongoStore) adaptEdges(path []DbPoint, simplifiedPathIds []int64) {
	var beginId int64 = NIL_ID
	var lastIx = 0

	for i := 0; i < len(simplifiedPathIds); i++ {
		log.Debugf("iter %d", i)
		if i == 0 {
			beginId = simplifiedPathIds[i]
			continue
		}

		endId := simplifiedPathIds[i]

		log.Debugf("updating edge %d -> %d", beginId, endId)

		if path[lastIx].Id != beginId {
			panic("path point doesn't match starting point of the edge")
		}
		var toRemove []int64

		// loop
		for j := lastIx + 1; j < len(path); j++ {
			//log.Debugf("checking point ix=%d id=%d", j, path[j].Id)

			if path[j].Id == endId {
				log.Debugf("match found, to be removed: %v", toRemove)
				ms.mergeEdges(beginId, endId, toRemove)
				toRemove = nil // make slice empty
				beginId = endId
				lastIx = j
				break
			} else {
				toRemove = append(toRemove, path[j].Id)
			}
		}
	}
}

func (ms *MongoStore) mergeEdges(beginId, endId int64, toRemoveIds []int64) {

	if len(toRemoveIds) == 0 {
		log.Debugf("noting to remove for begin: %d and end: %d", beginId, endId)
		return
	}

	// prepare final edge (create new or reuse existing)
	finalEdgePoints := []int64{min(beginId, endId), max(beginId, endId)}
	finalEdgeId := edgeIdFromPointIds(beginId, endId)

	log.Debugf("final edge id: %v", finalEdgeId)
	finalEdge := ms.getSingleEdge(bson.M{"id": finalEdgeId})
	if finalEdge == nil {
		log.Debugf("creating new final edge, id: %v", finalEdgeId)

		finalEdge := DbEdge{
			Id:     finalEdgeId,
			Points: finalEdgePoints,
			Tracks: []int64{},
			Count:  0,
		}
		_, err := ms.edges.InsertOne(context.TODO(), finalEdge)
		if err != nil {
			panic(err)
		}
	} else {
		log.Debugf("reusing existing edge, id: %v", finalEdgeId)
	}

	// get first edge to be removed
	firstEdgeId := edgeIdFromPointIds(beginId, toRemoveIds[0])
	log.Debugf("first edge (to be removed) id:: %v", firstEdgeId)
	firstEdge := ms.getSingleEdge(bson.M{"id": firstEdgeId})
	if firstEdge == nil {
		panic(fmt.Errorf("edge %s not found", firstEdgeId))
	}

	// update final edge with properties of the first edge
	finalEdgeUpdate := bson.M{
		"count":  firstEdge.Count,
		"tracks": firstEdge.Tracks,
	}

	_, err := ms.edges.UpdateOne(context.TODO(), bson.M{"id": finalEdgeId}, bson.M{"$set": finalEdgeUpdate})
	if err != nil {
		panic(err)
	}

	// delete all edges
	var allIds = append(append([]int64{beginId}, toRemoveIds...), endId)
	var edgeIdsToRemove []string
	log.Debugf("all ids: %v", allIds)
	for i := 1; i < len(allIds); i++ {
		edgeId := edgeIdFromPointIds(allIds[i-1], allIds[i])
		edgeIdsToRemove = append(edgeIdsToRemove, edgeId)
	}

	// delete redundant points
	log.Debugf("points to be deleted %v", toRemoveIds)
	ms.points.DeleteMany(context.TODO(), bson.M{"id": bson.M{"$in": toRemoveIds}})

	// delete redundant edges
	log.Debugf("edges to be deleted %v", edgeIdsToRemove)
	ms.edges.DeleteMany(context.TODO(), bson.M{"id": bson.M{"$in": edgeIdsToRemove}})
}
