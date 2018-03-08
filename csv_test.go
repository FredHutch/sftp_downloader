package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/udhos/equalfile"
)

func TestRemovePHI(t *testing.T) {
	tempdir, err := ioutil.TempDir("", "sftp_downloader_test")
	if err != nil {
		t.Fail()
	}
	defer os.RemoveAll(tempdir)
	testfile := fmt.Sprintf("%s/removephi.csv", tempdir)
	Copy("testdata/removephi.csv", testfile)
	rerr := removePHI(testfile)
	if rerr != nil {
		t.Error("did not expect error")
	}

	options := equalfile.Options{}
	cmp := equalfile.New(nil, options)
	result, err := cmp.CompareFile(testfile, "testdata/expected_shrunk.csv")
	if err != nil {
		t.Fail()
	}
	if !result {
		t.Error("files do not match")
	}

}

func Copy(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}
