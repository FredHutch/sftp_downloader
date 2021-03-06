package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
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

func convertDate(s series.Series) series.Series {
	if strings.Index(strings.ToLower(s.Name), "fecha") > -1 {
		var out []string
		strs := s.Records()
		for _, item := range strs {
			changed := changeDate(item)
			out = append(out, changed)
		}
		return series.Strings(out)
	}
	return s
}

func existingColumns(df dataframe.DataFrame, columnNames []string) []string {
	var out []string
	actualNames := df.Names()
	for i := 0; i < len(actualNames); i++ {
		for j := 0; j < len(columnNames); j++ {
			if columnNames[j] == actualNames[i] {
				out = append(out, actualNames[i])
			}
		}
	}
	return out
}

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
	segments := strings.Split(path, "/")
	project := segments[len(segments)-3]

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

	newDf = newDf.Capply(convertDate)

	if !strIsInSlice(newDf.Names(), "Proyecto") {
		recs := make([]string, newDf.Nrow())
		for i := 0; i < newDf.Nrow(); i++ {
			recs[i] = project
		}
		col := series.New(recs, series.String, "Proyecto")
		newDf = insertColumn(newDf, col, 2)
	}

	newDf, err = mergeDuplicateRows(newDf)
	if err != nil {
		fmt.Println("debug2")
		fmt.Println("got an error:", err.Error(), "path is", path)
		return err
	}

	dfFromMap, ok := dfMap[key]
	if ok {
		df = dfFromMap
	}

	// make a new df from the phi columms + ptid of newDf
	phiColumnNames := existingColumns(newDf, []string{"PTID", "Iniciales", "FechaNacimiento"})
	newPhi := newDf.Select(phiColumnNames)

	var indicesToKeep []int

	var wantedRows []int

	columnsToRemove := existingColumns(newPhi, []string{"Iniciales", "FechaNacimiento"})
	for i, name := range newDf.Names() {
		if !stringInSlice(name, columnsToRemove) {
			indicesToKeep = append(indicesToKeep, i)
		}
	}

	// remove PHI columns from newDf
	newDf = newDf.Select(indicesToKeep)

	// append phi rows to phiDataFrame

	ptids := newPhi.Col("PTID").Records()

	if ptidMap == nil {
		ptidMap = make(map[string]bool)
	}

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

func insertColumn(df dataframe.DataFrame, col series.Series, insertBefore int) dataframe.DataFrame {
	var seriesList []series.Series
	for i := 0; i < df.Ncol(); i++ {
		if i == insertBefore {
			seriesList = append(seriesList, col)
		}
		seriesList = append(seriesList, df.Col(df.Names()[i]))
	}
	return dataframe.New(seriesList...)
}

func compareSlices(s1 []string, s2 []string) bool {
	if len(s1) != len(s2) {
		return false
	}
	for i := 0; i < len(s1); i++ {
		if s1[i] != s2[i] {
			return false
		}
	}
	return true
}

func allTheSame(list []string) bool {
	// invariants: len(list) will be > 1
	for i := 0; i < len(list); i++ {
		if i > 0 {
			if list[i] != list[i-1] {
				return false
			}
		}
	}
	return true
}

func getMergedValue(col series.Series) (string, bool) {
	// invariants (should assert...):
	// - all IdData values will be the same
	// - there will be more than 1 row in the column
	records := col.Records()
	if allTheSame(records) { // should be true of IdData column....
		return records[0], true
	}
	empties := 0
	var nonEmpty int
	for i := 0; i < col.Len(); i++ { // TODO convert to range?
		record := records[i]
		if record == "" {
			empties++
		} else {
			nonEmpty = i
		}
	}
	if empties == (col.Len() - 1) {
		return records[nonEmpty], true
	}
	return "", false
}

func merge(df dataframe.DataFrame) dataframe.DataFrame {
	// invariants: all IdData values will be the same
	// there will be >= 2 rows

	singleRec := make([]string, df.Ncol())
	for colIdx := 0; colIdx < df.Ncol(); colIdx++ {
		col := df.Col(df.Names()[colIdx])
		mergedValue, canBeMerged := getMergedValue(col)
		if !canBeMerged { // if any column cannot be merged,
			return df // then this whole data frame cannot be merged. (is this true?)
		}
		singleRec[colIdx] = mergedValue

	}
	dfOut := dataframe.LoadRecords([][]string{
		singleRec,
	}, dataframe.DefaultType(series.String), dataframe.DetectTypes(false), dataframe.HasHeader(false))
	dfOut.SetNames(df.Names()...)
	return dfOut
}

// TODO do this for clinical files as well as lab files?
func mergeDuplicateRows(df dataframe.DataFrame) (dataframe.DataFrame, error) {
	var m map[string]int
	m = make(map[string]int)

	if strIsInSlice(df.Names(), "FechaImpresion") {
		col := df.Col("FechaImpresion")
		for i := 0; i < col.Len(); i++ {
			val := col.Val(i).(string)
			val = strings.Split(val, " ")[0]
			col.Set(i, series.Strings(val))
		}
		df = df.Mutate(col)
	}

	idDataCol := df.Select([]string{"IdData"}).Col("IdData").Records()

	for _, idData := range idDataCol {
		elem, ok := m[idData]
		if ok {
			m[idData] = elem + 1
		} else {
			m[idData] = 1
		}
	}

	// Loop:
	for k, v := range m {
		if v > 1 {
			fil := df.Filter(
				dataframe.F{
					Colname:    "IdData",
					Comparator: series.Eq,
					Comparando: k,
				})

			remainder := df.Filter(
				dataframe.F{
					Colname:    "IdData",
					Comparator: series.Neq,
					Comparando: k,
				})

			merged := merge(fil)

			df = remainder.RBind(merged) // TODO handle error if df.Err != nil??

			df = df.Arrange(
				dataframe.Sort("IdData"),
			)

		}
	}

	return df, nil

}

func strIsInSlice(slice []string, searchString string) bool {
	for _, item := range slice {
		if item == searchString {
			return true
		}
	}
	return false
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
	errp := filepath.Walk(rawLabFileDir, walkFn)
	if errp != nil {
		return errp
	}

	for k, v := range dfMap {
		filename, err1 := keyToFileName(k)
		if err1 != nil {
			return err1
		}
		fullpath := filepath.Join(rawLabFileDir, filename)
		outfh, errz := os.Create(fullpath)
		defer outfh.Close()
		if errz != nil {
			return errz
		}
		err := v.WriteCSV(outfh, dataframe.WriteHeader(true))
		if err != nil {
			return err
		}
		// outfh.Close()

		var buffer bytes.Buffer
		err0 := phiDataFrame.WriteCSV(&buffer, dataframe.WriteHeader(true))
		if err0 != nil {
			return err0
		}
		phiZipName := filepath.Join(rawLabFileDir, "phi.zip")
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

	m, err := groupFilesForCombining(rawLabFileDir)
	if err != nil {
		return err
	}
	for k, v := range m {
		combineCsvs(rawLabFileDir, k, v...)
		// TODO rename files that did NOT get combined so that they do not have lab type twice
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
	// negative lookahead would fix this but it's not supported in golang's regexp
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

func labCleanup(config Config) error {
	files, err := ioutil.ReadDir(config.LocalDownloadFolderLab)
	if err != nil {
		return err
	}
	for _, file := range files {
		fullPath := filepath.Join(config.LocalDownloadFolderLab, file.Name())
		if strings.HasSuffix(file.Name(), ".rar") {
			err = os.Remove(fullPath)
			if err != nil {
				return err
			}
		}
		if file.IsDir() {
			err = os.RemoveAll(fullPath)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
