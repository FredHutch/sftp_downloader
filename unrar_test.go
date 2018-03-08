package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestOpen(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "sftpdownloader-testing")
	if err != nil {
		t.Fail()
	}
	defer os.RemoveAll(tempDir)

	t.Run("badpassword", func(t *testing.T) {
		err := Open(filepath.Join("testdata", "test.rar"), tempDir, "boguspassword")
		if err.Error() != "rardecode: incorrect password" {
			t.Errorf("expected 'rardecode: incorrect password', got '%s'", err.Error())
		}
	})

	t.Run("goodpassword", func(t *testing.T) {
		err := Open(filepath.Join("testdata", "test.rar"), tempDir, "password")
		if err != nil {
			t.Errorf("expected no error, got '%s'", err.Error())
		}
	})

	t.Run("badfile", func(t *testing.T) {
		err := Open("afilethatdoesnotexist", tempDir, "password")
		exp := "afilethatdoesnotexist: failed to open file: open afilethatdoesnotexist: no such file or directory"
		if err.Error() != exp {
			t.Errorf("expected \n'%s', got \n'%s'", exp, err.Error())
		}
	})

}

func getTempDir() string {
	tempDir, err := ioutil.TempDir("", "sftpdownloader-testing")
	if err != nil {
		panic("can't create temp dir")
	}
	return tempDir
}

func TestUncompressFile(t *testing.T) {
	fileDate := "13-01-2017"
	var tempDir string
	var destFolder string
	var config Config

	setUp := func() {
		tempDir = getTempDir()
		config = Config{LocalDownloadFolderClinical: tempDir}
		destFolder = filepath.Join(config.LocalDownloadFolderClinical, fileDate)
	}

	tearDown := func() {
		os.RemoveAll(tempDir)
	}

	setUp()

	t.Run("badpassword", func(t *testing.T) {
		config.RarDecryptionPassword = "badpassword"
		err := UncompressFile(filepath.Join("testdata", "test.rar"), fileDate, config, ClinicalPhase)
		if err == nil {
			t.Error("Expected error, got nil")
		}
		if err.Error() != "rardecode: incorrect password" {
			t.Error("unexpected error:", err.Error())
		}
	})

	tearDown()
	setUp()

	t.Run("goodpassword", func(t *testing.T) {
		config.RarDecryptionPassword = "password"
		err := UncompressFile(filepath.Join("testdata", "test.rar"), fileDate, config, ClinicalPhase)
		if err != nil {
			t.Error("unexpected error:", err.Error())
		}

	})

	tearDown()
	setUp()

	t.Run("changeme", func(t *testing.T) {
		config.RarDecryptionPassword = "password"
		err := UncompressFile(filepath.Join("testdata", "test.rar"), fileDate, config, ClinicalPhase)
		if err != nil {
			t.Error("unexpected error:", err.Error())
		}
		ok, err := AllFilesExist(filepath.Join(destFolder, "dummyrar"),
			filepath.Join(destFolder, "dummyrar", "1"),
			filepath.Join(destFolder, "dummyrar", "2"),
		)
		if err != nil {
			t.Fail()
		}
		if !ok {
			t.Error("not all files in rar extracted!")
		}
	})

	tearDown()

}
