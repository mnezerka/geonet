package s2store

import (
	"mnezerka/geonet/config"
	"mnezerka/geonet/log"
	"mnezerka/geonet/store"
	"mnezerka/geonet/tracks"
	"slices"
)

const NIL_ID = -1

type S2EdgeKey struct {
	P1 int64 `json:"p1"`
	P2 int64 `json:"p2"`
}

type S2Edge struct {
	Id        S2EdgeKey `json:"id"`
	Tracks    []int64   `json:"tracks"`
	Processed bool      `json:"-"`
}

type S2Store struct {
	cfg         *config.Configuration
	index       *SpatialIndex
	lastPointId int64
	lastTrackId int64
	tracks      map[int64]*store.Track
	edges       map[S2EdgeKey]*S2Edge
	stat        store.Stat
}

func NewS2Store(cfg *config.Configuration) *S2Store {
	s := S2Store{}

	s.cfg = cfg
	s.index = NewSpatialIndex(15)
	s.tracks = make(map[int64]*store.Track)
	s.edges = make(map[S2EdgeKey]*S2Edge)

	return &s
}

func (s *S2Store) GenPointId() int64 {
	s.lastPointId++
	return s.lastPointId
}

func (s *S2Store) GenTrackId() int64 {
	s.lastTrackId++
	return s.lastTrackId
}

func (s *S2Store) GetStat() store.Stat {
	return s.stat
}

func (s *S2Store) AddGpx(track *tracks.Track) error {
	var lastPointId int64 = NIL_ID
	var finalPointId int64 = NIL_ID

	log.Debugf("adding %s to the rtree store", track.Meta.PostTitle)

	s2Track := store.Track{
		Id:   s.GenTrackId(),
		Meta: track.Meta,
	}

	s.tracks[s2Track.Id] = &s2Track
	log.Debugf("registered track %d", s2Track.Id)
	s.stat.TracksProcessed++

	for i := 0; i < len(track.Points); i++ {

		// --------------------  track point processing
		s.stat.PointsProcessed++

		point := track.Points[i]
		isBegin := i == 0
		isEnd := i == len(track.Points)-1

		nearest := s.index.Nearest(point.Latitude, point.Longitude, float64(s.cfg.MatchMaxDistance))

		if len(nearest) > 0 {
			log.Debugf("reusing point %d %.1fm", nearest[0].Location.Id, nearest[0].DistanceMeters)
			s.stat.PointsReused++

			// update isBeign, isEnd, tracks
			nearest[0].Location.Begin = nearest[0].Location.Begin || isBegin
			nearest[0].Location.End = nearest[0].Location.End || isEnd
			if !slices.Contains(nearest[0].Location.Tracks, s2Track.Id) {
				nearest[0].Location.Tracks = append(nearest[0].Location.Tracks, s2Track.Id)
			}

			finalPointId = nearest[0].Location.Id
		} else {
			finalPointId = s.GenPointId()

			log.Debugf("adding point %d", finalPointId)
			loc := Location{
				Id:     finalPointId,
				Lat:    point.Latitude,
				Lng:    point.Longitude,
				Tracks: []int64{s2Track.Id},
				Begin:  isBegin,
				End:    isEnd,
			}
			s.index.Add(&loc)
			s.stat.PointsCreated++
			finalPointId = loc.Id
		}

		// --------------------  edge processing

		// ignore self edges (in case point was reused)
		if lastPointId != NIL_ID && lastPointId != finalPointId {
			// create edge with sorted point ids to avoid duplicates (reverse direction of track movement)
			edgeId := edgeIdFromPointIds(lastPointId, finalPointId)

			log.Debugf("edge id: %v", edgeId)

			//edge, ok := s.edges[edgeId]
			edge, ok := s.edges[edgeId]
			if ok {
				log.Debugf("reusing existing edge: %v", edgeId)
				s.stat.EdgesReused++
				if !slices.Contains(edge.Tracks, s2Track.Id) {
					edge.Tracks = append(edge.Tracks, s2Track.Id)
				}
			} else {
				log.Debugf("registering new edge: %v", edgeId)
				s.stat.EdgesCreated++
				edge := S2Edge{
					Id:     edgeId,
					Tracks: []int64{s2Track.Id},
				}
				s.edges[edgeId] = &edge

				// new edge => some point could become a crossing
				s.updateCrossingForEdgePoints(edgeId)
			}
		}
		// remember current point id for next iteration (for edge construction)
		lastPointId = finalPointId
	}

	return nil
}

func (s *S2Store) GetEdgesFiltered(filter func(l *S2Edge) bool) []*S2Edge {
	result := []*S2Edge{}

	for _, edge := range s.edges {
		if filter(edge) {
			result = append(result, edge)
		}
	}
	return result
}

func (s *S2Store) GetFirstEdgeFiltered(filter func(l *S2Edge) bool) *S2Edge {
	for _, edge := range s.edges {
		if filter(edge) {
			return edge
		}
	}
	return nil
}

func (s *S2Store) getEdgesForPointId(id int64) []*S2Edge {
	result := []*S2Edge{}

	for edgeId, edge := range s.edges {

		if edgeId.P1 == id || edgeId.P2 == id {
			result = append(result, edge)
		}
	}

	return result
}

func (s *S2Store) removeEdgesByIds(ids []S2EdgeKey) {
	for _, edgeId := range ids {
		delete(s.edges, edgeId)
	}
}

func (s *S2Store) updateCrossingForEdgePoints(edgeId S2EdgeKey) {
	// first point
	edges := s.getEdgesForPointId(edgeId.P1)
	if len(edges) > 2 {
		// set P1 as crossing
		p1 := s.index.GetLocation(edgeId.P1)
		if p1 == nil {
			log.Errorf("inconsistent data, missing point %d for edge %v", edgeId.P1, edgeId)
		}
		p1.Crossing = true
	}

	// second point
	edges = s.getEdgesForPointId(edgeId.P2)
	if len(edges) > 2 {
		// set P2 as crossing
		p2 := s.index.GetLocation(edgeId.P2)
		if p2 == nil {
			log.Errorf("inconsistent data, missing point %d for edge %v", edgeId.P2, edgeId)
		}
		p2.Crossing = true
	}
}

func (s *S2Store) GetMeta() store.Meta {
	var meta store.Meta

	for _, t := range s.tracks {
		meta.Tracks = append(meta.Tracks, t)
	}

	return meta
}

func (s *S2Store) setEdgesNotProcessed() {
	for _, e := range s.edges {
		e.Processed = false
	}
}
