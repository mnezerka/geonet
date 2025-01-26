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

func WandererStoreTrack(post *WandererPost, track *WandererTrack, dir string, force bool) error {

	fileName := utils.ConvertToSafeFilename(track.Url)
	filePath := filepath.Join(dir, fileName)

	fileNameMeta := utils.SetExtension(fileName, "json")
	filePathMeta := filepath.Join(dir, fileNameMeta)

	log.Infof("  track file name      : %s", filePath)
	log.Infof("  track metad file name: %s", filePathMeta)

	// Fetch the file from the URL
	resp, err := http.Get(track.Url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// check if file exists in not-forced mode
	if !force {
		if utils.FileExists(filePath) {
			log.Info("  skipping, track file exists")
			return nil
		}
	}

	// -------------------------------------------------------------
	// 1. Create the track content file

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("Error creating the file: %s", err)
	}
	defer file.Close()

	// Copy the content from the response body to the local file
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("Error saving the file: %s", err)
	}

	// -------------------------------------------------------------
	// 2. Create the track meta file
	file, err = os.Create(filePathMeta)
	if err != nil {
		return fmt.Errorf("Error creating the meta file: %s", err)
	}
	defer file.Close()

	meta := TrackMeta{
		SourceId: fileName,
		Title:    post.Title + " (" + fileName + ")",
		Url:      post.Url,
	}

	jsonData, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}

	_, err = file.Write(jsonData)
	if err != nil {
		return err
	}

	return nil
}
