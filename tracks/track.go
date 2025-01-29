package tracks

import (
	"encoding/json"
	"fmt"
	"github.com/mnezerka/gpxcli/gpxutils"
	"github.com/tkrajina/gpxgo/gpx"
	"mnezerka/geonet/log"
	"mnezerka/geonet/utils"
	"os"
)

type Track struct {
	FilePath     string
	FilePathMeta string
	Meta         TrackMeta
	Points       []gpx.GPXPoint
	gpxFile      *gpx.GPX
}

func NewTrack(filePath string) *Track {
	t := Track{FilePath: filePath}

	t.FilePathMeta = utils.SetExtension(t.FilePath, "json")

	t.ReadPoints()

	t.ReadMeta()

	return &t
}

func (t *Track) ReadMeta() {

	// do nothing if meta file doesn't exist
	if !utils.FileExists(t.FilePathMeta) {
		return
	}

	jsonFile, err := os.Open(t.FilePathMeta)
	if err != nil {
		panic(err)
	}
	defer jsonFile.Close()

	err = json.NewDecoder(jsonFile).Decode(&t.Meta)
	if err != nil {
		panic(err)
	}
}

func (t *Track) WriteMeta() {

	file, err := os.Create(t.FilePathMeta)
	if err != nil {
		panic(fmt.Errorf("error creating the meta file: %s", err))
	}
	defer file.Close()

	jsonData, err := json.MarshalIndent(t.Meta, "", "  ")
	if err != nil {
		panic(err)
	}

	_, err = file.Write(jsonData)
	if err != nil {
		panic(err)
	}
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

	log.Infof("points read from gpx file: %d", len(t.Points))
}

func (t *Track) InterpolateDistance(distance int64) {
	var err error
	t.Points, err = gpxutils.InterpolateDistance(t.Points, float64(distance))
	if err != nil {
		panic(err)
	}

	log.Infof("points interpolated: %d", len(t.Points))
}

func (t *Track) GetTitleFromContent() string {
	result := ""

	if t.gpxFile != nil {

		// find first track with valid name
		for _, track := range t.gpxFile.Tracks {
			if len(track.Name) > 0 {
				result = track.Name
				break
			}
		}

		if len(result) == 0 {
			result = t.gpxFile.Name
		}
	}

	return result
}
