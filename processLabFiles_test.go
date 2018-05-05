package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kniren/gota/dataframe"
	"github.com/kniren/gota/series"
	"github.com/otiai10/copy"
)

func TestGetKey(t *testing.T) {
	t.Run("test1", func(t *testing.T) {
		input := "testdata/example_lab_data/REPORTE-CSV-BD-LAB/MERLIN/Biología Molecular/MERLIN-Biología Molecular-Carga Viral HIV1 Abbott RT-20180406223001.csv"
		actual := getKey(input)
		expected := "Biología Molecular/MERLIN-Biología Molecular-Carga Viral HIV1 Abbott RT"
		if actual != expected {
			t.Errorf("expected %s, got %s", expected, actual)
		}
		input = "testdata/example_lab_data/REPORTE-CSV-BD-LAB/MERLIN/BSL-III/SABES 1/BSL III/SABES 1-BSL III-Carga Viral HIV1 Xpert-20180406223001.csv"
		actual = getKey(input)
		expected = "BSL III/SABES 1-BSL III-Carga Viral HIV1 Xpert"
		if actual != expected {
			t.Errorf("expected %s, got %s", expected, actual)
		}
		input = "testdata/example_lab_data/REPORTE-CSV-BD-LAB/SABES 2A/BSL III/SABES 2A-BSL III-Carga Viral HIV1 Xpert-20180406223001.csv"
		actual = getKey(input)
		expected = "BSL III/SABES 2A-BSL III-Carga Viral HIV1 Xpert"
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
		// fmt.Println("tempDir is", tempDir)

		err = copy.Copy(filepath.Join("testdata", "example_lab_data"), tempDir)
		if err != nil {
			t.Fail()
		}
		err = processLabFiles(config, tempDir)
		if err != nil {
			t.Error(err.Error())
		}

	})
}

func TestPtidExists(t *testing.T) {
	// var df dataframe.DataFrame

	f, err := os.Open("testdata/example_lab_data/REPORTE-CSV-BD-LAB/MERLIN/Anglolab/MERLIN-Anglolab-Gamma Glutamil Transpeptidasa-20180406223001.csv")
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

func TestMergeDuplicateRows(t *testing.T) {
	t.Run("test1", func(t *testing.T) {

		path := filepath.Join("testdata", "semi-processed-lab-files",
			"citometria_SABES_3_citometria_linf_CD3_CD4_CD8.csv")

		f, err := os.Open(path)
		if err != nil {
			t.Fail()
		}
		defer f.Close()

		df := dataframe.ReadCSV(f,
			dataframe.HasHeader(true),
			dataframe.DetectTypes(false))

		out, err := mergeDuplicateRows(df)
		if err != nil {
			fmt.Println(err.Error())
			t.Fail()
			return
		}

		expectedRows := df.Nrow() - 2
		actualRows := out.Nrow()

		if actualRows != expectedRows {
			t.Error("expected", expectedRows, "rows in data frame, got", actualRows)
		}

		firstRow := out.Subset([]int{0})
		lastRow := out.Subset([]int{out.Nrow() - 1})

		actualFirstRow := firstRow.Records()[1][0]
		actualLastRow := lastRow.Records()[1][0]

		expectedFirstRow := "090000031652"
		expectedLastRow := "410000021960"

		if actualFirstRow != expectedFirstRow {
			t.Error("first row should be", expectedFirstRow, ", got", actualFirstRow)
		}

		if actualLastRow != expectedLastRow {
			t.Error("last row should be", expectedLastRow, ", got", actualLastRow)
		}

		df1 := out.Filter(
			dataframe.F{
				Colname:    "IdData",
				Comparator: series.Eq,
				Comparando: "410000009898"},
		)
		df2 := out.Filter(
			dataframe.F{
				Colname:    "IdData",
				Comparator: series.Eq,
				Comparando: "090000045264"},
		)

		d1 := df1.Col("FechaImpresion").Records()[0]
		d2 := df2.Col("FechaImpresion").Records()[0]

		if strings.Index(d1, " ") != -1 {
			t.Error("space occurs in FechaImpression")
		}

		if strings.Index(d2, " ") != -1 {
			t.Error("space occurs in FechaImpression")
		}

	})
}
