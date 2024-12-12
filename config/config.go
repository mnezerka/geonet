package config

type Configuration struct {
	SimplifyMinDistance   int64 // in meters
	MatchMaxDistance      int64 // in meters
	InterpolationDistance int64 // in meters
	ShowPoints            bool  // render points to map
	ShowEdges             bool  // render edges to map
	ShowTrackColors       bool  // render tracks with different colors
}

var Cfg = Configuration{
	SimplifyMinDistance:   50,
	MatchMaxDistance:      75,
	InterpolationDistance: 30,
	ShowPoints:            false,
	ShowEdges:             true,
	ShowTrackColors:       false,
}
