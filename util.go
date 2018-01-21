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

// IsDir tells you if fileName is a directory
func IsDir(fileName string) (bool, error) {
	stat, err := os.Stat(fileName)
	if err != nil {
		return false, fmt.Errorf("Could not stat '%s'", fileName)
	}
	return stat.IsDir(), nil
}
