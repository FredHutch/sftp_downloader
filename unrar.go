package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"github.com/nwaples/rardecode"
)

// Open extracts the RAR file at source and puts the contents
// into destination.
func Open(source, destination string, password string) error {
	rf, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("%s: failed to open file: %v", source, err)
	}
	defer rf.Close()

	return Read(rf, destination, password)
}

// Read extracts the RAR file read from input and puts the contents
// into destination.
func Read(input io.Reader, destination string, password string) error {
	rr, err := rardecode.NewReader(input, password)
	if err != nil {
		return fmt.Errorf("read: failed to create reader: %v", err)
	}

	for {
		header, err := rr.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		if header.IsDir {
			err = mkdir(filepath.Join(destination, header.Name))
			if err != nil {
				return err
			}
			continue
		}

		// if files come before their containing folders, then we must
		// create their folders before writing the file
		err = mkdir(filepath.Dir(filepath.Join(destination, header.Name)))
		if err != nil {
			return err
		}

		err = writeNewFile(filepath.Join(destination, header.Name), rr, header.Mode())
		if err != nil {
			return err
		}
	}

	return nil
}

func mkdir(dirPath string) error {
	err := os.MkdirAll(dirPath, 0755)
	if err != nil {
		return fmt.Errorf("%s: making directory: %v", dirPath, err)
	}
	return nil
}

func writeNewFile(fpath string, in io.Reader, fm os.FileMode) error {
	err := os.MkdirAll(filepath.Dir(fpath), 0755)
	if err != nil {
		return fmt.Errorf("%s: making directory for file: %v", fpath, err)
	}

	out, err := os.Create(fpath)
	if err != nil {
		return fmt.Errorf("%s: creating new file: %v", fpath, err)
	}
	defer out.Close()

	err = out.Chmod(fm)
	if err != nil && runtime.GOOS != "windows" {
		return fmt.Errorf("%s: changing file mode: %v", fpath, err)
	}

	_, err = io.Copy(out, in)
	if err != nil {
		return fmt.Errorf("%s: writing file: %v", fpath, err)
	}
	return nil
}

// UncompressFile uncompresses a rar file with a password
func UncompressFile(rarFile, fileDate string, config Config, phase Phase) error {
	destFolder := filepath.Join(getDownloadFolder(phase, config), fileDate)
	exists, err := FileExists(destFolder)
	if err != nil {
		return fmt.Errorf("Error checking directory existence")
	}
	if exists {
		return fmt.Errorf("Uncompress destination directory '%s' already exists", destFolder)
	}
	err = os.Mkdir(destFolder, os.ModePerm)
	err = Open(rarFile, destFolder, config.RarDecryptionPassword)
	if err != nil {
		return fmt.Errorf(err.Error())
	}
	return nil
}
