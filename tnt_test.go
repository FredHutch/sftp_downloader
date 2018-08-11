package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kniren/gota/dataframe"

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

func TestConvertTNTDates(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "sftpdownloader-testing")
	if err != nil {
		t.Fail()
	}
	defer os.RemoveAll(tempDir)

	copy.Copy("testdata/example_tnt_data/2018-07-27/REPORTE-TNTSTUDIES", tempDir)
	err = deleteOlderFiles(tempDir)
	if err != nil {
		t.Fail()
	}
	err = renameFiles(tempDir)
	if err != nil {
		t.Fail()
	}
	err = convertTNTDates(tempDir)
	if err != nil {
		t.Error("did not expect error, got", err.Error())
	}
	files, err := ioutil.ReadDir(tempDir)
	if err != nil {
		t.Fail()
	}

	var colname string
	var expected string
	expectedCol := "PTID"
	for _, finfo := range files {
		if strings.Index(finfo.Name(), "enr") > -1 {
			colname = "FechaEnr"
			expected = "08/29/2013"
		} else {
			colname = "FechaCita"
			expected = "07/26/2018"
		}
		f, err := os.Open(filepath.Join(tempDir, finfo.Name()))
		if err != nil {
			t.Fail()
		}
		df := dataframe.ReadCSV(f, dataframe.HasHeader(true), dataframe.DetectTypes(false))
		f.Close()
		sub := df.Subset([]int{0})
		actual := sub.Col(colname).Records()[0]
		if actual != expected {
			t.Error("expected", expected, "got", actual)
		}
		if !stringInSlice(expectedCol, df.Names()) {
			t.Error("expected column", expectedCol, "in columns", df.Names())
		}
	}
}

func TestMoveFilesUpOneLevel(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "sftpdownloader-testing")
	if err != nil {
		fmt.Println("fail0", err.Error())
		t.Fail()
	}
	defer os.RemoveAll(tempDir)
	f, err := os.Create(filepath.Join(tempDir, "1.txt"))
	if err != nil {
		fmt.Println("fail1", err.Error())
		t.Fail()
	}
	f.Close()
	fulldir := filepath.Join(tempDir, "REPORTE-TNTSTUDIES")
	err = mkdir(fulldir)
	if err != nil {
		fmt.Println("fail2", err.Error())
		t.Fail()
	}
	f, err = os.Create(filepath.Join(fulldir, "f1.txt"))
	if err != nil {
		fmt.Println("fail3", err.Error())
		t.Fail()
	}
	f.Close()
	err = moveFilesUpOneLevel(tempDir)
	if err != nil {
		t.Error("did not expect error, got", err.Error())
	}
	infos, err := ioutil.ReadDir(tempDir)
	if err != nil {
		fmt.Println("got an error reading tempDir")
		t.Fail()
	}
	if len(infos) != 2 {
		t.Error("Expected len(infos) to be 2, got", len(infos))
	}
}
