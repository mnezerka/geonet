package s2store

import (
	"mnezerka/geonet/config"
	"mnezerka/geonet/tracks"
	"mnezerka/geonet/utils"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimplifySingleTrack(t *testing.T) {

	cfg := config.Cfg
	cfg.SimplifyMinDistance = 2000 // 2km

	// log.SetVerbosity(log.LOG_LEVEL_DEBUG)

	s := NewS2Store(&cfg)

	track := tracks.NewTrack("../test_data/t1.gpx")
	assert.Nil(t, s.AddGpx(track))

	s.Simplify()

	locIds := utils.MapKeys(s.index.GetLocations())
	assert.Len(t, locIds, 2)

	// add both directions
	expectedLocations := [][]int64{
		{1, 7},
		{7, 1},
	}

	assert.Contains(t, expectedLocations, locIds)

	edges := s.GetEdgesFiltered(nil)
	assert.Len(t, edges, 1)
}
