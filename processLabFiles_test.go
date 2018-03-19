package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/kniren/gota/dataframe"
)

func TestGetKey(t *testing.T) {
	t.Run("test1", func(t *testing.T) {
		input := "testdata/example_lab_data/REPORTE-CSV-BD-LAB/MERLIN/Biología Molecular/Carga Viral HIV1 Abbott RT-09032018223003.csv"
		actual := getKey(input)
		expected := "Biología Molecular/Carga Viral HIV1 Abbott RT"
		if actual != expected {
			t.Errorf("expected %s, got %s", expected, actual)
		}
	})
}

func TestProcessLabFiles(t *testing.T) {
	t.Run("test1", func(t *testing.T) {
		tempDir, err := ioutil.TempDir("", "sftp-test-dir")
		if err != nil {
			t.Fail()
		}
		config := Config{LocalDownloadFolderLab: tempDir, PhiZipPassword: "foobar"}
		defer os.RemoveAll(tempDir)
		err = processLabFiles(config, filepath.Join("testdata", "example_lab_data"))
		if err != nil {
			t.Error(err.Error())
		}

	})
}

func TestPtidExists(t *testing.T) {
	// var df dataframe.DataFrame

	f, err := os.Open("testdata/example_lab_data/REPORTE-CSV-BD-LAB/MERLIN/Anglolab/Gamma Glutamil Transpeptidasa-09032018223003.csv")
	if err != nil {
		t.Fail()
	}
	defer f.Close()
	newDf := dataframe.ReadCSV(f,
		dataframe.HasHeader(true),
		dataframe.DetectTypes(false))
	// fmt.Println(newDf)

	t.Run("testgood", func(t *testing.T) {
		actual := ptidExists("09219-0044-2", newDf)
		if !actual {
			t.Errorf("got false, expected true")
		}
	})

	t.Run("testbad", func(t *testing.T) {
		actual := ptidExists("badvalue", newDf)
		if actual {
			t.Errorf("got true, expected false")
		}
	})
}

var keyToFileNameTable = []struct {
	in  string
	out string
}{
	{"Bioquímica/Transaminasa Piruvica(TGP)-ALT", "bioquimica_transaminasa_piruvica_TGP_ALT.csv"},
	{"Blufstein/Estudio directo de BK en Esputo", "blufstein_estudio_directo_de_BK_en_esputo.csv"},
	{"Urianálisis/OneStep EtG test Dip Card", "urianalisis_onestep_etg_test_dip_card.csv"},
}

func TestKeyToFileName(t *testing.T) {
	for _, tt := range keyToFileNameTable {
		actual, err := keyToFileName(tt.in)
		if err != nil {
			t.Error()
		}
		if actual != tt.out {
			t.Errorf("Expected %s, got %s", tt.out, actual)
		}
	}
}
