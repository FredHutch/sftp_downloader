package main

import (
	"fmt"
	"os"
	path0 "path"
	"path/filepath"
	"strings"
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

func moveFiles(root string) error {
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if !predicate(root, path) {
			return nil
		}
		newname := filepath.Join(root, filepath.Base(path))
		ok, err := FileExists(newname)
		if err != nil {
			return err
		}
		if ok {
			return fmt.Errorf("File %s already exists", newname)
		}
		err = Rename(path, newname)
		return err
	})
	return err
}
