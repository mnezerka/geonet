package store

import (
	"github.com/ptrv/go-gpx"

	"mnezerka/geonet/auxiliary"
)

type StoreId int64

type Store struct {
	Hulls        map[StoreId]*Hull4
	Lines        map[StoreId]*Line
	LastId       StoreId
	LinesEnabled bool
}

func NewStore() *Store {
	s := Store{}
	s.LinesEnabled = true

	auxiliary.Debugf("creating store")

	s.LastId = 0
	s.Hulls = make(map[StoreId]*Hull4)
	s.Lines = make(map[StoreId]*Line)

	return &s
}

func (s *Store) GetId() StoreId {
	s.LastId++
	return s.LastId
}

func (s *Store) AddPoint(p gpx.Wpt, sizeTreshold float64, spotRecord *SpotRecord) *Hull4 {
	auxiliary.Debugf("add point [%f, %f] %v", p.Lat, p.Lon, spotRecord)

	new := NewHull4FromVector(Vector{p.Lat, p.Lon}, s.GetId())

	// look for existing hull which is close to the point
	var result *Hull4 = nil
	for id, h := range s.Hulls {

		// try to join existing hull with the new point
		joined := h.Add(*new)

		// if joined size is under threshold
		if joined.Size() < sizeTreshold {
			auxiliary.Debugf("  adding to existing hull (%d), size %f ", joined.Id, joined.Size())
			s.Hulls[id] = &joined
			result = &joined
			break
		}
	}

	// no existing hull was close enough -> create new one
	if result == nil {
		auxiliary.Debugf("  registering new hull (%d)", new.Id)
		s.Hulls[new.Id] = new
		result = new

		if s.LinesEnabled {

			// try to find a line, which is close by
			lineInRange := s.LineInHullRange(new, sizeTreshold)
			if lineInRange != nil {
				auxiliary.Debugf("  close to line %s", lineInRange.Str())

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

				auxiliary.Debugf("    cnew: %v", cnew)

				// get projection of new hull to the line
				proj := Projection(ca, cb, cnew)
				auxiliary.Debugf("    proj: %v", proj)

				// vectror from new to projection point
				// and shrink newproj to half length
				move := VectorFromPoints(proj, cnew).Scalar(.5)

				auxiliary.Debugf("    move: %v", move)

				center := VectorFromPoint(cnew).Sub(move)

				auxiliary.Debugf("    center: %v", center)

				// create new hull with the same id
				new = NewHull4FromVector(center, new.Id)
				s.Hulls[new.Id] = new
				result = new

				auxiliary.Debugf("    new hull moved to %v", new.BoundRect().Center())

				// split line to two segments
				// TODO: lineInRange.A.
				auxiliary.Debugf("    splitting %s", lineInRange.Str())
				newLine := &Line{new.Id, lineInRange.B, s.GetId()}
				s.Lines[newLine.Id] = newLine
				lineInRange.B = new.Id

				auxiliary.Debugf("      new line registered %s", newLine.Str())
				auxiliary.Debugf("      existing line changed to %s", lineInRange.Str())
			}
		}
	}

	if spotRecord != nil {
		result.AddRecord(spotRecord)
	}

	return result
}

func (s *Store) AddLine(lastHull, targetHull *Hull4) {

	if !s.LinesEnabled {
		return
	}

	auxiliary.Debug("add line")

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
		auxiliary.Debugf("  new line registered (%d)", newLine.Id)
		s.Lines[newLine.Id] = newLine
	} else {
		auxiliary.Debugf("  updating existing line (%d)", line.Id)
	}

}

func (s *Store) AddGpx(gpx *gpx.Gpx, sizeTreshold float64, id string) {
	auxiliary.Infof("adding gpx '%s' (%d tracks)", gpx.Metadata.Name, len(gpx.Tracks))

	// tracks
	for i := 0; i < len(gpx.Tracks); i++ {

		// track segments
		var lastHull *Hull4 = nil
		for j := 0; j < len(gpx.Tracks[i].Segments); j++ {
			auxiliary.Infof("adding segment (%d points)", len(gpx.Tracks[i].Segments[j].Waypoints))

			// segment points
			for k := 0; k < len(gpx.Tracks[i].Segments[j].Waypoints); k++ {
				targetHull := s.AddPoint(gpx.Tracks[i].Segments[j].Waypoints[k], sizeTreshold, &SpotRecord{Id: id})

				if lastHull != nil {
					s.AddLine(lastHull, targetHull)
				}

				lastHull = targetHull
			}
		}
	}

	auxiliary.Infof("gpx added hulls: %d lines: %d", len(s.Hulls), len(s.Lines))

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
