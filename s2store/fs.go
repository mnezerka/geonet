package s2store

import (
	"encoding/json"
	"fmt"
	"mnezerka/geonet/log"
	"mnezerka/geonet/store"
	"os"
)

type S2StoreJson struct {
	Tracks      []*store.Track `json:"tracks"`
	Edges       []*S2Edge      `json:"edges"`
	Locations   []*Location    `json:"locations"`
	LastPointId int64          `json:"last-point-id"`
	LastTrackId int64          `json:"last-track-id"`
}

func (s *S2Store) Save() {

	toJson := S2StoreJson{}

	toJson.LastPointId = s.lastPointId
	toJson.LastTrackId = s.lastTrackId

	for _, t := range s.tracks {
		toJson.Tracks = append(toJson.Tracks, t)
		s.stat.TracksRendered = int64(len(s.tracks))
	}

	for _, e := range s.edges {
		toJson.Edges = append(toJson.Edges, e)
		s.stat.EdgesRendered = int64(len(s.edges))
	}

	for _, l := range s.index.flat {
		toJson.Locations = append(toJson.Locations, l)
		s.stat.PointsRendered = int64(len(s.index.flat))
	}

	// meta - json
	sJson, err := json.MarshalIndent(toJson, "", " ")
	if err != nil {
		log.ExitWithError(err)
	}

	fmt.Print(string(sJson))
}

func (s *S2Store) Load(filePath string) {

	// Read the JSON file
	file, err := os.Open(filePath)
	if err != nil {
		log.ExitWithError(err)
	}
	defer file.Close()

	fromJson := S2StoreJson{}

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&fromJson)
	if err != nil {
		log.ExitWithError(err)
	}

	for _, t := range fromJson.Tracks {
		s.tracks[t.Id] = t
	}
	s.stat.TracksLoaded = int64(len(fromJson.Tracks))

	for _, l := range fromJson.Locations {
		s.index.Add(l)
		l.Edges = make(map[int64]*S2Edge)
	}
	s.stat.PointsLoaded = int64(len(fromJson.Locations))

	for _, e := range fromJson.Edges {

		s.AddEdge(e)
	}
	s.stat.EdgesLoaded = int64(len(fromJson.Edges))

	s.lastPointId = fromJson.LastPointId
	s.lastTrackId = fromJson.LastTrackId
}
