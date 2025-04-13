package s2store

import (
	"mnezerka/geonet/log"
	"mnezerka/geonet/utils"
	"slices"
	"strconv"

	"github.com/mnezerka/gpxcli/gpxutils"
	"github.com/tkrajina/gpxgo/gpx"
	"go.mongodb.org/mongo-driver/bson"
)

var freeNotProcessedPointsFilter bson.M = bson.M{
	"begin":     false, // point cannot be first in some track
	"end":       false, // point cannot be last in some track
	"crossing":  false, // point cannot be crossing
	"processed": false, // point cannot be processed
}

func (s *S2Store) Simplify() {
	log.Debug("========== simplification ===========")

	// reset all points to not processed state to be sure we start with clean setup
	s.index.SetLocationsNotProcessed()

	// simplify segments until there is no more to simplify
	for {
		path := s.getNextFreeSegment()
		if len(path) < 2 {
			break
		}

		//  simplify
		simplifiedPath := simplifyPath(path, s.cfg.SimplifyMinDistance)
		log.Debugf("simplified path: %v", simplifiedPath)

		// adapt edges
		s.adaptEdges(path, simplifiedPath)

		// adapt statistics
		s.Stat.SegmentsProcessed++
		s.Stat.SegmentsSimplified++
	}
}

func (s *S2Store) getNextFreeSegment() []*Location {

	log.Debug("---------- get next free segment -----------")

	path := []*Location{}

	// 1. get random free point (free = not crossing, not start nor end of track)
	// pick the first free point and start processing
	freePoint := s.index.GetFirstLocation(func(loc *Location) bool {
		return !loc.Begin && !loc.End && !loc.Crossing && !loc.Processed
	})

	if freePoint == nil {
		log.Debugf("no free point found")
		return path
	}

	log.Debugf("free point: %d, setting processed -> true", freePoint.Id)
	freePoint.Processed = true

	// let's start building path from freePoint
	path = append(path, freePoint)

	// legs are paths starting at freePoint
	for leg := 0; leg < 2; leg++ {
		log.Debugf("traversing leg %d with path %v", leg+1, pointsToIds(path))

		//if len(neighbourIds) > 0 {
		nextPointExists := true
		for nextPointExists {
			nextPoint := s.getNextPoint(path)
			if nextPoint != nil {
				log.Debugf("mark point: %v, setting processed -> true", nextPoint.Id)
				nextPoint.Processed = true
				path = append(path, nextPoint)
				log.Debugf("path was extended to: %v", pointsToIds(path))
			} else {
				nextPointExists = false
			}
		}

		// reverse path to have last item the one that needs further processing (second let traversing)
		utils.Reverse(path)

		log.Debugf("path at the end of leg %d traversing: %v", leg+1, pointsToIds(path))
	}

	return path
}

func (s *S2Store) getNextPoint(path []*Location) *Location {

	// nothing to do if path is empty
	if len(path) == 0 {
		return nil
	}

	// get last point from path
	p := path[len(path)-1]

	log.Debugf("looking for path continuation from last path point: %d", p.Id)

	// stop on track ends, begins and crossings
	if p.Begin || p.End || p.Crossing {
		log.Debugf("last path point %d is begin | end | crossing -> path is complete", p.Id)
		return nil
	}

	// else we have another free point that is being processed
	neighbourIds := s.findNeighbours(*p)
	log.Debugf("all neighbours: %v", neighbourIds)
	neighbourFreeIds := s.index.GetLocationsFiltered(func(l *Location) bool {
		return !l.Begin && !l.End && !l.Crossing && !l.Processed && slices.Contains(neighbourIds, l.Id)
	})

	log.Debugf("neighbours (not processed, not begin, not end, not crossing): %v", pointsToIds(neighbourFreeIds))
	// pick first free neighbour point if exists
	if len(neighbourFreeIds) > 0 {
		return neighbourFreeIds[0]
	}

	// there are no free unprocessed points, try to find begin, end or crossing as closing point of the path
	c := s.index.GetLocationsFiltered(func(l *Location) bool {
		return (l.Begin || l.End || l.Crossing) &&
			slices.Contains(neighbourIds, l.Id) &&
			!slices.Contains(pointsToIds(path), l.Id)
	})

	log.Debugf("possible closing points : %v", pointsToIds(c))

	// pick closing point if exists
	if len(c) > 0 {
		return c[0]
	}

	return nil
}

func (s *S2Store) findNeighbours(p Location) []int64 {

	log.Debugf("looking for neighbours of point %d", p.Id)

	var result []int64

	edges := s.findPointEdges(p.Id)

	// no edges found
	if len(edges) == 0 {
		return result
	}

	for i := 0; i < len(edges); i++ {
		var newId int64
		if edges[i].P1 == p.Id {
			newId = edges[i].P2
		} else {
			newId = edges[i].P1
		}

		if !slices.Contains(result, newId) {
			result = append(result, newId)
		}

	}

	//log.Debugf("neighbours found: %v", result)

	return result
}

