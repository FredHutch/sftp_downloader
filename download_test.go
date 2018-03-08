package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/FredHutch/sftp_downloader/mocks"
	"github.com/golang/mock/gomock"
	"github.com/udhos/equalfile"
)

func areFilesEqual(file1 string, file2 string) (bool, error) {
	options := equalfile.Options{}
	cmp := equalfile.New(nil, options)
	return cmp.CompareFile(file1, file2)
}

type FakeFileInfo struct {
	name string
}

func (f FakeFileInfo) Name() string {
	return f.name
}

func (f FakeFileInfo) Size() int64 {
	return 0
}

func (f FakeFileInfo) Mode() os.FileMode {
	return 0
}

func (f FakeFileInfo) ModTime() time.Time {
	return time.Now()
}

func (f FakeFileInfo) IsDir() bool {
	return false
}

func (f FakeFileInfo) Sys() interface{} {
	return nil
}

func bollocks() os.FileInfo {
	f := FakeFileInfo{name: "Reportes_diarios_acumulados-2018-01-01"}
	fmt.Println(f.Name())
	return f
}

// verify that FakeFileInfo implements os.FileInfo, see
// https://golang.org/doc/faq#guarantee_satisfies_interface
var _ os.FileInfo = FakeFileInfo{}
var _ os.FileInfo = (*FakeFileInfo)(nil)

func TestGetFileNameToDownload(t *testing.T) {
	// see https://github.com/golang/mock/issues/51#issuecomment-324427140
	// for why we have to put code in subtests that should be in outer test func

	t.Run("basic", func(t *testing.T) {
		tempDir, err := ioutil.TempDir("", "sftp-test-dir")
		if err != nil {
			t.Fail()
		}

		config := Config{LocalDownloadFolderClinical: tempDir}
		defer os.RemoveAll(tempDir)

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		mockSftper := mocks.NewMockSftper(mockCtrl)
		f := FakeFileInfo{name: "Reportes_diarios_acumulados-2018-01-01"}

		var fs []os.FileInfo
		fs = append(fs, f)

		mockSftper.EXPECT().ReadDir("/").Return(fs, nil).Times(1)

		res, err := getFileNameToDownload("2018-01-01", config, mockSftper, ClinicalPhase)
		if err != nil {
			t.Error("Expected success, got error")
		}
		if res != "/Reportes_diarios_acumulados-2018-01-01" {
			t.Error("Expected '/Reportes_diarios_acumulados-2018-01-01', got", res)
		}
		if err == nil {
			t.Log("res is", res)
		}
	})

	t.Run("changeme", func(t *testing.T) {
	})

}

func TestDoDownload(t *testing.T) {

	tempDir, err := ioutil.TempDir("", "sftp-test-dir")
	if err != nil {
		t.Fail()
	}
	defer os.RemoveAll(tempDir)

	t.Run("changeme", func(t *testing.T) {
		config := Config{LocalDownloadFolderClinical: tempDir}

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		mockSftper := mocks.NewMockSftper(mockCtrl)

		fh, err := os.Open(filepath.Join("testdata", "test.rar"))
		if err != nil {
			t.Fail()
		}
		mockSftper.EXPECT().Open(gomock.Any()).Return(fh, nil).Times(1)

		res, err := doDownload("remoteFile", config, mockSftper, ClinicalPhase)
		if err != nil {
			t.Error("did not expect error")
		}

		ok, err := areFilesEqual(res, filepath.Join("testdata", "test.rar"))
		if err != nil {
			t.Fail()
		}

		if !ok {
			t.Error("copied file not identical to text fixture!")
		}

	})

}
