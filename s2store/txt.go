package s2store

import (
	"mnezerka/geonet/utils"

	"github.com/jedib0t/go-pretty/v6/table"
)

func boolToStr(x bool) string {
	if x {
		return "X"
	}
	return ""
}

func (s *S2Store) ToTxt() string {

	t := table.NewWriter()
	t.AppendHeader(table.Row{"Point", "Tracks", "Neighbours", "Begin", "End", "Crossing"})

	points := s.index.GetLocations()
	//var neighbours []int64

	for _, point := range points {

		t.AppendRow(table.Row{
			point.Id,
			utils.MapKeys(point.Tracks),
			utils.MapKeys(point.Edges),
			boolToStr(point.Begin),
			boolToStr(point.End),
			boolToStr(point.Crossing),
		})

	}

	return t.Render()
}
