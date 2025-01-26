package tracks

import (
	"encoding/json"
	"mnezerka/geonet/utils"
	"os"
)

func ReadTrackMeta(filePath string) (TrackMeta, error) {

	result := TrackMeta{}

	filePathMeta := utils.SetExtension(filePath, "json")

	jsonFile, err := os.Open(filePathMeta)
	if err != nil {
		return result, err
	}
	defer jsonFile.Close()

	err = json.NewDecoder(jsonFile).Decode(&result)

	return result, err
}
