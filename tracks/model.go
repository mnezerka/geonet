package tracks

type TrackMeta struct {
	TrackId    string `json:"track_id" bson:"track_id"`
	TrackUrl   string `json:"track_url" bson:"track_url"`
	TrackTitle string `json:"track_title" bson:"track_title"`

	SourceUrl  string `json:"source_url" bson:"source_url"`
	SourceType string `json:"source_type" bson:"source_type"`

	PostTitle string `json:"post_title" bson:"post_title"`
	PostUrl   string `json:"post_url" bson:"post_url"`
}

type WandererTrack struct {
	Url string
}

type WandererPost struct {
	Title  string
	Tracks []WandererTrack
	Url    string
}
