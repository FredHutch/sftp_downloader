package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/FredHutch/sftp_downloader/iface"
)

func getMonthFirstDate(dayFirstDate string) string {
	segs := strings.Split(dayFirstDate, "-")
	return fmt.Sprintf("%s-%s-%s", segs[0], segs[2], segs[1])
}

func getFileNameToDownload(fileDate string, config Config, sftpclient iface.Sftper, phase Phase) (string, error) {
	var filePattern string
	if phase == ClinicalPhase {
		filePattern = fmt.Sprintf("Reportes_diarios_acumulados-%s", fileDate)
	} else if phase == LabPhase {
		filePattern = fmt.Sprintf("Reportes_diarios_acumulados_laboratorio-%s", fileDate)
	} else if phase == TNTPhase {
		// yyyy-mm-dd
		filePattern = fmt.Sprintf("Reportes_diarios_acumulados_TNT-%s", fileDate)
	}
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
		return "", fmt.Errorf("found more than one file matching pattern '%s'", filePattern)
	}
	return fmt.Sprintf("%s%s", remoteDir, matches[0]), nil
}

func getDownloadFolder(phase Phase, config Config) string {
	if phase == ClinicalPhase {
		return config.LocalDownloadFolderClinical
	} else if phase == LabPhase {
		return config.LocalDownloadFolderLab
	} else if phase == TNTPhase {
		return config.LocalDownloadFolderTNT
	}
	fmt.Println("Invalid phase:", phase)
	os.Exit(1)
	return ""
}

func doDownload(remoteFile string, config Config, sftpclient iface.Sftper, phase Phase) (rarFile string, retErr error) {
	f, err := sftpclient.Open(remoteFile)
	if err != nil {
		return "", fmt.Errorf("Could not open remote file %s: %s", remoteFile, err.Error())
	}
	downloadFolder := getDownloadFolder(phase, config)
	// check that destination directory exists
	exists, err := FileExists(downloadFolder)
	if err != nil {
		return "", fmt.Errorf("error checking if local download folder exists: %s", err.Error())
	}
	if !exists {
		return "", fmt.Errorf("local download directory '%s' does not exist",
			downloadFolder)
	}

	isdir, err := IsDir(downloadFolder)
	if err != nil {
		return "", fmt.Errorf("Error checking if local download dir is a dir: %s", err.Error())
	}

	if !isdir {
		return "", fmt.Errorf("'%s' exists but is not a directory", downloadFolder)
	}

	destFile := filepath.Join(downloadFolder, filepath.Base(remoteFile))
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
		n, err := f.Read(buf)
		if err != nil && err != io.EOF {
			return "", fmt.Errorf("premature EOF on input file: %s", err.Error())
		}
		if n == 0 {
			break
		}

		// write a chunk
		if _, err := outfile.Write(buf[:n]); err != nil {
			return "", fmt.Errorf("error writing to outfile: %s", err.Error())
		}
	}
	return filepath.Join(downloadFolder, filepath.Base(remoteFile)), nil
}

func downloadFile(fileDate string, config Config, sftpclient iface.Sftper, phase Phase) (rarFile string, retErr error) {
	remoteFile, err := getFileNameToDownload(fileDate, config, sftpclient, phase)
	if err != nil {
		return "", err
	}
	rarFile, err = doDownload(remoteFile, config, sftpclient, phase)
	if err != nil {
		return "", err
	}
	return rarFile, nil
}
