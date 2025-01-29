package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func ConvertToSafeFilename(input string) string {
	ext := filepath.Ext(input)

	name := strings.TrimSuffix(input, ext)

	// Replace any character that isn't a letter, number, underscore, or hyphen with an underscore
	re := regexp.MustCompile(`[^a-zA-Z0-9_]+`)
	safeString := re.ReplaceAllString(name, "_") + ext

	return safeString
}

func EnsureDirExists(path string) error {
	// Check if the directory exists
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		// Directory does not exist, create it
		err = os.MkdirAll(path, 0755)
		if err != nil {
			return fmt.Errorf("failed to create directory: %v", err)
		}
		fmt.Println("Directory created:", path)
	} else if err != nil {
		// Some other error occurred
		return fmt.Errorf("failed to check directory: %v", err)
	} else {
		fmt.Println("Directory already exists:", path)
	}

	return nil
}

func SetExtension(path string, newExt string) string {

	ext := filepath.Ext(path)

	// if path is not terminated by extension
	if ext == "" {
		return path + "." + newExt
	}

	name := strings.TrimSuffix(path, ext)
	return name + "." + newExt
}

func FileExists(path string) bool {

	_, err := os.Stat(path)
	if err != nil {
		// If error is not nil, it means file doesn't exist
		if os.IsNotExist(err) {
			return false
		} else {
			panic(err)
		}
	}
	return true
}

func RemoveExtension(input string) string {
	ext := filepath.Ext(input)
	return strings.TrimSuffix(input, ext)
}

func GetBasename(path string) string {
	return filepath.Base(path)
}
