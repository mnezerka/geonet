package s2store

import (
	"encoding/json"
	"fmt"
	"mnezerka/geonet/log"
	"mnezerka/geonet/store"
	"os"
)

type S2StoreJson struct {
	Tracks    []*store.Track `json:"tracks"`
	Edges     []*S2Edge      `json:"edges"`
	Locations []*Location    `json:"locations"`
}

func (s *S2Store) Save() {

	toJson := S2StoreJson{}

	for _, t := range s.tracks {
		toJson.Tracks = append(toJson.Tracks, t)
	}

	for _, e := range s.edges {
		toJson.Edges = append(toJson.Edges, e)
	}

	for _, l := range s.index.flat {
		toJson.Locations = append(toJson.Locations, l)
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

	for _, e := range fromJson.Edges {
		s.edges[e.Id] = e
	}
	s.stat.EdgesLoaded = int64(len(fromJson.Edges))

	for _, l := range fromJson.Locations {
		s.index.Add(l)
	}
	s.stat.PointsLoaded = int64(len(fromJson.Locations))
}
