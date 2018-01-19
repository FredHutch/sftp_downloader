package main

import (
	"errors"
	"testing"

	// "github.com/FredHutch/sftp_downloader/main"
	"github.com/FredHutch/sftp_downloader/mocks"
	"github.com/golang/mock/gomock"
)

type TestHarness struct {
	Sftper Sftper
}

func TestDoTheWork(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	// testWorker := &main
	mockSftper := mocks.NewMockSftper(mockCtrl)
	// testHarness := &TestHarness{Sftper: mockSftper}
	dummyError := errors.New("dummy error")
	mockSftper.EXPECT().Close().Return(dummyError).Times(1)
	// mockSftper.EXPECT().Create("hello.txt").return()

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
