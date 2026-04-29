package tracks

import (
	"mnezerka/geonet/log"
	"mnezerka/geonet/utils"
	"time"

	"github.com/mnezerka/gpxcli/gpxutils"
	"github.com/tkrajina/gpxgo/gpx"
)

const ns = "http://geonet.bluesoft"

type TrackMeta struct {
	TrackId    string    `json:"track_id" bson:"track_id"`
	TrackUrl   string    `json:"track_url" bson:"track_url"`
	TrackTitle string    `json:"track_title" bson:"track_title"`
	SourceUrl  string    `json:"source_url" bson:"source_url"`
	SourceType string    `json:"source_type" bson:"source_type"`
	PostTitle  string    `json:"post_title" bson:"post_title"`
	PostUrl    string    `json:"post_url" bson:"post_url"`
	LengthKm   float64   `json:"length_km" bson:"length_km"`
	TrackDate  time.Time `json:"track_date" bson:"track_date"`
}
type Track struct {
	FilePath string
	Meta     TrackMeta
	Points   []gpx.GPXPoint
	gpxFile  *gpx.GPX
}

func NewTrack(filePath string) *Track {
	t := Track{FilePath: filePath}

	t.ReadPoints()

	// load data from xml extension
	if node, found := t.gpxFile.MetadataExtensions.GetNode(ns, "trackid"); found {
		t.Meta.TrackId = node.Data
	}

	if node, found := t.gpxFile.MetadataExtensions.GetNode(ns, "tracktitle"); found {
		t.Meta.TrackTitle = node.Data
	}

	if node, found := t.gpxFile.MetadataExtensions.GetNode(ns, "trackurl"); found {
		t.Meta.TrackUrl = node.Data
	}

	if node, found := t.gpxFile.MetadataExtensions.GetNode(ns, "posttitle"); found {
		t.Meta.PostTitle = node.Data
	}

	if node, found := t.gpxFile.MetadataExtensions.GetNode(ns, "posturl"); found {
		t.Meta.PostUrl = node.Data
	}

	if node, found := t.gpxFile.MetadataExtensions.GetNode(ns, "sourcetype"); found {
		t.Meta.SourceType = node.Data
	}

	if node, found := t.gpxFile.MetadataExtensions.GetNode(ns, "sourceurl"); found {
		t.Meta.SourceUrl = node.Data
	}

	// compute fields from gpx content
	t.Meta.LengthKm = t.computeLengthKm()
	t.Meta.TrackDate = t.computeTrackDate()

	// compute fields if not read from meta
	if len(t.Meta.TrackTitle) == 0 {
		t.Meta.TrackTitle = getTitleFromGpxContent(t.gpxFile, utils.GetBasename(t.FilePath))
	}
	return &t
}

func (t *Track) ReadPoints() {

	var err error

	t.gpxFile, err = gpx.ParseFile(t.FilePath)
	if err != nil {
		panic(err)
	}

	t.Points, err = gpxutils.GpxFileToPoints(t.gpxFile)
	if err != nil {
		panic(err)
	}

	log.Debugf("points read from gpx file: %d", len(t.Points))
}

func (t *Track) InterpolateDistance(distance int64) {
	var err error
	t.Points, err = gpxutils.InterpolateDistance(t.Points, float64(distance))
	if err != nil {
		panic(err)
	}
}

func getTitleFromGpxContent(gpxFile *gpx.GPX, defaultValue string) string {
	result := ""

	if gpxFile != nil {

		// find first track with valid name
		for _, track := range gpxFile.Tracks {
			if len(track.Name) > 0 {
				result = track.Name
				break
			}
		}

		if len(result) == 0 {
			result = gpxFile.Name
		}
	}

	// if name was not inserted into gpx file, use filename
	if len(result) == 0 {
		result = defaultValue
	}

	return result
}

func (t *Track) computeLengthKm() float64 {
	if len(t.Points) == 0 {
		return 0
	}

	points := make([]gpx.Point, len(t.Points))
	for i, p := range t.Points {
		points[i] = gpx.Point{
			Latitude:  p.Latitude,
			Longitude: p.Longitude,
			Elevation: p.Elevation,
		}
	}

	lengthMeters := gpx.Length3D(points)
	return lengthMeters / 1000.0
}

func (t *Track) computeTrackDate() time.Time {
	for _, p := range t.Points {
		if !p.Timestamp.IsZero() {
			return p.Timestamp
		}
	}
	return time.Time{}
}
