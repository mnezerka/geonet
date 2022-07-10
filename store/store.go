package store

import (
	"github.com/apex/log"
	"github.com/ptrv/go-gpx"
)

type Store struct {
	Hulls  map[HullId]*Hull4
	Lines  []*Line
	Log    *log.Entry
	LastId HullId
}

func NewStore() *Store {
	s := Store{}

	s.Log = log.WithFields(log.Fields{
		"comp": "store",
	})

	s.Log.Info("creating store")

	s.LastId = 0
	s.Hulls = make(map[HullId]*Hull4)

	return &s
}

func (s *Store) GetId() HullId {
	s.LastId++
	return s.LastId
}

func (s *Store) AddPoint(p gpx.Wpt, sizeTreshold float64) *Hull4 {
	s.Log.Infof("add point [%f, %f]", p.Lat, p.Lon)

	new := NewHull4FromVector(Vector{p.Lat, p.Lon}, s.GetId())

	// look for existing hull which is close to the point
	var result *Hull4 = nil
	for id, h := range s.Hulls {

		joined := h.Add(*new)

		// if joined size is under threshold
		if joined.Size() < sizeTreshold {
			s.Log.Infof("  adding to existing hull (%d), size %f ", joined.Id, joined.Size())
			s.Hulls[id] = &joined
			result = &joined
			break
		}
	}

	// no existing hull was close -> create new one
	if result == nil {
		s.Log.Infof("  registering new hull (%d)", new.Id)
		s.Hulls[new.Id] = new
		result = new
	}

	return result
}

func (s *Store) AddLine(lastHull, targetHull *Hull4) {
	s.Log.Info("add line")

	newLine := &Line{lastHull.Id, targetHull.Id}

	// check if line already exists
	var line *Line = nil
	for i := 0; i < len(s.Lines); i++ {
		if s.Lines[i].Equals(newLine) {
			line = s.Lines[i]
			break
		}
	}

	if line == nil {
		s.Log.Info("  new line registered")
		s.Lines = append(s.Lines, newLine)
	} else {
		s.Log.Info("  updating existing line")
	}

}

func (s *Store) AddGpx(gpx *gpx.Gpx, sizeTreshold float64) {
	s.Log.Info("add gpx")

	// tracks
	for i := 0; i < len(gpx.Tracks); i++ {

		s.Log.Infof("  track: %s", gpx.Tracks[i].Name)

		// track segments
		var lastHull *Hull4 = nil
		for j := 0; j < len(gpx.Tracks[i].Segments); j++ {
			s.Log.Infof("    segment with %d points", len(gpx.Tracks[i].Segments[j].Waypoints))

			// segment points
			for k := 0; k < len(gpx.Tracks[i].Segments[j].Waypoints); k++ {
				targetHull := s.AddPoint(gpx.Tracks[i].Segments[j].Waypoints[k], sizeTreshold)

				if lastHull != nil {
					s.AddLine(lastHull, targetHull)
				}

				lastHull = targetHull
			}
		}
	}

	s.Log.Info("gpx added")
	s.Log.Infof("  hulls: %d", len(s.Hulls))
	s.Log.Infof("  lines: %d", len(s.Lines))
}
