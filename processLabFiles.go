package main

import (
	"bytes"
	"errors"
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

	newDf = newDf.Capply(convertDate)

	newDf, err = mergeDuplicateRows(newDf)
	if err != nil {
		fmt.Println("got an error:", err.Error(), "path is", path)
		return err
	}

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

func allTheSame(cols series.Series) bool {
	same := true
	records := cols.Records()
	for i := 1; i < len(records); i++ {
		same = records[i] == records[i-1]
		if !same {
			return false
		}
	}
	return true
}

func getNonBlankRecord(cols series.Series) (string, error) {
	records := cols.Records()
	// fmt.Println("records are:\n", records)
	nonBlankRecords := 0
	var idx int
	for i := 0; i < len(records); i++ {
		if records[i] != "" {
			nonBlankRecords++
			idx = i
		}
	}
	if nonBlankRecords > 1 {
		return "", errors.New("more than one non-blank record")
	}
	if nonBlankRecords == 0 {
		return "", errors.New("no non-blank records")
	}
	return records[idx], nil
}

func getMergeableValue(cols series.Series) (string, error) {
	if allTheSame(cols) {
		return cols.Records()[0], nil
	}
	nbr, err := getNonBlankRecord(cols)
	if err != nil {
		return "", err
	}
	return nbr, nil

}

func unique(strSlice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range strSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func strIsInSlice(slice []string, searchString string) bool {
	for _, item := range slice {
		if item == searchString {
			return true
		}
	}
	return false
}

func isInSlice(slice [][]int, item []int) bool {
	for _, row := range slice {
		if Equal(row, item) {
			return true
		}
	}
	return false
}

// Equal tells whether a and b contain the same elements.
// A nil argument is equivalent to an empty slice.
func Equal(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func strEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func getAllPairs(numOfItems int) [][]int {
	// out := make([][]int, numOfItems)
	out := [][]int{}
	for x := 0; x < numOfItems; x++ {
		for y := 0; y < numOfItems; y++ {
			//    x is 0 and y is 1 the first time through
			// if x is 1 and y is 0, we want to skip it.
			var row []int
			if x == y {
				continue
			}
			var reverse []int
			reverse = append(reverse, y)
			reverse = append(reverse, x)
			if isInSlice(out, reverse) {
				continue
			}
			row = append(row, x)
			row = append(row, y)
			out = append(out, row)
		}
	}
	return out
}

func merge(df dataframe.DataFrame, pair []int) dataframe.DataFrame {
	// fmt.Println("pair is", pair)
	firstRow := df.Subset(pair[0]).Records()[1]
	secondRow := df.Subset(pair[1]).Records()[1]
	// fmt.Println("firstRow is\n", firstRow)
	// fmt.Println("secondRow is\n", secondRow)

	if strEqual(firstRow, secondRow) {
		// return a new dataframe with just one of the rows
		// since they are identical
		newDf := dataframe.LoadRecords(
			[][]string{
				firstRow,
			}, dataframe.HasHeader(false),
		)
		newDf.SetNames(df.Names()...)
		return newDf
	}
	// fmt.Println("hi0")
	newRecord := []string{}
	canBeMerged := true
	for i := 0; i < len(firstRow); i++ {
		firstItem := firstRow[i]
		secondItem := secondRow[i]
		if firstItem == secondItem {
			// fmt.Println("items are equal")
			newRecord = append(newRecord, firstItem)
		} else if firstItem != "" && secondItem == "" {
			// fmt.Println("first item not empty, second item empty")
			newRecord = append(newRecord, firstItem)
		} else if firstItem == "" && secondItem != "" {
			// fmt.Println("first item empty, second item not empty")
			newRecord = append(newRecord, secondItem)
		} else {
			// fmt.Println("...else...")
			canBeMerged = false
			break
		}
	}
	// fmt.Println("hi1")

	if !canBeMerged {
		// fmt.Println("hi2")

		// rows cannot be merged, just return a df of our input rows
		return df.Subset(pair)
	}

	newDf := dataframe.LoadRecords(
		[][]string{
			newRecord,
		}, dataframe.HasHeader(false),
		dataframe.DetectTypes(false),
		dataframe.DefaultType(series.String),
	)
	newDf.SetNames(df.Names()...)
	fmt.Println(newDf)
	// fmt.Println("hi3")

	return newDf

}

// TODO do this for clinical files as well as lab files?
func mergeDuplicateRows(df dataframe.DataFrame) (dataframe.DataFrame, error) {

	numRows := 0 // df.Nrow()

	var rowHash map[string]int
	rowHash = make(map[string]int)

	for i := 0; i < df.Nrow(); i++ {
		row := df.Subset(i).Records()[1]
		hashStr := strings.Join(row[:], "\t")
		rowHash[hashStr] = 1
	}

	// for numRows != df.Nrow() {
	for {
		pairs := getAllPairs(df.Nrow())
		newDf := dataframe.New()
		newDf.SetNames(df.Names()...)
		for _, pair := range pairs {
			mret := merge(df, pair)
			for i := 0; i < mret.Nrow(); i++ {
				rowDf := mret.Subset(i)
				hashStr := strings.Join(rowDf.Records()[1][:], "\t")
				_, inHash := rowHash[hashStr]
				if !inHash {
					newDf = newDf.RBind(rowDf)
				}
			}

		}
		df = newDf
		if numRows == df.Nrow() { // no change
			break
		} else {
			numRows = df.Nrow()
		}
	}

	df = df.Arrange(
		dataframe.Sort("IdData"),
	)

	if strIsInSlice(df.Names(), "FechaImpresion") {
		col := df.Col("FechaImpresion")
		for i := 0; i < col.Len(); i++ {
			val := col.Val(i).(string)
			val = strings.Split(val, " ")[0]
			col.Set(i, series.Strings(val))
		}
		df = df.Mutate(col)
	}
	return df, nil

	/*

	   pseudocode v 2.0

	   outer loop:

	   x = df.nrow()

	   def merge(pair):
	   	... can we merge?
	   	if yes:
	   		... do merge
	   		return df containing 1 merged row
	   	else:
	   		return df containing both rows in pair

	   while x != df.nrow: # or < instead of != # did nrow change in last iteration?
	   	newdf = df()
	   	rowHash = getHashOfRows(df) # keys are rows (series), values unimportant
	   	pairs = getAllPairs(df)
	   	for pair in pairs:
	   		ret = merge(pair)
	   		for row in ret:
	   			if row not in rowHash:
	   				newdf.appendrow(row)

	   sort newdf by IdData
	   fix fechaimpresion
	   return newdf





	*/

	// var m map[string]int
	// m = make(map[string]int)

	// idDataCol := df.Select([]string{"IdData"}).Col("IdData").Records()

	// fmt.Println("idDataCol is\n", idDataCol)

	// for _, idData := range idDataCol {
	// 	elem, ok := m[idData]
	// 	if ok {
	// 		m[idData] = elem + 1
	// 	} else {
	// 		m[idData] = 1
	// 	}
	// }

	// // Loop:
	// for k, v := range m {
	// 	if v > 1 {

	// 		// fil may contain some rows that can be merged
	// 		// and others which cannot. Those which cannot
	// 		// can be added on to remainder.

	// 		/*
	// 			Actually, there could be several sets of rows that can
	// 			be merged (the rows in each set could be merged with each
	// 			other but not with those of other sets).

	// 			Pseudocode:

	// 			- Create a slice of data frames above this loop (dfSlice)
	// 			- In this loop, for each instance of fil:
	// 				- starting with the first row, see if it can be merged with any
	// 				  subsequent rows (if so, do it and make a new DF out of
	// 				  the result and put it in dfSlice). All rows that
	// 				  were merged with the current row
	// 				  can now be either removed from fil or skipped over
	// 				  in this loop. (How exactly to do that?)
	// 				- If the current row cannot be merged, we can actually treat
	// 				  it the same as above (create a new DF with just one row and add
	// 				  it to dfSlice).
	// 				- rbind all the rows in dfSlice (plus remainder) together into df
	// 				- sort df by IdData.

	// 			This seems kind of clunky....

	// 			Maybe we want a function canBeMerged() which takes a row and
	// 			then a list of candidate rows and returns two things:
	// 					1) rows that can be merged with the row passed in; and
	// 					2) rows which cannot.

	// 			Or perhaps the function can take all rows that share the same IdData and
	// 			just return the merged data frame....Not sure if that is easier.

	// 			What would the pseudocode of that function look like?

	// 		*/

	// 		fil := df.Filter(
	// 			dataframe.F{
	// 				Colname:    "IdData",
	// 				Comparator: series.Eq,
	// 				Comparando: k,
	// 			})

	// 		remainder := df.Filter(
	// 			dataframe.F{
	// 				Colname:    "IdData",
	// 				Comparator: series.Neq,
	// 				Comparando: k,
	// 			})

	// 		// TODO first we should divide fil into rows that
	// 		// can be merged and rows that cannot.

	// 		singleRec := make([]string, fil.Ncol())
	// 		//iterate over the columns of fil
	// 		for i := 0; i < fil.Ncol(); i++ {
	// 			cols := fil.Col(fil.Names()[i])
	// 			val, err := getMergeableValue(cols)
	// 			if err != nil {
	// 				return df, err
	// 			}
	// 			singleRec[i] = val
	// 		}

	// 		single := dataframe.LoadRecords(
	// 			[][]string{
	// 				df.Names(),
	// 				singleRec,
	// 			},
	// 			dataframe.DetectTypes(false),
	// 			dataframe.DefaultType(series.String),
	// 		)

	// 		df = remainder.RBind(single)

	// 		// TODO sort here by IdData

	// 		df = df.Arrange(
	// 			dataframe.Sort("IdData"),
	// 		)

	// 		// fmt.Println(rows)
	// 	}
	// }

	// return df, nil

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
		if errz != nil {
			return errz
		}
		err := v.WriteCSV(outfh, dataframe.WriteHeader(true))
		if err != nil {
			return err
		}
		outfh.Close()

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
