package s2store

import (
	"slices"
	"sort"

	"github.com/golang/geo/s1"
	"github.com/golang/geo/s2"
)

type Location struct {
	Id        int64             `json:"id"`
	Lat       float64           `json:"lat"`
	Lng       float64           `json:"lng"`
	Tracks    []int64           `json:"tracks"`
	Count     int               `json:"-"`
	Begin     bool              `json:"begin"`
	End       bool              `json:"end"`
	Crossing  bool              `json:"crossing"`
	Processed bool              `json:"-"`
	Edges     map[int64]*S2Edge `json:"-"`
}

type NearestResult struct {
	Location       *Location
	DistanceMeters float64
}

/////////////////////////////////////////////////////////// SpatialIndex

func NewLocation() *Location {
	l := &Location{}
	l.Edges = make(map[int64]*S2Edge)
	return l
}

type SpatialIndex struct {
	data  map[s2.CellID][]*Location
	level int // S2 cell level for indexing
	flat  map[int64]*Location
}

func NewSpatialIndex(level int) *SpatialIndex {
	return &SpatialIndex{
		data:  make(map[s2.CellID][]*Location),
		level: level,
		flat:  make(map[int64]*Location), // flat structure - mapping of Location Id -> Location object (optional attribute, not related to S2 library)
	}
}

func (si *SpatialIndex) Add(loc *Location) {
	cell := s2.CellIDFromLatLng(s2.LatLngFromDegrees(loc.Lat, loc.Lng)).Parent(si.level)
	si.data[cell] = append(si.data[cell], loc)
	si.flat[loc.Id] = loc
}

func (si *SpatialIndex) Nearest(lat, lng float64, radiusMeters float64) []NearestResult {
	queryLatLng := s2.LatLngFromDegrees(lat, lng)
	queryPoint := s2.PointFromLatLng(queryLatLng)

	// Convert radius (meters) to angle (radians)
	angle := s1.Angle(radiusMeters / 6371000) // Earth's radius in meters

	// Create a spherical cap (disc on the globe)
	cap := s2.CapFromCenterAngle(queryPoint, angle)

	// Use RegionCoverer to find CellIDs intersecting the cap
	rc := &s2.RegionCoverer{
		MinLevel: si.level,
		MaxLevel: si.level,
		MaxCells: 20,
	}
	cellUnion := rc.Covering(cap)

	var results []NearestResult
	for _, cellID := range cellUnion {
		if locs, ok := si.data[cellID]; ok {
			for _, loc := range locs {
				dist := haversineDistance(lat, lng, loc.Lat, loc.Lng)
				if dist <= radiusMeters {
					results = append(results, NearestResult{
						Location:       loc,
						DistanceMeters: dist,
					})
				}
			}
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].DistanceMeters < results[j].DistanceMeters
	})

	return results
}

func (si *SpatialIndex) Remove(loc *Location) {

	cell := s2.CellIDFromLatLng(s2.LatLngFromDegrees(loc.Lat, loc.Lng)).Parent(si.level)

	locations, ok := si.data[cell]
	if !ok {
		return
	}

	// Find and remove by exact match
	for i, l := range locations {
		if l.Id == loc.Id {
			// Remove from slice
			//si.data[cell] = append(locations[:i], locations[i+1:]...)
			si.data[cell] = slices.Delete(locations, i, i+1)
			break
		}
	}

	// Clean up empty cell list
	if len(si.data[cell]) == 0 {
		delete(si.data, cell)
	}

	delete(si.flat, loc.Id)
}

func (si *SpatialIndex) RemoveByIds(ids []int64) {
	for _, loc := range si.flat {
		if slices.Contains(ids, loc.Id) {
			si.Remove(loc)
			delete(si.flat, loc.Id)
		}
	}
}

func (si *SpatialIndex) GetLocations() map[int64]*Location {
	return si.flat
}

func (si *SpatialIndex) GetLocationsByIds(ids []int64) []*Location {
	result := []*Location{}

	for _, loc := range si.flat {
		if slices.Contains(ids, loc.Id) {
			result = append(result, loc)
		}
	}
	return result
}

func (si *SpatialIndex) GetLocationsFiltered(filter func(l *Location) bool) []*Location {
	result := []*Location{}

	for _, loc := range si.flat {
		if filter(loc) {
			result = append(result, loc)
		}
	}
	return result
}

func (si *SpatialIndex) GetLocation(id int64) *Location {
	l, ok := si.flat[id]
	if ok {
		return l
	}
	return nil
}

func (si *SpatialIndex) GetFirstLocation(filter func(l *Location) bool) *Location {

	for _, loc := range si.flat {
		if filter(loc) {
			return loc
		}
	}

	return nil
}

func (si *SpatialIndex) SetLocationsNotProcessed() {
	for _, loc := range si.flat {
		loc.Processed = false
	}
}
