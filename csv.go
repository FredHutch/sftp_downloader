package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
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

// FIXME TODO add a function that will list all the csv files (in semi-processed)
// and return a list of groups of files that can be combined, to be fed into combineCsvs().
// It should also return 'orphans' that don't get combined as these still need
// to be renamed so the lab type does not appear twice. Or better yet, fix that issue
// upstream (though that will have impacts here).

/*
anglolab_MERLIN_a_anglolab_gamma_glutamil_transpeptidasa.csv
and
anglolab_MERLIN_anglolab_gamma_glutamil_transpeptidasa.csv
can be combined as
MERLIN_anglolab_gamma_glutamil_transpeptidasa.csv
Likely,
anglolab_MERLIN_PL_a_anglolab_estudio_bioquimico_de_LCR.csv
and
anglolab_MERLIN_PL_b_anglolab_estudio_bioquimico_de_LCR.csv
can be combined as
MERLIN_anglolab_estudio_bioquimico_de_LCR.csv.

Basically, _PL, _a, _b after MERLIN_ can be lumped into one file per lab type.

-          Sabes 2 and 2a can be combined.
-          Sabes 3, 3a_, and 3b_ can be combined.
*/
func combineCsvs(csvs ...string) (newCsvName string, err error) {
	fmt.Println(csvs[0])
	// dir1 := path.Dir(csv1)
	// dir2 := path.Dir(csv2)
	// if dir1 != dir2 {
	// 	fmt.Println("Directories of two CSVs to combine should be the same!")
	// 	os.Exit(1)
	// }
	// base1 := path.Base(csv1)
	// base2 := path.Base(csv2)

	segs := strings.Split(path.Base(csvs[0]), "_")
	// FIXME TODO this logic will need to change as there is one lab type (biologia_molecular) that has an underscore in it.
	// This needs to be treated as one segment.
	labType := segs[0]
	studyName := segs[1]
	// fmt.Println("labType is", labType, "and studyName is", studyName)

	// FIXME TODO this method of generating the output file name probably has to change
	// For example, if combining

	newNameSegs := []string{studyName, labType}
	foundLabType := false
	for i := 2; i < len(segs); i++ {
		if segs[i] == labType {
			foundLabType = true
			continue
		}
		if foundLabType {
			newNameSegs = append(newNameSegs, segs[i])
		} else {
			// don't do nuthin'
		}
	}
	dir := path.Dir(csvs[0])
	newName := filepath.Join(dir, strings.Join(newNameSegs, "_"))

	newFile, err := os.Create(newName)
	if err != nil {
		fmt.Println("Could not create combined csv", newName)
		return "", err
	}

	w := csv.NewWriter(newFile)

	for idx, csvFileName := range csvs {
		csvfile, err := os.Open(csvFileName)
		if err != nil {
			fmt.Println("error opening csv", csvFileName)
			return "", err
		}
		r := csv.NewReader(csvfile)
		first := true

		for {
			record, err := r.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				fmt.Println("error reading from", csvFileName)
				return "", err
			}
			if first {
				first = false
				if idx == 0 {
					fmt.Println("header record is")
					fmt.Println(record)
					if err = w.Write(record); err != nil {
						return "", err
					}
					continue

				}
			} else {
				fmt.Println("in endless for loop, first is", first, "and idx is", idx)
				fmt.Println("writing a data record")
				if err = w.Write(record); err != nil {
					return "", err
				}
			}

		}

		w.Flush()
		csvfile.Close()
		os.Remove(csvFileName)
	}

	if err0 := newFile.Close(); err0 != nil {
		return "", err0
	}

	return newName, nil
}
