package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
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

func TestMergeFunction(t *testing.T) {
	t.Run("merge-blank", func(t *testing.T) {
		df := dataframe.LoadRecords(
			[][]string{
				{"A", "B", "C", "D"},
				{"a", "", "c", "d"},
				{"a", "b", "c", "d"},
			},
		)
		actual := merge(df, []int{0, 1})
		if actual.Nrow() != 1 {
			t.Errorf("Expected 1 row, got %d.", actual.Nrow())
		}
		expected := []string{"a", "b", "c", "d"}
		data := actual.Subset(0).Records()[1]
		if !strEqual(data, expected) {
			t.Error("Expected", expected, "got", data)
		}
		fmt.Println(data)
	})

	t.Run("merge-identical", func(t *testing.T) {
		df := dataframe.LoadRecords(
			[][]string{
				{"A", "B", "C", "D"},
				{"a", "b", "c", "d"},
				{"a", "b", "c", "d"},
			},
		)
		actual := merge(df, []int{0, 1})
		if actual.Nrow() != 1 {
			t.Errorf("Expected 1 row, got %d.", actual.Nrow())
		}
		expected := []string{"a", "b", "c", "d"}
		data := actual.Subset(0).Records()[1]
		if !strEqual(data, expected) {
			t.Error("Expected", expected, "got", data)
		}
		fmt.Println(data)
	})

	t.Run("cannot-merge", func(t *testing.T) {
		df := dataframe.LoadRecords(
			[][]string{
				{"A", "B", "C", "D"},
				{"a", "b", "c", "e"},
				{"a", "b", "c", "d"},
			},
		)
		fmt.Println("---")
		// col := df.Col("B")
		// col.Set(0, series.Strings("x"))
		// col.Set(1, series.Strings("y"))
		// df = df.Mutate(col)
		// fmt.Println(df)
		// t.Fail()
		fmt.Println("---")

		actual := merge(df, []int{0, 1})
		if actual.Nrow() != 2 {
			t.Errorf("Expected 2 rows, got %d.", actual.Nrow())
		}

		ok := true
		for i := 0; i < actual.Nrow(); i++ {
			if !strEqual(df.Subset(i).Records()[1], actual.Subset(i).Records()[1]) {
				ok = false
				break
			}
		}
		if !ok {
			t.Error("expected:", df, "got:", actual)
		}
	})

}

func TestGetAllPairs(t *testing.T) {
	t.Run("foot-test", func(t *testing.T) {
		x := getAllPairs(6)
		expected := [][]int{
			{0, 1}, {0, 2}, {0, 3}, {0, 4}, {0, 5}, {1, 2}, {1, 3}, {1, 4}, {1, 5}, {2, 3}, {2, 4}, {2, 5}, {3, 4}, {3, 5}, {4, 5},
		}
		if len(x) != len(expected) {
			t.Errorf("expected length %d; got length %d", len(expected), len(x))
		} else {
			ok := true
			for i := 0; i < len(x); i++ {
				if !Equal(x[i], expected[i]) {
					ok = false
					break
				}
			}
			if !ok {
				t.Error("expected:\n", expected, "got:\n", x)
			}

		}
	})
}

//TODO merge with the rest of the TestMergeDuplicateRows tests
func TestMergeDuplicateRows3(t *testing.T) {
	t.Run("multiple-mergeable-and-non-mergeable", func(t *testing.T) {
		df0 := dataframe.LoadRecords(
			[][]string{
				{"IdData", "B"},
				{"aoo", "bar"},    // 1. Cannot be merged
				{"aoo", "bat"},    // 2. Cannot be merged.
				{"aoo", "quux"},   // 3. Can be merged with 4.
				{"aoo", "quux"},   // 4.
				{"bar", "cluck"},  // 5. Can be merged with 6 and 7.
				{"bar", "cluck"},  // 6.
				{"bar", "cluck"},  // 7.
				{"bar", "hodad"},  // 8. Doesn't match anything else so can't be merged.
				{"bar", "oblast"}, // 9. Can be merged with 10
				{"bar", "oblast"}, // 10.
			},
		)

		// TODO save this into a variable when ready
		_ = dataframe.LoadRecords(
			[][]string{
				{"IdData", "B", "C", "D"},
				{"aoo", "-", "", ""},  // 1. 1-3 can be merged.
				{"aoo", "", "-", ""},  // 2.
				{"aoo", "", "", "-"},  // 3.
				{"aoo", "+", "", ""},  // 4. 4 and 5 can be merged.
				{"aoo", "+", "", "+"}, // 5.
				// {"bar", "cluck"},     // 5. Can be merged with 6 and 7.
				// {"bar", "cluck"},     // 6.
				// {"bar", "cluck"},     // 7.
				// {"bar", "hodad"},     // 8. Doesn't match anything else so can't be merged.
				// {"bar", "oblast"},    // 9. Can be merged with 10
				// {"bar", "oblast"},    // 10.
			},
		)

		_, err := mergeDuplicateRows(df0)
		if err != nil {
			t.Error("unexpected error:", err.Error())
			return
		}
		// if out.Nrow() != 1 {
		// 	t.Error("expected 1 row, got", out.Nrow())
		// }

	})
}

