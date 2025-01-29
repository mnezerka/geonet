package tracks

import (
	"encoding/json"
	"fmt"
	"io"
	"mnezerka/geonet/log"
	"mnezerka/geonet/utils"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

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

func WandererStoreTrackMeta(post *WandererPost, track *WandererTrack, dir string, force bool) error {

	fileName := utils.ConvertToSafeFilename(track.Url)
	filePath := filepath.Join(dir, fileName)

	log.Infof("  track file name      : %s", filePath)

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

	// write file to the file system
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("Error creating the file: %s", err)
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("Error saving the file: %s", err)
	}

	// update metadata
	t := NewTrack(filePath)
	t.Meta.SourceId = fileName
	t.Meta.Url = post.Url
	t.Meta.Creators = append(t.Meta.Creators, "wanderer")

	// try to get title from the track data
	title := t.GetTitleFromContent()

	if len(title) > 0 {
		t.Meta.Title = post.Title + " (" + title + ")"
	} else {
		t.Meta.Title = post.Title + " (" + utils.GetBasename(fileName) + ")"
	}

	t.WriteMeta()

	return nil
}