func (s *S2Store) findPointEdges(id int64) []S2EdgeKey {
	result := []S2EdgeKey{}

	for edgeId := range s.edges {
		if edgeId.P1 == id || edgeId.P2 == id {
			result = append(result, edgeId)
		}
	}

	return result
}

func (s *S2Store) getEdgeById(id S2EdgeKey) *S2Edge {

	if edge, ok := s.edges[id]; ok {
		return edge
	}
	return nil
}

func pointsToIds(points []*Location) []int64 {
	var ids []int64
	for i := 0; i < len(points); i++ {
		ids = append(ids, points[i].Id)
	}

	return ids
}

func simplifyPath(path []*Location, minDistance int64) []int64 {

	gpxFile := gpx.GPX{}

	for i := 0; i < len(path); i++ {
		np := gpx.GPXPoint{
			Point: gpx.Point{
				Latitude:  path[i].Lat,
				Longitude: path[i].Lng,
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

func (s *S2Store) adaptEdges(path []*Location, simplifiedPathIds []int64) {
	var beginId int64 = NIL_ID
	var lastIx = 0

	for i := 0; i < len(simplifiedPathIds); i++ {
		if i == 0 {
			beginId = simplifiedPathIds[i]
			continue
		}

		endId := simplifiedPathIds[i]

		log.Debugf("updating edge of simplified path %d -> %d", beginId, endId)

		if path[lastIx].Id != beginId {
			panic("path point doesn't match starting point of the edge")
		}
		var pointsToRemove []int64

		// loop
		for j := lastIx + 1; j < len(path); j++ {
			//log.Debugf("checking point ix=%d id=%d", j, path[j].Id)

			if path[j].Id == endId {
				log.Debugf("match found, to be removed: %v", pointsToRemove)
				s.adaptNetToEdge(beginId, endId, pointsToRemove)
				pointsToRemove = nil // make slice empty
				beginId = endId
				lastIx = j
				break
			} else {
				pointsToRemove = append(pointsToRemove, path[j].Id)
			}
		}
	}
}

func (s *S2Store) adaptNetToEdge(beginId, endId int64, toRemoveIds []int64) {

	log.Debugf("---------- adapt net to edge %d %d -----------", beginId, endId)

	if len(toRemoveIds) == 0 {
		log.Debugf("noting to remove for begin: %d and end: %d", beginId, endId)
		return
	}

	// prepare final edge (create new or reuse existing)
	finalEdgePoints := []int64{min(beginId, endId), max(beginId, endId)}
	finalEdgeId := edgeIdFromPointIds(beginId, endId)

	log.Debugf("final edge id: %v", finalEdgeId)
	finalEdge := s.getEdgeById(finalEdgeId)

	if finalEdge == nil {
		log.Debugf("creating new final edge, id: %v", finalEdgeId)

		finalEdge = &S2Edge{
			Id:     finalEdgeId,
			Points: finalEdgePoints,
			Tracks: []int64{},
			// Count: 0
		}
		s.edges[finalEdgeId] = finalEdge
	} else {
		log.Debugf("reusing existing edge, id: %v", finalEdgeId)
	}

	// get first edge to be removed
	firstEdgeId := edgeIdFromPointIds(beginId, toRemoveIds[0])
	log.Debugf("first edge (to be removed) id:: %v", firstEdgeId)
	firstEdge := s.getEdgeById(firstEdgeId)

	if firstEdge == nil {
		log.Exitf("edge %v not found", firstEdgeId)
	}

	// update final edge with properties of the first edge, copying slices is important
	log.Debugf("tracks before merge: %v %v", finalEdge.Tracks, firstEdge.Tracks)
	finalEdge.Tracks = append(finalEdge.Tracks, firstEdge.Tracks...)
	log.Debugf("merged tracks: %v", finalEdge.Tracks)

	//finalEdge.Count = firstEdge.Count

	// delete all edges
	var allIds = append(append([]int64{beginId}, toRemoveIds...), endId)
	var edgeIdsToRemove []S2EdgeKey
	log.Debugf("all ids: %v", allIds)
	for i := 1; i < len(allIds); i++ {
		edgeId := edgeIdFromPointIds(allIds[i-1], allIds[i])
		edgeIdsToRemove = append(edgeIdsToRemove, edgeId)
	}

	// delete redundant points
	log.Debugf("points to be deleted %v", toRemoveIds)
	s.index.RemoveByIds(toRemoveIds)
	s.Stat.PointsSimplified += int64(len(toRemoveIds))

	// delete redundant edges
	log.Debugf("edges to be deleted %v", edgeIdsToRemove)
	s.removeEdgesByIds(edgeIdsToRemove)
	s.Stat.EdgesSimplified += int64(len(edgeIdsToRemove))

}
