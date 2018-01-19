package main

import (
	"os"
	"testing"
	"time"

	"github.com/FredHutch/sftp_downloader/mocks"
	"github.com/golang/mock/gomock"
)

func getTheTime() time.Time {
	ret, _ := time.Parse(time.RFC822, time.RFC822) // 2006-01-02 15:04:00 +0000 MST
	return ret
}

// shite = getTheTime
// currentTimeFunction = getTheTime

func TestDoTheWork(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockSftper := mocks.NewMockSftper(mockCtrl)
	mockFiler := mocks.NewMockFiler(mockCtrl)
	mockWalker := mocks.NewMockWalker(mockCtrl)
	mockFileInfo := mocks.NewMockFileInfo(mockCtrl)

	mockSftper.EXPECT().Close().Return(nil).Times(1)
	mockSftper.EXPECT().Walk("/tmp/").Return(mockWalker).Times(1)
	mockSftper.EXPECT().Create("hello.txt").Return(mockFiler, nil).Times(1)
	mockSftper.EXPECT().Lstat("hello.txt").Return(mockFileInfo, nil).Times(1)
	mockWalker.EXPECT().Step().Return(true).Times(1)
	mockWalker.EXPECT().Step().Return(false).Times(1)
	mockWalker.EXPECT().Err().Return(nil).Times(1)
	mockWalker.EXPECT().Path().Return("/tmp/something").Times(1)
	mockFiler.EXPECT().Write([]byte("Hello world!")).Return(12, nil).Times(1)

	doTheWork(mockSftper)

}

func exitcode() string {
	ret, ok := os.LookupEnv("SFTP_DOWNLOADER_EXIT_CODE")
	if !ok {
		return "<exit() not called>"
	}
	return ret
}

func TestGetDateString(t *testing.T) {
	oldArgs := os.Args
	os.Setenv("TESTING_SFTP_DOWNLOADER", "true")
	defer func() { os.Unsetenv("TESTING_SFTP_DOWNLOADER") }()
	defer func() { os.Args = oldArgs }()
	defer func() { os.Unsetenv("SFTP_DOWNLOADER_EXIT_CODE") }()

	t.Run("garbage", func(t *testing.T) {
		os.Args = []string{"sftp_uploader", "config.json", "garbage"}
		getDateString()
		if exitcode() != "1" {
			t.Error("Expected exit with code 1, got", exitcode(), ".")
		}
	})

	t.Run("yesterday", func(t *testing.T) {
		os.Args = []string{"sftp_uploader", "config.json"}
		currentTimeFunction = getTheTime
		ds := getDateString()
		if ds != "01-01-2006" {
			t.Error("Expected 01-01-2006, got", ds)
		}
	})

}
