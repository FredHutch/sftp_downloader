package main

import (
	"fmt"
	"os"
	path0 "path"
	"path/filepath"
	"strings"

	"github.com/kniren/gota/dataframe"
)

// Rename can be overridden for testing
var Rename = os.Rename

func predicate(root string, path string) bool {
	ok, _ := IsDir(path)
	if ok {
		return false
	}
	low := strings.ToLower(path)
	if !(strings.HasSuffix(low, ".csv") || strings.HasSuffix(low, ".sav")) {
		return false
	}
	return path0.Dir(root) != path0.Dir(path)
}

func removeSuffix(filename string) string {
	dotsegs := strings.Split(filename, ".")
	extension := dotsegs[len(dotsegs)-1]
	segs := strings.Split(filename, "-")
	if len(segs) > 1 {
		segs = segs[:len(segs)-1]
	}
	out := strings.Join(segs, "-")
	if !strings.HasSuffix(out, fmt.Sprintf(".%s", extension)) {
		out = fmt.Sprintf("%s.%s", out, extension)
	}
	return out
}

func changeClinicalDates(path, newname string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	newDf := dataframe.ReadCSV(f,
		dataframe.HasHeader(true),
		dataframe.DetectTypes(false))

	newDf = newDf.Capply(convertDate)
	outfh, err := os.Create(newname)
	if err != nil {
		return err
	}
	err = newDf.WriteCSV(outfh, dataframe.WriteHeader(true))
	if err != nil {
		return err
	}
	outfh.Close()

	return nil

}

func moveFiles(root string) error {
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if !predicate(root, path) {
			return nil
		}
		newname := filepath.Join(root, removeSuffix(filepath.Base(path)))
		ok, err := FileExists(newname)
		if err != nil {
			return err
		}
		if ok {
			return fmt.Errorf("File %s already exists", newname)
		}

		if strings.HasSuffix(newname, ".sav") || info.Size() == 0 {
			err = Rename(path, newname)
			return err
		}
		err = changeClinicalDates(path, newname)
		return err
	})

	if err != nil {
		return err
	}

	fh, err := os.Open(root)
	if err != nil {
		return fmt.Errorf("Could not open %s for reading", root)
	}
	defer fh.Close()
	infos, err := fh.Readdir(-1)
	if err != nil {
		return fmt.Errorf("Could not read directory %s", root)
	}
	for _, info := range infos {
		if info.IsDir() {
			err = os.RemoveAll(filepath.Join(root, info.Name()))
			if err != nil {
				return fmt.Errorf("Could not remove subdirectory %s", info.Name())
			}
		}
	}

	return nil
}
