package store

type SpotRecord struct {
	Id string
}

type Spot struct {
	records map[string]*SpotRecord
}

func NewSpot() *Spot {
	spot := Spot{}
	spot.records = make(map[string]*SpotRecord)
	return &spot
}

func NewSpotRecords() *SpotRecord {
	record := SpotRecord{}
	return &record
}

func (s *Spot) GetRecords() *(map[string]*SpotRecord) {
	return &s.records
}

func (s *Spot) AddRecord(new *SpotRecord) {

	if new == nil {
		return
	}

	_, exists := s.records[new.Id]

	if !exists {
		s.records[new.Id] = new
	}
}
