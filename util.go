package main

import (
	"fmt"
	"os"
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
