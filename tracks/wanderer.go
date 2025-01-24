package tracks

import (
	"encoding/json"
	"mnezerka/geonet/log"
	"net/http"
	"net/url"
	"os"
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

	return result, nil
}
