package tracks

type TrackMeta struct {
	SourceId string   `json:"sourceid" bson:"sourceid"`
	Title    string   `json:"title" bson:"title"`
	Url      string   `json:"url" bson:"url"`
	Creators []string `json:"creators" bson:"creators"`
}

type WandererTrack struct {
	Url string
}

type WandererPost struct {
	Title  string
	Tracks []WandererTrack
	Url    string
}
