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
			nameComponents         key
			expectedOutputFileName string
			expectedNumberOfLines  int
		}{
			{
				[]string{"anglolab_MERLIN_a_anglolab_gamma_glutamil_transpeptidasa.csv", "anglolab_MERLIN_anglolab_gamma_glutamil_transpeptidasa.csv"},
				key{"gamma_glutamil_transpeptidasa", "anglolab", "MERLIN"},
				"MERLIN_anglolab_gamma_glutamil_transpeptidasa.csv",
				3,
			},
			{
				[]string{"anglolab_MERLIN_PL_a_anglolab_estudio_bioquimico_de_LCR.csv", "anglolab_MERLIN_PL_b_anglolab_estudio_bioquimico_de_LCR.csv"},
				key{"estudio_bioquimico_de_LCR", "anglolab", "MERLIN"},
				"MERLIN_anglolab_estudio_bioquimico_de_LCR.csv",
				3,
			},
		}

		copy.Copy("testdata/semi-processed-lab-files", tempDir)

		for _, tester := range testers {
			tester.expectedOutputFileName = filepath.Join(tempDir, tester.expectedOutputFileName)
			newName, err := combineCsvs(tempDir, tester.nameComponents, tester.inputFiles...)

			if err != nil {
				t.Error("got an unexpected error: ", err.Error())
			}
			if newName != tester.expectedOutputFileName {
				t.Error("expected", tester.expectedOutputFileName, "got", newName)
			}

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
				fullPath := filepath.Join(tempDir, inputFile)
				exists, err := FileExists(fullPath)
				if err != nil {
					t.Fail()
				}
				if exists {
					t.Error(fullPath, "should not exist!")
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

func TestGetLabType(t *testing.T) {
	actual := getLabType("foo_bar")
	expected := "foo"
	if actual != expected {
		t.Error("got", actual, "expected", expected)
	}
	actual = getLabType("biologia_molecular_foo")
	expected = "biologia_molecular"
	if actual != expected {
		t.Error("got", actual, "expected", expected)
	}

}

func TestGetFileSegments(t *testing.T) {
	res := getFileSegments("anglolab_SABES_2a_anglolab_amilasa.csv")
	if res.labType != "anglolab" {
		t.Error("expected anglolab, got", res.labType)
	}
	if res.experiment != "amilasa" {
		t.Error("expected amilasa, got", res.experiment)
	}
	var segtesters = []struct {
		input    string
		segments fileSegments
	}{
		{
			"anglolab_SABES_2a_anglolab_amilasa.csv",
			fileSegments{"anglolab", "SABES_2", "2a", "amilasa"},
		},
		{
			"anglolab_SABES_3b_TD_anglolab_TSH_ultrasensible.csv",
			fileSegments{"anglolab", "SABES_3", "3b_TD", "TSH_ultrasensible"},
		},
		{
			"DISA_II_c_s_chorrillos_i_SABES_3b_TD_DISA_II_c_s_chorrillos_i_investigacion_bacteriologica_en_tuberculosis.csv",
			fileSegments{"DISA_II_c_s_chorrillos_i", "SABES_3", "3b_TD", "investigacion_bacteriologica_en_tuberculosis"},
		},
		{
			"bioquimica_MERLIN_b_CU_bioquimica_transaminasa_piruvica_TGP_ALT.csv",
			fileSegments{"bioquimica", "MERLIN", "b_CU", "transaminasa_piruvica_TGP_ALT"},
		},
		{
			"BSL_III_SABES_1_BSL_III_carga_viral_HIV1_xpert.csv",
			fileSegments{"BSL_III", "SABES_1", "1", "carga_viral_HIV1_xpert"},
		},
		{
			"BSL_III_SABES_2a_BSL_III_carga_viral_HIV1_xpert.csv",
			fileSegments{"BSL_III", "SABES_2", "2a", "carga_viral_HIV1_xpert"},
		},
	}
	for _, tester := range segtesters {
		actual := getFileSegments(tester.input)
		if actual != tester.segments {
			t.Errorf("Expected\n %+v, got \n%+v\n", tester.segments, actual)
		}
	}
}

func TestGroupFilesForCombining(t *testing.T) {
	groupFilesForCombining("testdata/semi-processed-lab-files")
}
