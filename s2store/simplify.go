package s2store

import (
	"mnezerka/geonet/log"
	"mnezerka/geonet/utils"
	"strconv"

	"github.com/mnezerka/gpxcli/gpxutils"
	"github.com/tkrajina/gpxgo/gpx"
)

func (s *S2Store) Simplify() {
	log.Debug("======================== simplify =========================")
	log.Debugf("minimal simplify distance: %d", s.cfg.SimplifyMinDistance)

	// reset all points to not processed state to be sure we start with clean setup
	//s.index.SetLocationsNotProcessed()
	s.setEdgesNotProcessed()

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
		s.stat.SegmentsProcessed++
		s.stat.SegmentsSimplified++
	}
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

/*
path:              1 --------2 -------- 3 -- 4 --------- 5 ----6
simplifiedPathIds: 1 ------------------------------------------6
*/
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

	log.Debugf("---------- adapt net to edge %d-%d + remove points %v -----------", beginId, endId, toRemoveIds)

	if len(toRemoveIds) == 0 {
		log.Debugf("noting to remove for begin: %d and end: %d", beginId, endId)
		return
	}

	// prepare final edge (create new or reuse existing)
	finalEdgeId := edgeIdFromPointIds(beginId, endId)

	log.Debugf("final edge id: %v", finalEdgeId)
	finalEdge := s.getEdgeById(finalEdgeId)

	if finalEdge == nil {
		log.Debugf("creating new final edge, id: %v", finalEdgeId)

		finalEdge = NewS2Edge()
		finalEdge.Id = finalEdgeId
		s.AddEdge(finalEdge)

		s.edges[finalEdgeId] = finalEdge
	} else {
		log.Debugf("reusing existing edge, id: %v", finalEdgeId)
	}

	// get first edge which is planned to be removed to have source
	// of attributes (e.g. list of tracks) to be assigned to final edge
	firstEdgeId := edgeIdFromPointIds(beginId, toRemoveIds[0])
	log.Debugf("first edge (to be removed) id:: %v", firstEdgeId)
	firstEdge := s.getEdgeById(firstEdgeId)

	if firstEdge == nil {
		log.Exitf("edge %v not found", firstEdgeId)
	}

	// update final edge with properties of the first edge
	log.Debugf("tracks before merge: %v %v", finalEdge.Tracks, firstEdge.Tracks)
	utils.MapsMerge(finalEdge.Tracks, firstEdge.Tracks)
	log.Debugf("merged tracks: %v", finalEdge.Tracks)

	// delete all edges one by one
	var allIds = append(append([]int64{beginId}, toRemoveIds...), endId)
	var edgeIdsToRemove []S2EdgeKey
	log.Debugf("all ids: %v", allIds)

	for i := 1; i < len(allIds); i++ {
		edgeId := edgeIdFromPointIds(allIds[i-1], allIds[i])
		edgeIdsToRemove = append(edgeIdsToRemove, edgeId)
	}

	// delete redundant edges
	log.Debugf("edges to be deleted %v", edgeIdsToRemove)
	s.removeEdgesByIds(edgeIdsToRemove)
	s.stat.EdgesSimplified += int64(len(edgeIdsToRemove))

	// delete redundant points
	log.Debugf("points to be deleted %v", toRemoveIds)
	s.index.RemoveByIds(toRemoveIds)
	s.stat.PointsSimplified += int64(len(toRemoveIds))

}
