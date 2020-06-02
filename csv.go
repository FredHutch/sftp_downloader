package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// remove PHI columns from filename, in-place
// Note: it appears this function is not called, except from a unit test.
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

type fileSegments struct {
	labType    string
	studyName  string
	suffix     string
	experiment string
}

func getFileSegments(input string) fileSegments {
	var fs fileSegments
	fs.labType = getLabType(input)
	ltStart := strings.LastIndex(input, fs.labType)
	expStart := ltStart + len(fs.labType) + 1
	fs.experiment = input[expStart: /*len(input)*/]
	fs.experiment = strings.Replace(fs.experiment, ".csv", "", 1)
	//anglolab_SABES_2a_anglolab_amilasa.csv

	snStart := len(fs.labType) + 1
	// fmt.Println(fs)
	// fmt.Println("input is", input, ", snStart is", snStart, ", ltStart is", ltStart)
	tmp := input[snStart:ltStart]
	segs := strings.Split(tmp, "_")
	fs.studyName = segs[0]
	//fs.suffix = segs[1]
	fs.suffix = strings.Join(segs[1:len(segs)-1], "_")

	if fs.studyName == "SABES" {
		fs.studyName += "_" + string(fs.suffix[0])
	}

	return fs
}

func getLabType(s string) string {
	longStudies := []string{"biologia_molecular", "DISA_II_c_s_chorrillos_i", "DISA_v", "BSL_III", "NAMRU_6"}
	for _, longStudy := range longStudies {
		if strings.HasPrefix(s, longStudy) {
			return longStudy
		}
	}
	segs := strings.Split(s, "_")
	return segs[0]
}

func getStudyName(s string) string {
	labType := getLabType(s)

	snStart := len(labType) + 1
	var out string
	for i := snStart; i < len(s); i++ {
		if string(s[i]) == "_" {
			break
		}
		out += string(s[i])
	}
	return out
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
func combineCsvs(outDir string, newNameInfo key, csvs ...string) (newCsvName string, err error) {
	newName := filepath.Join(outDir, fmt.Sprintf("%s_%s_%s.csv", newNameInfo.study,
		newNameInfo.labType, newNameInfo.experiment))

	newFile, err := os.Create(newName)
	if err != nil {
		fmt.Println("Could not create combined csv", newName)
		return "", err
	}

	w := csv.NewWriter(newFile)

	for idx, csvFileName := range csvs {
		csvfile, err := os.Open(filepath.Join(outDir, csvFileName))
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
					if err = w.Write(record); err != nil {
						return "", err
					}
					continue

				}
			} else {
				if err = w.Write(record); err != nil {
					return "", err
				}
			}

		}

		w.Flush()
		csvfile.Close()
		err = os.Remove(filepath.Join(outDir, csvFileName))
		if err != nil {
			return "", err
		}
	}

	if err0 := newFile.Close(); err0 != nil {
		return "", err0
	}

	return newName, nil
}

type key struct {
	experiment string
	labType    string
	study      string
}

func groupFilesForCombining(inputDir string) (mm map[key][]string, err error) {
	fileInfos, err := ioutil.ReadDir(inputDir)
	if err != nil {
		return nil, err
	}
	var fileNames []string
	for _, fileInfo := range fileInfos {
		if strings.HasSuffix(fileInfo.Name(), ".csv") {
			fileNames = append(fileNames, fileInfo.Name())
		}
	}
	var m map[key][]string
	m = make(map[key][]string)
	for _, fileName := range fileNames {
		fs := getFileSegments(fileName)
		study := fs.studyName
		kee := key{fs.experiment, fs.labType, study}
		m[kee] = append(m[kee], fileName)
	}
	return m, nil
}
