package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/FredHutch/sftp_downloader/mocks"
	"github.com/golang/mock/gomock"
)

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

func TestDownloadFile(t *testing.T) {
	// see https://github.com/golang/mock/issues/51#issuecomment-324427140
	// for why we have to put code in subtests that should be in outer test func

	t.Run("basic", func(t *testing.T) {
		tempDir, err := ioutil.TempDir("", "sftp-test-dir")
		if err != nil {
			t.Fail()
		}
		config := Config{LocalDownloadFolder: tempDir}
		// TODO remove tempdir in defer...

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		mockSftper := mocks.NewMockSftper(mockCtrl)
		mockReader := mocks.NewMockReader(mockCtrl)
		f := FakeFileInfo{name: "Reportes_diarios_acumulados-2018-01-01"}

		var fs []os.FileInfo
		fs = append(fs, f)

		// reader, err := os.Open("testdata/test.rar")
		// if err != nil {
		// 	t.Fail()
		// }
		// defer reader.Close()

		mockSftper.EXPECT().ReadDir("/").Return(fs, nil).Times(1)
		mockSftper.EXPECT().Open(gomock.Any()).Return(mockReader, nil).Times(1)
		mockReader.EXPECT().Read(gomock.Any()).Times(1)

		res, err := downloadFile("2018-01-01", config, mockSftper)
		t.Log("after calling downloadfile")
		if err == nil {
			t.Log("res is", res)
			// t.Error("Expected non-nil value")

		}
		// t.Log("athabasca")
		// t.Log("err is", err.Error())
		// if err.Error() != "Found no file matching pattern 'Reportes_diarios_acumulados-2018-01-01'" {
		// 	t.Errorf("Got unexpected error message: %s", err.Error())
		// }
	})

	t.Run("changeme", func(t *testing.T) {
	})

}
