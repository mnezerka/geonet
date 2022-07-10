package store

import (
	"github.com/apex/log"
	"github.com/ptrv/go-gpx"
)

type Store struct {
	Hulls []Hull4
	Log   *log.Entry
}

func NewStore() *Store {
	s := Store{}
	//s.Hulls = []

	s.Log = log.WithFields(log.Fields{
		"comp": "store",
	})

	s.Log.Info("creating store")

	return &s
}

func (s *Store) AddPoint(p gpx.Wpt, sizeTreshold float64) {
	s.Log.Infof("add point [%f, %f]", p.Lat, p.Lon)

	h := NewHull4FromVector(Vector{p.Lat, p.Lon})

	// look for existing hull which is close to the point
	found := false
	for i := 0; i < len(s.Hulls); i++ {
		joined := s.Hulls[i].Add(*h)

		// if joined size is under threshold
		if joined.Size() < sizeTreshold {
			s.Log.Infof("adding to existing hull, size %f ", joined.Size())
			s.Hulls[i] = joined
			found = true
			break
		}
	}

	// no existing hull was close -> create new one
	if !found {
		s.Log.Infof("creating new hull from %v", h)
		s.Hulls = append(s.Hulls, *h)
	}

	//s.Log.Infof("current hull: %v", s.Hulls[0])
}

func (s *Store) AddGpx(gpx *gpx.Gpx, sizeTreshold float64) {
	s.Log.Info("add gpx")

	// tracks
	for i := 0; i < len(gpx.Tracks); i++ {

		s.Log.Infof("  track: %s", gpx.Tracks[i].Name)

		// track segments
		for j := 0; j < len(gpx.Tracks[i].Segments); j++ {
			s.Log.Infof("    segment with %d points", len(gpx.Tracks[i].Segments[j].Waypoints))

			// segment points
			for k := 0; k < len(gpx.Tracks[i].Segments[j].Waypoints); k++ {
				s.AddPoint(gpx.Tracks[i].Segments[j].Waypoints[k], sizeTreshold)
			}
		}
	}

	s.Log.Info("gpx added")
	s.Log.Infof("  hulls: %d", len(s.Hulls))
}
