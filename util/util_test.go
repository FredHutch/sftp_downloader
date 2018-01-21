package util

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