//TODO merge with the rest of the TestMergeDuplicateRows tests
func TestMergeDuplicateRows2(t *testing.T) {

	t.Run("multiple-IdData-rows", func(t *testing.T) {

		path := filepath.Join("testdata", "multiple-id-data.csv")
		f, err := os.Open(path)
		if err != nil {
			t.Fail()
		}
		defer f.Close()
		df := dataframe.ReadCSV(f,
			dataframe.HasHeader(true),
			dataframe.DefaultType(series.String),
			dataframe.DetectTypes(false))

		// fmt.Println("---\n", df, "n---")

		merged, err := mergeDuplicateRows(df)
		if err != nil {
			fmt.Println(err.Error())
			t.Fail()
			return
		}

		// for i := 0; i < merged.Nrow(); i++ {
		// 	fmt.Println(merged.Records()[i])
		// }

		if true {
			return
		}

		if merged.Nrow() != 1 {
			t.Error("expected 1 row, got", merged.Nrow())
		}

		type mergeTest struct {
			n        string
			expected string
		}
		var mergeTests = []mergeTest{
			{"090000062236", "090000062236"},
			{"San Miguel", "San Miguel"},
			{"32158-6972-6", "32158-6972-6"},
			{"SHZ", "SHZ"},
			{"13/23/2575", "13/23/2575"},
			{"10/02/2020", "10/02/2020"},
			{"11/02/2020 03:36:35 p.m.", "11/02/2020 03:36:35 p.m."},
			{"INT,W230", "INT,W230"},
			{"Si", "Si"},
			{"CT no detectado", "CT no detectado"},
			{"NG no detectado", "NG no detectado"},
			{"10/02/2020", "10/02/2020"},
			{"Las muestras de orina guardada", "Las muestras de orina guardada"},
		}

		// fmt.Println(merged)
		// fmt.Println("^^^^^")
		records := merged.Records()[1]
		for idx, tt := range mergeTests {
			actual := records[idx]
			if actual != tt.expected {
				t.Errorf("expected %s, actual %s", tt.expected, actual)
			}
		}

	})

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

	t.Run("rows can't be merged", func(t *testing.T) {
		df0 := dataframe.LoadRecords(
			[][]string{
				{"IdData", "B"},
				{"foo", "bar"},
				{"foo", "bat"},
			},
		)
		out, err := mergeDuplicateRows(df0)
		if err != nil {
			t.Error("unexpected error:", err.Error())
			return
		}
		if out.Nrow() != 2 {
			t.Errorf("Expected 2 rows, got %d", out.Nrow())
		}
		// Don't understand the following warning, neither arg is an error!
		if !reflect.DeepEqual(df0, out) {
			t.Error("dataframes do not match")
		}
	})

	t.Run("merge identical records", func(t *testing.T) {
		df0 := dataframe.LoadRecords(
			[][]string{
				[]string{"IdData", "B"},
				[]string{"foo", "bat"},
				[]string{"foo", "bat"},
			},
		)
		out, err := mergeDuplicateRows(df0)
		if err != nil {
			t.Error("unexpected error:", err.Error())
			return
		}
		if out.Nrow() != 1 {
			t.Error("expected 1 row, got", out.Nrow())
		}
	})

}
