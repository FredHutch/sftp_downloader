package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// FileExists tells you if fileName exists
func FileExists(fileName string) (bool, error) {
	_, err := os.Stat(fileName)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("Other error Stat()ing local download directory: %s",
			err.Error())
	}
	return true, nil
}

// AllFilesExist tells you if all files in a list exist
func AllFilesExist(fileNames ...string) (bool, error) {
	for _, fileName := range fileNames {
		ok, err := FileExists(fileName)
		if err != nil {
			return false, err
		}
		if !ok {
			return false, nil
		}
	}
	return true, nil
}

// IsDir tells you if fileName is a directory
func IsDir(fileName string) (bool, error) {
	stat, err := os.Stat(fileName)
	if err != nil {
		return false, fmt.Errorf("Could not stat '%s'", fileName)
	}
	return stat.IsDir(), nil
}

func renameDownloadDir(config Config, fileDate string, phase Phase) (string, error) {
	downloadFolder := getDownloadFolder(phase, config)
	t, err := time.Parse("02-01-2006", fileDate)
	if err != nil {
		return "", fmt.Errorf("Could not convert %s to Time object: %s", fileDate, err.Error())
	}
	newName := t.Format("2006-01-02")
	err = os.Rename(filepath.Join(downloadFolder, fileDate),
		filepath.Join(downloadFolder, newName))
	if err != nil {
		return "", fmt.Errorf("Could not rename download directory: %s", err.Error())
	}
	return newName, nil

}
