package s2store

import (
	"mnezerka/geonet/config"
	"mnezerka/geonet/log"
	"mnezerka/geonet/tracks"
	"testing"

	"github.com/stretchr/testify/assert"
)

func segmentToList(sg []*Location) []int64 {
	result := []int64{}
	for _, l := range sg {
		result = append(result, l.Id)
	}
	return result
}

func TestSegmentSingleTrack(t *testing.T) {

	s := NewS2Store(&config.Cfg)

	track := tracks.NewTrack("../test_data/t1.gpx")
	assert.Nil(t, s.AddGpx(track))

	// first segment has seven points - whole track
	sg := s.getNextFreeSegment()
	assert.Len(t, sg, 7)

	// no more segments exist
	sg = s.getNextFreeSegment()
	assert.Len(t, sg, 0)
}

func TestSegmentTwoTracks(t *testing.T) {

	log.SetVerbosity(log.LOG_LEVEL_DEBUG)

	s := NewS2Store(&config.Cfg)

	track := tracks.NewTrack("../test_data/t1.gpx")
	assert.Nil(t, s.AddGpx(track))

	track2 := tracks.NewTrack("../test_data/t2.gpx")
	assert.Nil(t, s.AddGpx(track2))

	// add both directions
	segments := [][]int64{
		{1, 2, 3, 4},
		{4, 3, 2, 1},
		{4, 5},
		{5, 4},
		{5, 6, 7},
		{7, 6, 5},
		{5, 9, 8},
		{8, 9, 5},
		{4, 10},
		{10, 4},
	}

	// first segment
	sg := s.getNextFreeSegment()
	assert.Contains(t, segments, segmentToList(sg))

	// second segment
	sg = s.getNextFreeSegment()
	assert.Contains(t, segments, segmentToList(sg))

	// third segment
	sg = s.getNextFreeSegment()
	assert.Contains(t, segments, segmentToList(sg))

	// fifth segment
	sg = s.getNextFreeSegment()
	assert.Contains(t, segments, segmentToList(sg))

	// fifth segment
	sg = s.getNextFreeSegment()
	assert.Contains(t, segments, segmentToList(sg))

	// no more segments exist
	sg = s.getNextFreeSegment()
	assert.Len(t, sg, 0)
}
