package s2store

import (
	"github.com/jedib0t/go-pretty/v6/table"
)

func boolToStr(x bool) string {
	if x {
		return "X"
	}
	return ""
}

func mapKeys(m map[int64]*S2Edge) []int64 {
	result := []int64{}

	for key := range m {
		result = append(result, key)
	}

	return result
}

func (s *S2Store) ToTxt() string {

	t := table.NewWriter()
	t.AppendHeader(table.Row{"Point", "Tracks", "Neighbours", "Begin", "End", "Crossing"})

	points := s.index.GetLocations()
	//var neighbours []int64

	for _, point := range points {

		t.AppendRow(table.Row{
			point.Id,
			point.Tracks,
			mapKeys(point.Edges),
			boolToStr(point.Begin),
			boolToStr(point.End),
			boolToStr(point.Crossing),
		})

	}

	return t.Render()
}
