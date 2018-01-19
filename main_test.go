package main

import (
	"testing"

	// "github.com/FredHutch/sftp_downloader/main"
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

/*

package main

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/kr/fs"
	"github.com/pkg/sftp"
)

type FakeSftper struct {
	thefile *os.File
	sfile   *sftp.File
	fi      os.FileInfo
	walker  *fs.Walker
}

// type FakeWalker struct {
//
// }

func (f FakeSftper) Create(path string) (*sftp.File, error) {
	f.thefile, _ = ioutil.TempFile("/tmp", "fakesftper")
	// f.sfile.path = f.thefile.Name()
	return f.sfile, nil
}

func (f FakeSftper) Close() error {
	f.thefile.Close()
	return nil
}

func (f FakeSftper) Lstat(p string) (os.FileInfo, error) {
	return f.fi, nil
}

func (f FakeSftper) Walk(root string) *fs.Walker {
	return f.walker
}

func TestDoTheWork(t *testing.T) {
	f := FakeSftper{}
	doTheWork(f)
}
*/
