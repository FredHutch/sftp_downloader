package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kniren/gota/dataframe"
	"github.com/kniren/gota/series"
)

var dfMap map[string]dataframe.DataFrame

var ptidMap map[string]bool

var phiDataFrame dataframe.DataFrame

func walkFn(path string, info os.FileInfo, err error) error {

	if info.IsDir() {
		return nil
	}
	if info.Size() == 0 {
		return nil
	}
	if !strings.HasSuffix(strings.ToLower(path), ".csv") {
		return nil
	}

	key := getKey(path)

	var df dataframe.DataFrame

	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	if dfMap == nil {
		dfMap = make(map[string]dataframe.DataFrame)
	}

	newDf := dataframe.ReadCSV(f,
		dataframe.HasHeader(true),
		dataframe.DetectTypes(false))

	dfFromMap, ok := dfMap[key]
	if ok {
		df = dfFromMap
	}

	// make a new df from the phi columms + ptid of newDf
	phiColumnNames := []string{"PTID", "Iniciales", "FechaNacimiento"}
	newPhi := newDf.Select(phiColumnNames)

	var indicesToKeep []int

	var wantedRows []int

	columnsToRemove := []string{"Iniciales", "FechaNacimiento"}
	for i, name := range newDf.Names() {
		if !stringInSlice(name, columnsToRemove) {
			indicesToKeep = append(indicesToKeep, i)
		}
	}

	// remove PHI columns from newDf
	newDf = newDf.Select(indicesToKeep)

	// append phi rows to phiDataFrame

	ptids := newPhi.Col("PTID").Records()

	if r, c := phiDataFrame.Dims(); r == 0 && c == 0 {
		for _, ptid := range ptids {
			ptidMap[ptid] = true
		}
		phiDataFrame = newPhi
	} else {
		//TODO check if pt already exists? or do that later w/uniq?

		for i, ptid := range ptids {
			if _, ok := ptidMap[ptid]; !ok {
				wantedRows = append(wantedRows, i)
			}
			ptidMap[ptid] = true
		}

		noDupes := newPhi.Subset(wantedRows)
		if noDupes.Nrow() > 0 {
			phiDataFrame = phiDataFrame.RBind(noDupes)
			if phiDataFrame.Err != nil {
				return phiDataFrame.Err
			}
		}
	}

	// append newDf rows to df
	if r, c := df.Dims(); r == 0 && c == 0 {
		df = newDf
	} else {
		df = df.RBind(newDf)
		if df.Err != nil {
			return df.Err
		}
	}
	// put df back in map
	dfMap[key] = df

	return nil

}

func getKey(path string) string {
	segs := strings.Split(path, string(os.PathSeparator))
	baseName := segs[len(segs)-1]
	labType := segs[len(segs)-2]

	bSegs := strings.Split(baseName, "-")
	return fmt.Sprintf("%s%s%s", labType, string(os.PathSeparator),
		strings.Join(bSegs[0:len(bSegs)-1], "-"))
}

func processLabFiles(config Config, rawLabFileDir string) error {
	ptidMap = make(map[string]bool)
	err := filepath.Walk(rawLabFileDir, walkFn)
	if err != nil {
		return err
	}
	// fmt.Println("dfMap:")
	// fmt.Println(dfMap)
	// fmt.Println("phiDataFrame:")
	// fmt.Println(phiDataFrame)
	return nil
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func ptidExists(ptid string, df dataframe.DataFrame) bool {
	filt := df.Filter(
		dataframe.F{
			Colname:    "PTID",
			Comparator: series.Eq,
			Comparando: ptid,
		},
	)
	return filt.Nrow() > 0
}
