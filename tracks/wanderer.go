package tracks

import (
	"encoding/json"
	"fmt"
	"mnezerka/geonet/log"
	"mnezerka/geonet/utils"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/tkrajina/gpxgo/gpx"
)

type WandererTrack struct {
	Url string
}

type WandererPost struct {
	Title  string
	Tracks []WandererTrack
	Url    string
}

func WandererReadMediaIndexFromUrl(wandererUrl string) ([]WandererPost, error) {

	result := []WandererPost{}

	urlMediaIndex, err := url.JoinPath(wandererUrl, "mediaindex.json")
	if err != nil {
		return result, err
	}

	log.Infof("fetching media index from '%s'", urlMediaIndex)

	resp, err := http.Get(urlMediaIndex)
	if err != nil {
		return result, err
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return result, err
	}

	return result, nil
}

func WandererReadMediaIndexFromFileSystem(filePath string) ([]WandererPost, error) {

	result := []WandererPost{}

	log.Infof("fetching media index from local file '%s'", filePath)

	jsonFile, err := os.Open(filePath)
	if err != nil {
		return result, err
	}
	defer jsonFile.Close()

	err = json.NewDecoder(jsonFile).Decode(&result)

	return result, err
}

func WandererStoreTrackMeta(post *WandererPost, track *WandererTrack, sourceUrl string, dir string, force bool) error {

	fileName := utils.ConvertToSafeFilename(track.Url)
	filePath := filepath.Join(dir, fileName)

	log.Infof("  track file name      :%s", filePath)
	log.Infof("  source url::%s", sourceUrl)

	// check if file exists in not-forced mode
	if !force {
		if utils.FileExists(filePath) {
			log.Info("  skipping, track file exists")
			return nil
		}
	}

	// fetch the gpx file from the URL
	resp, err := http.Get(track.Url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	gpxFile, err := gpx.Parse(resp.Body)
	if err != nil {
		return fmt.Errorf("error parsing gpx file: %s", err)
	}

	// remove empty tracks and segments + merge all tracks into one signle track
	gpxFile.RemoveEmpty()
	gpxFile.ReduceGpxToSingleTrack()

	// update metadata
	gpxFile.RegisterNamespace("gn", ns)
	gpxFile.MetadataExtensions.GetOrCreateNode(ns, "trackid").Data = fileName
	gpxFile.MetadataExtensions.GetOrCreateNode(ns, "trackurl").Data = track.Url
	gpxFile.MetadataExtensions.GetOrCreateNode(ns, "tracktitle").Data = getTitleFromGpxContent(gpxFile, utils.GetBasename(fileName))
	gpxFile.MetadataExtensions.GetOrCreateNode(ns, "posttitle").Data = post.Title
	gpxFile.MetadataExtensions.GetOrCreateNode(ns, "posturl").Data = post.Url
	gpxFile.MetadataExtensions.GetOrCreateNode(ns, "sourcetype").Data = "wanderer"
	gpxFile.MetadataExtensions.GetOrCreateNode(ns, "sourceurl").Data = sourceUrl

	// write file to the file system
	xmlBytes, err := gpxFile.ToXml(gpx.ToXmlParams{Version: "1.1", Indent: true})
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("error creating the file: %s", err)
	}
	defer file.Close()

	file.Write(xmlBytes)

	return nil
}
