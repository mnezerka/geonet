package cmd

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

	"github.com/spf13/cobra"
)

var fetchWandererCmdDir string
var fetchWandererCmdForce bool

type wandererTrack struct {
	Url string
}

type wandererPost struct {
	Title  string
	Tracks []wandererTrack
	Url    string
}

var fetchWandererCmd = &cobra.Command{
	Use:   "wanderer url",
	Short: "Fetch tracks from site powered by wanderer hugo template",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		// guess if first arg is file path or url
		urlCheck, err := url.Parse(args[0])
		if err != nil {
			return err
		}

		var m []wandererPost

		if urlCheck.Scheme == "" {
			m, err = readMediaIndexFromFileSystem(args[0])
		} else {
			m, err = readMediaIndexFromUrl(args[0])

		}
		if err != nil {
			return err
		}

		// ensure directory to be used for fetched tracks exists
		err = utils.EnsureDirExists(fetchWandererCmdDir)
		if err != nil {
			return err
		}

		log.Infof("%d wanderer posts will be processed", len(m))

		for i := 0; i < len(m); i++ {
			log.Infof("post %d '%s' (%d tracks)", i, m[i].Title, len(m[i].Tracks))
			for j := 0; j < len(m[i].Tracks); j++ {
				storeTrack(&m[i], &m[i].Tracks[j])
			}

			break

		}

		return nil
	},
}

func init() {
	fetchWandererCmd.PersistentFlags().StringVar(&fetchWandererCmdDir, "dir", "tracks", "directory for storing fetched tracks")
	fetchWandererCmd.PersistentFlags().BoolVarP(&fetchWandererCmdForce, "force", "f", false, "download all tracks, no matter if already fetched")

	fetchCmd.AddCommand(fetchWandererCmd)
}

func storeTrack(post *wandererPost, track *wandererTrack) error {

	fileName := utils.ConvertToSafeFilename(track.Url)
	filePath := filepath.Join(fetchWandererCmdDir, fileName)

	fileNameMeta := utils.SetExtension(fileName, "json")
	filePathMeta := filepath.Join(fetchWandererCmdDir, fileNameMeta)

	log.Infof("  track file name      : %s", filePath)
	log.Infof("  track metad file name: %s", filePathMeta)

	// Fetch the file from the URL
	resp, err := http.Get(track.Url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// check if file exists in not-forced mode
	if !fetchWandererCmdForce {
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
		Title: post.Title + " (" + fileName + ")",
		Url:   post.Url,
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

func readMediaIndexFromUrl(wandererUrl string) ([]wandererPost, error) {

	result := []wandererPost{}

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

func readMediaIndexFromFileSystem(filePath string) ([]wandererPost, error) {

	result := []wandererPost{}

	log.Infof("fetching media index from local file '%s'", filePath)

	jsonFile, err := os.Open(filePath)
	if err != nil {
		return result, err
	}
	defer jsonFile.Close()

	err = json.NewDecoder(jsonFile).Decode(&result)

	return result, nil
}
