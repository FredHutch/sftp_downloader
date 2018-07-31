package main

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/otiai10/copy"
)

func TestDeleteOlderFiles(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "sftpdownloader-testing")
	if err != nil {
		t.Fail()
	}
	defer os.RemoveAll(tempDir)

	copy.Copy("testdata/example_tnt_data/2018-07-27/REPORTE-TNTSTUDIES", tempDir)

	err = deleteOlderFiles(tempDir)
	if err != nil {
		t.Error("Expected success but got", err.Error())
	}
	files, err := ioutil.ReadDir(tempDir)
	if err != nil {
		t.Fail()
	}

	if len(files) != 2 {
		t.Error("Expected 2 files to be left, got", len(files))
	}

	var names []string
	for _, file := range files {
		names = append(names, file.Name())
	}

	expected := []string{"TNTstudies-Enrolamiento-20180727223543.csv", "TNTstudies-VisitSummary-20180727224002.csv"}

	if !(stringInSlice(expected[0], names) && stringInSlice(expected[1], names)) {
		t.Error("Expected", expected, "got", names)
	}
}
