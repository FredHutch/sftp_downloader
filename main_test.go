package main

import (
	"testing"

	"github.com/FredHutch/sftp_downloader/mocks"
	"github.com/golang/mock/gomock"
)

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
