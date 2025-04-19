package store

import (
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
)

type Stat struct {
	TracksLoaded       int64
	TracksProcessed    int64
	TracksRendered     int64
	PointsLoaded       int64
	PointsGpx          int64
	PointsCreated      int64
	PointsSimplified   int64
	PointsReused       int64
	PointsFinal        int64
	PointsRendered     int64
	EdgesLoaded        int64
	EdgesCreated       int64
	EdgesReused        int64
	EdgesSimplified    int64
	EdgesFinal         int64
	EdgesRendered      int64
	SegmentsProcessed  int64
	SegmentsSimplified int64
	SegmentsRendered   int64
}

func (s Stat) Print() {

	t := table.NewWriter()
	t.SetOutputMirror(os.Stderr)
	t.AppendHeader(table.Row{"Entity", "Loaded", "GPX", "Reused", "Created", "Simplified", "Final", "Rendered"})
	t.AppendRows([]table.Row{{
		"tracks",
		s.TracksLoaded,
		s.TracksProcessed,
		"-",
		"-",
		"-",
		"-",
		s.TracksRendered,
	}, {
		"points",
		s.PointsLoaded,
		s.PointsGpx,
		s.PointsReused,
		s.PointsCreated,
		s.PointsSimplified,
		s.PointsFinal,
		s.PointsRendered,
	}, {
		"edges",
		s.EdgesLoaded,
		"-",
		s.EdgesReused,
		s.EdgesCreated,
		s.EdgesSimplified,
		s.EdgesFinal,
		s.EdgesRendered,
	}, {
		"segments",
		"-",
		s.SegmentsProcessed,
		"-",
		"-",
		s.SegmentsSimplified,
		"-",
		s.SegmentsRendered,
	}})
	t.Render()
}
