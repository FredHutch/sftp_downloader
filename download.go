package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/FredHutch/sftp_downloader/iface"
)

// TODO refactor into smaller functions
func downloadFile(fileDate string, config Config, sftpclient iface.Sftper) (rarFile string, retErr error) {
	filePattern := fmt.Sprintf("Reportes_diarios_acumulados-%s", fileDate)
	remoteDir := "/" // TODO factor this out to json config
	matchingFiles, err := sftpclient.ReadDir(remoteDir)
	if err != nil {
		return "", fmt.Errorf("Error in ReadDir: %s", err.Error())
	}
	predicate := func(item string) bool {
		return strings.HasPrefix(item, filePattern)
	}
	matches := filter(matchingFiles, predicate)
	if len(matches) == 0 {
		return "", fmt.Errorf("Found no file matching pattern '%s'", filePattern)
	} else if len(matches) > 1 {
		return "", fmt.Errorf("Found more than one file matching pattern '%s'.", filePattern)
	}
	remoteFile := fmt.Sprintf("%s/%s", remoteDir, matches[0])
	f, err := sftpclient.Open(remoteFile)
	if err != nil {
		return "", fmt.Errorf("Could not open remote file %s: %s", remoteFile, err.Error())
	}
	// check that destination directory exists
	exists, err := FileExists(config.LocalDownloadFolder)
	if err != nil {
		return "", fmt.Errorf("error checking if local download folder exists: %s", err.Error())
	}
	if !exists {
		return "", fmt.Errorf("Local download directory '%s' does not exist.",
			config.LocalDownloadFolder)
	}

	isdir, err := IsDir(config.LocalDownloadFolder)
	if err != nil {
		return "", fmt.Errorf("Error checking if local download dir is a dir: %s", err.Error())
	}

	if !isdir {
		return "", fmt.Errorf("'%s' exists but is not a directory", config.LocalDownloadFolder)
	}

	destFile := filepath.Join(config.LocalDownloadFolder, matches[0])
	outfile, err := os.Create(destFile)
	if err != nil {
		return "", fmt.Errorf("Error creating destination file '%s'", destFile)
	}
	defer func() {
		if err := outfile.Close(); err != nil {
			retErr = fmt.Errorf("Error closing outfile: %s", err.Error())
		}
	}()
	buf := make([]byte, 1024)
	for {
		// read a chunk
		log.Println("inside Read loop")
		n, err := f.Read(buf)
		log.Println("after Read call")
		if err != nil && err != io.EOF {
			log.Println("got premature eof")
			return "", fmt.Errorf("premature EOF on input file: %s", err.Error())
		}
		if n == 0 {
			log.Println("n==0, breaking")
			break
		}

		// write a chunk
		log.Println("writing a chnk...")
		if _, err := outfile.Write(buf[:n]); err != nil {
			log.Println("error writing to outfile")
			return "", fmt.Errorf("error writing to outfile: %s", err.Error())
		}
	}
	log.Println("returning")
	return filepath.Join(config.LocalDownloadFolder, matches[0]), nil
}
