package config

import (
	"fmt"
)

type Configuration struct {
	SimplifyMinDistance   int64 // in meters
	MatchMaxDistance      int64 // in meters
	InterpolationDistance int64 // in meters
	ShowPoints            bool  // render points to map
	ShowEdges             bool  // render edges to map
	ShowTrackColors       bool  // render tracks with different colors
	GeoJsonMergeEdges     bool  // export edge segments ans continuous line instead of individual lines
}

func (c *Configuration) ToString() string {
	return fmt.Sprintf("int-dist: %d, simplify-min-dist: %d, match-max-dist: %d",
		c.InterpolationDistance,
		c.SimplifyMinDistance,
		c.MatchMaxDistance,
	)
}

var Cfg = Configuration{
	SimplifyMinDistance:   50,
	MatchMaxDistance:      75,
	InterpolationDistance: 30,
	ShowPoints:            false,
	ShowEdges:             true,
	ShowTrackColors:       false,
	GeoJsonMergeEdges:     false,
}
