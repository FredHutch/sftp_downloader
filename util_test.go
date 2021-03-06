package main

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestFileExists(t *testing.T) {
	tempFh, err := ioutil.TempFile("", "sftp_downloader_tests")
	if err != nil {
		t.Fail()
	}
	fileName := tempFh.Name()
	tempFh.Close()
	defer func() { os.Remove(fileName) }()

	exists, err := FileExists(fileName)
	if err != nil {
		t.Errorf("Got unexpected error checking if file exists: %s", err.Error())
	}
	if !exists {
		t.Error("FileExists() unexpectedly returned false")
	}

	tempDir, err := ioutil.TempDir("", "another-temp-dir")
	if err != nil {
		t.Fail()
	}
	tempFile2fh, err := ioutil.TempFile(tempDir, "afile")
	fileName2 := tempFile2fh.Name()
	tempFile2fh.Close()
	os.Chmod(tempDir, 0000)

	defer func() {
		os.Chmod(tempDir, 0777)
		os.Remove(fileName2)
		os.Remove(tempDir)
	}()

	_, err = FileExists(fileName2)
	if err == nil {
		t.Error("Expected error checking existence of unreadable file")
	}

	exists, err = FileExists("/123_-_-/456/789")
	if err != nil {
		t.Errorf("Got unexpected error checking if file exists: %s", err.Error())
	}
	if exists {
		t.Error("FileExists() unexpectedly returned true")
	}

}

func TestIsDir(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "sftp_downloader_tests")
	if err != nil {
		t.Fail()
	}
	defer func() { os.Remove(tempDir) }()

	tempFh, err := ioutil.TempFile("", "sftp_downloader_tests")
	if err != nil {
		t.Fail()
	}
	fileName := tempFh.Name()

	isdir, err := IsDir(tempDir)
	if err != nil {
		t.Error("Got unexpected error calling IsDir()")
	}
	if !isdir {
		t.Error("Expected true, got false")
	}

	_, err = IsDir("/123_-_-/456/789")
	if err == nil {
		t.Error("Expected error calling IsDir() on nonexistent file")
	}

	isdir, err = IsDir(fileName)
	if err != nil {
		t.Error("Got unexpected error calling IsDir()")
	}

	if isdir {
		t.Error("Expected false, got true")
	}

}

func TestChangeDate(t *testing.T) {
	var datetests = []struct {
		in  string
		out string
	}{
		{"25/10/2002", "10/25/2002"},
		{"08-07-2014", "07/08/2014"},
		{"14.07.16", "07/14/2016"},
		{"19/07/2016 03:48:33 p.m.", "07/19/2016 03:48:33 p.m."},
	}
	for _, datetest := range datetests {
		actual := changeDate(datetest.in)
		if actual != datetest.out {
			t.Error("got", actual, "expected", datetest.out)
		}
	}
}

//func renameDownloadDir(config Config, fileDate string) (string, error)
// func TestRenameDownloadDir(t *testing.T) {
// 	tempDir, err := ioutil.TempDir("", "sftp_downloader_test")
// 	if err != nil {
// 		t.Fail()
// 	}
// 	defer os.RemoveAll(tempDir)
//
// 	t.Run("changeme", func(t *testing.T) {
// 		config := Config{LocalDownloadFolderClinical: tempDir}
// 		oldName := "2018-01-24"
// 		err := os.Mkdir(filepath.Join(tempDir, oldName), os.ModePerm)
// 		if err != nil {
// 			t.Fail()
// 		}
// 		newName, err := renameDownloadDir(config, oldName, ClinicalPhase)
// 		if err != nil {
// 			t.Errorf("Got unexpected error: %s", err.Error())
// 		}
// 		if newName != "2018-01-24" {
// 			t.Errorf("Expected 2018-01-24, got %s", newName)
// 		}
// 		fh, err := os.Open(tempDir)
// 		if err != nil {
// 			t.Fail()
// 		}
// 		infos, err := fh.Readdir(-1)
// 		if err != nil {
// 			t.Fail()
// 		}
// 		if err != nil {
// 			t.Fail()
// 		}
// 		if len(infos) != 1 {
// 			t.Errorf("expected 1 item, got %d", len(infos))
// 		}
// 		if infos[0].Name() != newName {
// 			t.Errorf("expected %s, got %s", newName, infos[0].Name())
// 		}
// 		if !infos[0].IsDir() {
// 			t.Error("expected directory, got file")
// 		}
// 	})
// }
