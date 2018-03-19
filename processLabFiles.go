package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"github.com/kniren/gota/dataframe"
	"github.com/kniren/gota/series"
	"github.com/yeka/zip"
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
	for k, v := range dfMap {
		filename, err := keyToFileName(k)
		if err != nil {
			return err
		}
		fullpath := filepath.Join(config.LocalDownloadFolderLab, filename)
		outfh, err := os.Create(fullpath)
		if err != nil {
			return err
		}
		err = v.WriteCSV(outfh, dataframe.WriteHeader(true))
		if err != nil {
			return err
		}
		outfh.Close()

		var buffer bytes.Buffer
		err0 := phiDataFrame.WriteCSV(&buffer, dataframe.WriteHeader(true))
		if err0 != nil {
			return err0
		}
		phiZipName := filepath.Join(config.LocalDownloadFolderLab, "phi.zip")
		fzip, err := os.Create(phiZipName)
		if err != nil {
			return err
		}
		zipw := zip.NewWriter(fzip)
		defer zipw.Close()
		w, err := zipw.Encrypt("phi.csv", config.PhiZipPassword, zip.StandardEncryption)
		if err != nil {
			return err
		}
		_, err = io.Copy(w, bytes.NewReader(buffer.Bytes()))
		if err != nil {
			return err
		}
		zipw.Flush()

	}
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

// This should probably be done with a regex but that gave a lot of problems
// with mixed-case abbreviations.
func toLowerExceptAbbreviations(in string) string {
	runes := []rune(in)
	max := len(runes) - 1

	for i, r := range runes {
		if unicode.IsUpper(r) {
			/* pseudocode:
			if i am uppercase and
				either there is no character before me or it is non-uppercase
					and
				either there is no character after me or it is non-uppercase
				then:
					change me to lowercase
				otherwise, leave me alone
			*/
			if unicode.IsUpper(r) && (i == 0 || !unicode.IsUpper(runes[i-1])) &&
				(i == max || !unicode.IsUpper(runes[i+1])) {
				runes[i] = unicode.ToLower(r)
			}
		}
	}
	out := string(runes)
	return out
}

func keyToFileName(key string) (string, error) {
	key, err := convertAccentedToPlain(key)
	if err != nil {
		return "", err
	}
	replaceMe := []string{".", "-", "(", ")", "/", " "}
	for _, badchar := range replaceMe {
		key = strings.Replace(key, badchar, "_", -1)
	}
	re := regexp.MustCompile("_{2,}")
	key = re.ReplaceAllString(key, "_")
	// failed regex attempt would not change e.g. CnC to cnc:
	// re2 := regexp.MustCompile("[^[[:upper:]]]*([[:upper:]]{1})[^[[:upper:]]]*")
	// key = " " + key // this is silly but i am dumb
	// matches := re2.FindAllStringSubmatch(key, -1)
	// for _, match := range matches {
	// 	key = strings.Replace(key, match[0], strings.ToLower(match[0]), -1)
	// }
	//
	// key = re2.ReplaceAllString(key, strings.ToLower("$1"))
	key = toLowerExceptAbbreviations(key)
	key += ".csv"
	// key = strings.TrimSpace(key)
	return key, nil
}
