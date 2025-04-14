package s2store

import (
	"mnezerka/geonet/log"
	"mnezerka/geonet/utils"
	"slices"
)

func (s *S2Store) getNextFreeSegment() []*Location {

	log.Debug("---------- get next free segment -----------")

	path := []*Location{}

	// get random free edge (free = not processed yet)
	freeEdge := s.GetFirstEdgeFiltered(func(edge *S2Edge) bool {
		return !edge.Processed
	})

	if freeEdge == nil {
		log.Debugf("no free edges found")
		return path
	}

	log.Debugf("free edge: %v, setting processed -> true", freeEdge.Id)
	freeEdge.Processed = true

	// let's start building path from freePoint
	path = append(path, s.index.GetLocation(freeEdge.Id.P1), s.index.GetLocation(freeEdge.Id.P2))

	// legs are paths starting at freePoint
	for leg := 0; leg < 2; leg++ {
		log.Debugf("traversing leg %d with path %v", leg+1, pointsToIds(path))

		//if len(neighbourIds) > 0 {
		nextPointExists := true
		for nextPointExists {
			nextPoint := s.getNextPoint(path)
			if nextPoint != nil {
				path = append(path, nextPoint)
				// set appropriate edge to processed
				lastEdgeId := edgeIdFromPointIds(path[len(path)-1].Id, path[len(path)-2].Id)
				lastEdge := s.getEdgeById(lastEdgeId)
				if lastEdge != nil {
					log.Debugf("mark edge: %v, setting processed -> true", lastEdgeId)
					lastEdge.Processed = true
				} else {
					log.Exitf("inconsistent data, edge not found: %v", lastEdgeId)
				}
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
	// there is an exception for path of length 1 - it can happen that first picked
	// point was begin or end of track, and we have to process it correctly - case: track of two
	// points where one is "begin" and second is "end"
	if len(path) > 1 && (p.Begin || p.End || p.Crossing) {
		log.Debugf("last path point %d is begin | end | crossing -> path is complete", p.Id)
		return nil
	}

	// else we have another free point that is being processed
	neighbourIds := s.findNeighbours(*p)
	log.Debugf("all neighbours: %v", neighbourIds)
	neighbourFreeIds := s.index.GetLocationsFiltered(func(l *Location) bool {
		return !l.Begin && !l.End && !l.Crossing && slices.Contains(neighbourIds, l.Id)
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

	return result
}

// look in not processed edges only
func (s *S2Store) findPointEdges(id int64) []S2EdgeKey {
	result := []S2EdgeKey{}

	for edgeId, edge := range s.edges {
		if !edge.Processed && (edgeId.P1 == id || edgeId.P2 == id) {
			result = append(result, edgeId)
		}
	}

	return result
}
