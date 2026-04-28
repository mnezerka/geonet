package s2store

import (
	"mnezerka/geonet/config"
	"mnezerka/geonet/tracks"
	"testing"
)

func BenchmarkAddGpx(b *testing.B) {
	cfg := config.Configuration{
		MatchMaxDistance:      75,
		SimplifyMinDistance:   50,
		InterpolationDistance: 30,
	}
	track := tracks.NewTrack("../test_data/Lunch_Ride.gpx")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s := NewS2Store(&cfg)
		s.AddGpx(track)
	}
}

func BenchmarkAddGpxMultiple(b *testing.B) {
	cfg := config.Configuration{
		MatchMaxDistance:      75,
		SimplifyMinDistance:   50,
		InterpolationDistance: 30,
	}
	files := []string{
		"../test_data/Lunch_Ride.gpx",
		"../test_data/Smelcovna.gpx",
		"../test_data/libava1.gpx",
		"../test_data/Kras_a_prehrada.gpx",
		"../test_data/mum.gpx",
	}
	parsedTracks := make([]*tracks.Track, len(files))
	for i, f := range files {
		parsedTracks[i] = tracks.NewTrack(f)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s := NewS2Store(&cfg)
		for _, t := range parsedTracks {
			s.AddGpx(t)
		}
	}
}

func BenchmarkSimplify(b *testing.B) {
	cfg := config.Configuration{
		MatchMaxDistance:      75,
		SimplifyMinDistance:   50,
		InterpolationDistance: 30,
	}
	files := []string{
		"../test_data/Lunch_Ride.gpx",
		"../test_data/Smelcovna.gpx",
		"../test_data/libava1.gpx",
	}
	parsedTracks := make([]*tracks.Track, len(files))
	for i, f := range files {
		parsedTracks[i] = tracks.NewTrack(f)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s := NewS2Store(&cfg)
		for _, t := range parsedTracks {
			s.AddGpx(t)
		}
		s.Simplify()
	}
}

func BenchmarkSegmentation(b *testing.B) {
	cfg := config.Configuration{
		MatchMaxDistance:      75,
		SimplifyMinDistance:   50,
		InterpolationDistance: 30,
	}
	s := NewS2Store(&cfg)
	for _, f := range []string{
		"../test_data/Lunch_Ride.gpx",
		"../test_data/Smelcovna.gpx",
		"../test_data/libava1.gpx",
	} {
		t := tracks.NewTrack(f)
		s.AddGpx(t)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.setEdgesNotProcessed()
		for {
			path := s.getNextFreeSegment()
			if len(path) < 2 {
				break
			}
		}
	}
}
