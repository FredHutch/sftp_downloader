package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/otiai10/copy"
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

func TestCombineCsvs(t *testing.T) {
	t.Run("changeme", func(t *testing.T) {

		tempDir, err := ioutil.TempDir("", "sftp_downloader_test")
		if err != nil {
			t.Fail()
		}
		defer os.RemoveAll(tempDir)

		var testers = []struct {
			inputFiles             []string
			expectedOutputFileName string
			expectedNumberOfLines  int
		}{
			{
				[]string{"anglolab_MERLIN_a_anglolab_gamma_glutamil_transpeptidasa.csv", "anglolab_MERLIN_anglolab_gamma_glutamil_transpeptidasa.csv"},
				"MERLIN_anglolab_gamma_glutamil_transpeptidasa.csv",
				3,
			},
			{
				[]string{"anglolab_MERLIN_PL_a_anglolab_estudio_bioquimico_de_LCR.csv", "anglolab_MERLIN_PL_b_anglolab_estudio_bioquimico_de_LCR.csv"},
				"MERLIN_anglolab_estudio_bioquimico_de_LCR.csv",
				3,
			},
		}

		fmt.Println(testers)

		copy.Copy("testdata/semi-processed-lab-files", tempDir)

		// // fmt.Println("tempDir is", tempDir)

		for _, tester := range testers {
			for i := range tester.inputFiles {
				tester.inputFiles[i] = filepath.Join(tempDir, tester.inputFiles[i])
			}
			tester.expectedOutputFileName = filepath.Join(tempDir, tester.expectedOutputFileName)
			newName, err := combineCsvs(tester.inputFiles...)

			if err != nil {
				t.Error("got an unexpected error: ", err.Error())
			}
			if newName != tester.expectedOutputFileName {
				t.Error("expected", tester.expectedOutputFileName, "got", newName)
			}

			fmt.Println("newName is", newName)

			newFile, err := os.Open(newName)
			if err != nil {
				t.Fail()
			}
			defer newFile.Close()
			lines, err := lineCounter(newFile)
			if err != nil {
				t.Fail()
			}
			if lines != tester.expectedNumberOfLines {
				t.Error("expected", tester.expectedNumberOfLines, "lines, got", lines)
			}

			for _, inputFile := range tester.inputFiles {
				exists, err := FileExists(inputFile)
				if err != nil {
					t.Fail()
				}
				if exists {
					t.Error(inputFile, "should not exist!")
				}
			}

		}

	})

}

func lineCounter(r io.Reader) (int, error) {
	buf := make([]byte, 32*1024)
	count := 0
	lineSep := []byte{'\n'}

	for {
		c, err := r.Read(buf)
		count += bytes.Count(buf[:c], lineSep)

		switch {
		case err == io.EOF:
			return count, nil

		case err != nil:
			return count, err
		}
	}
}
