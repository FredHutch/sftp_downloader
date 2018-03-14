package main

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kniren/gota/dataframe"
	"github.com/kniren/gota/series"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var upperCaseLetters = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ")

var numbers = []rune("0123456789")

func randStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = upperCaseLetters[rand.Intn(len(upperCaseLetters))]
	}
	return string(b)
}

func randNums(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = numbers[rand.Intn(len(numbers))]
	}
	return string(b)
}

func walkies(path string, info os.FileInfo, err error) error {
	if err != nil {
		fmt.Println("got an error:", err.Error())
		return err
	}

	if strings.HasSuffix(strings.ToLower(path), ".csv") {
		fmt.Println("yay", path)
		if info.Size() == 0 {
			fmt.Println(path, "is empty, skipping")
			return nil
		}
		f, err := os.Open(path)
		if err != nil {
			panic(err.Error())
		}

		df := dataframe.ReadCSV(f,
			dataframe.HasHeader(true),
			dataframe.DetectTypes(false))

		f.Close()

		fmt.Println("here is your data frame")
		fmt.Println(df.Names())
		fmt.Println(df)
		df2 := anonymizeInitials(df)
		df3 := anonymizePtID(df2)
		df4 := anonymizeDate(df3)
		fmt.Println("initials and ptid and date anonymized:")
		fmt.Println(df4)
		// pathdir := filepath.Dir(path)
		// newfile := filepath.Join(pathdir, "new.csv") // FIXME change this

		outfh, err := os.Create(path)
		if err != nil {
			panic(err)
		}
		err = df4.WriteCSV(outfh)
		outfh.Close()

		// os.Exit(0) // FIXME remove

	}
	return nil
}

func anonymizeInitials(df dataframe.DataFrame) dataframe.DataFrame {
	var s []string
	for i := 0; i < df.Nrow(); i++ {
		s = append(s, randStringRunes(3))
	}
	df2 := df.Mutate(series.New(s, series.String, "Iniciales"))
	return df2
}

func anonymizePtID(df dataframe.DataFrame) dataframe.DataFrame {
	var s []string
	for i := 0; i < df.Nrow(); i++ {
		s = append(s, fmt.Sprintf("%s-%s-%s", randNums(5), randNums(4), randNums(1)))
	}
	df2 := df.Mutate(series.New(s, series.String, "PTID"))
	return df2
}

func anonymizeDate(df dataframe.DataFrame) dataframe.DataFrame {
	var s []string
	for i := 0; i < df.Nrow(); i++ {
		s = append(s, fmt.Sprintf("%s/%s/%s", randNums(2), randNums(2), randNums(4)))
	}
	df2 := df.Mutate(series.New(s, series.String, "FechaNacimiento"))
	return df2
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Supply the name of a directory of csv files to anonymize.")
		os.Exit(1)
	}
	csvdir := os.Args[1]
	err := filepath.Walk(csvdir, walkies)
	if err != nil {
		panic(err)
	}
}
