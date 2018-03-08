package main

import (
	"encoding/csv"
	"io/ioutil"
	"os"
)

// remove PHI columns from filename, in-place
func removePHI(filename string) error {
	file, err := os.Open(filename)
	defer file.Close()
	if err != nil {
		return err
	}

	r := csv.NewReader(file)
	records, err := r.ReadAll()
	if err != nil {
		return err
	}

	header := records[0]
	var columnsToScrub []int

	for i, columnName := range header {
		if columnName == "Iniciales" || columnName == "FechaNacimiento" {
			columnsToScrub = append(columnsToScrub, i)
		}
	}

	tempDir, err := ioutil.TempDir("", "sftp_downloader")
	if err != nil {
		return err
	}

	defer os.RemoveAll(tempDir)

	outfile, err := ioutil.TempFile(tempDir, "temp.csv")
	if err != nil {
		return err
	}

	w := csv.NewWriter(outfile)
	for _, record := range records {
		var tempRecord []string
		for i, cell := range record {
			if !in(i, columnsToScrub) {
				tempRecord = append(tempRecord, cell)
			}
		}
		if err := w.Write(tempRecord); err != nil {
			return err
		}
		w.Flush()
	}

	outName := outfile.Name()

	outfile.Close()

	removeErr := os.Remove(filename)
	if removeErr != nil {
		return removeErr
	}
	renameErr := os.Rename(outName, filename)
	if renameErr != nil {
		return renameErr
	}

	return nil
}

func in(item int, slice []int) bool {
	for _, curr := range slice {
		if curr == item {
			return true
		}
	}
	return false
}
