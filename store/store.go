package store

import (
	"github.com/apex/log"
	"github.com/ptrv/go-gpx"
)

type StoreId int64

type Store struct {
	Hulls  map[StoreId]*Hull4
	Lines  map[StoreId]*Line
	Log    *log.Entry
	LastId StoreId
}

func NewStore() *Store {
	s := Store{}

	s.Log = log.WithFields(log.Fields{
		"comp": "store",
	})

	s.Log.Info("creating store")

	s.LastId = 0
	s.Hulls = make(map[StoreId]*Hull4)
	s.Lines = make(map[StoreId]*Line)

	return &s
}

func (s *Store) GetId() StoreId {
	s.LastId++
	return s.LastId
}

func (s *Store) AddPoint(p gpx.Wpt, sizeTreshold float64) *Hull4 {
	s.Log.Infof("add point [%f, %f]", p.Lat, p.Lon)

	new := NewHull4FromVector(Vector{p.Lat, p.Lon}, s.GetId())

	// look for existing hull which is close to the point
	var result *Hull4 = nil
	for id, h := range s.Hulls {

		// try to join existing hull with the new point
		joined := h.Add(*new)

		// if joined size is under threshold
		if joined.Size() < sizeTreshold {
			s.Log.Infof("  adding to existing hull (%d), size %f ", joined.Id, joined.Size())
			s.Hulls[id] = &joined
			result = &joined
			break
		}
	}

	// no existing hull was close enough -> create new one
	if result == nil {
		s.Log.Infof("  registering new hull (%d)", new.Id)
		s.Hulls[new.Id] = new
		result = new

		// try to find a line, which is close by
		lineInRange := s.LineInHullRange(new, sizeTreshold)
		if lineInRange != nil {
			s.Log.Infof("  close to line %s", lineInRange.Str())

			//         lineInRange
			//   a -------------------------- b
			//            |
			//            |
			//            x new center
			//            |
			//            |
			//            new

			// move new hull to the middle distance between line and it's position
			ha := s.Hulls[lineInRange.A]
			hb := s.Hulls[lineInRange.B]
			ca := ha.BoundRect().Center()
			cb := hb.BoundRect().Center()

			// center of new hull
			cnew := new.BoundRect().Center()

			s.Log.Infof("    cnew: %v", cnew)

			// get projection of new hull to the line
			proj := Projection(ca, cb, cnew)
			s.Log.Infof("    proj: %v", proj)

			// vectror from new to projection point
			// and shrink newproj to half length
			move := VectorFromPoints(proj, cnew).Scalar(.5)

			s.Log.Infof("    move: %v", move)

			center := VectorFromPoint(cnew).Sub(move)

			s.Log.Infof("    center: %v", center)

			// create new hull with the same id
			new = NewHull4FromVector(center, new.Id)
			s.Hulls[new.Id] = new
			result = new

			s.Log.Infof("    new hull moved to %v", new.BoundRect().Center())

			// split line to two segments
			// TODO: lineInRange.A.
			s.Log.Infof("    splitting %s", lineInRange.Str())
			newLine := &Line{new.Id, lineInRange.B, s.GetId()}
			s.Lines[newLine.Id] = newLine
			lineInRange.B = new.Id

			s.Log.Infof("      new line registered %s", newLine.Str())
			s.Log.Infof("      existing line changed to %s", lineInRange.Str())
		}
	}

	return result
}

func (s *Store) AddLine(lastHull, targetHull *Hull4) {
	s.Log.Info("add line")

	newLine := &Line{lastHull.Id, targetHull.Id, s.GetId()}

	// check if line already exists
	var line *Line = nil
	for _, l := range s.Lines {
		if l.Equals(newLine) {
			line = l
			break
		}
	}

	if line == nil {
		s.Log.Infof("  new line registered (%d)", newLine.Id)
		s.Lines[newLine.Id] = newLine
	} else {
		s.Log.Infof("  updating existing line (%d)", line.Id)
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

func (s *Store) LineInHullRange(h *Hull4, sizeTreshold float64) *Line {

	var result *Line = nil
	min := sizeTreshold

	for _, line := range s.Lines {
		ha := s.Hulls[line.A]
		hb := s.Hulls[line.B]
		ca := ha.BoundRect().Center()
		cb := hb.BoundRect().Center()

		d, p := Distance(ca, cb, h.BoundRect().Center())
		// if perpendicular projection lays on line and distance is in range
		if p && d < min {
			result = line
			min = d
		}
	}

	return result
}
