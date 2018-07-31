package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func deleteOlderFiles(path string) error {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return err
	}
	patterns := []string{"Enrolamiento", "VisitSummary"}
	for _, pattern := range patterns {
		var max int
		var maxName string
		for _, file := range files {
			if strings.Index(file.Name(), pattern) > -1 {
				segs := strings.Split(file.Name(), "-")
				segs = strings.Split(segs[2], ".")
				num, err := strconv.Atoi(segs[0])
				if err != nil {
					return err
				}
				if num > max {
					max = num
					maxName = file.Name()
				}
			}
		}

		// now loop through again and delete all files that are not maxName
		for _, file := range files {
			if strings.Index(file.Name(), pattern) > -1 {
				if file.Name() != maxName {
					err = os.Remove(filepath.Join(path, file.Name()))
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}
