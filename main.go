package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/FredHutch/sftp_downloader/iface"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// values provided at compile time (see build.sh):
var gitBranch, gitCommit string

// sub replaceable by tests:
var currentTimeFunction = time.Now // use currentTimeFunction() instead of time.Now() throughout...

func doTheWork(sftpclient iface.Sftper) {
	defer sftpclient.Close()

	// walk a directory
	w := sftpclient.Walk("/tmp/")
	for w.Step() {
		if w.Err() != nil {
			continue
		}
		log.Println(w.Path())
	}

	// leave your mark
	f, err := sftpclient.Create("hello.txt")
	if err != nil {
		log.Fatal(err)
	}
	if _, err1 := f.Write([]byte("Hello world!")); err1 != nil {
		log.Fatal(err1)
	}

	// check it's there
	fi, err := sftpclient.Lstat("hello.txt")
	if err != nil {
		log.Fatal(err)
	}
	log.Println(fi)

}

func getDateString() (string, error) {
	var fileDate string
	if len(os.Args) == 3 {
		t, err := time.Parse("2006-01-02", os.Args[2])
		if err != nil {
			return "", fmt.Errorf("error formatting date %s, must be in YYYY-MM-DD format", os.Args[2])
		}
		fileDate = t.Format("2006-01-02")
	} else {
		now := currentTimeFunction()
		yesterday := now.AddDate(0, 0, -1)
		fileDate = yesterday.Format("2006-01-02")
	}
	return fileDate, nil
}

func filter(vs []os.FileInfo, f func(string) bool) []string {
	vsf := make([]string, 0)
	for _, v := range vs {
		if f(v.Name()) {
			vsf = append(vsf, v.Name())
		}
	}
	return vsf
}

func main() {
	if len(os.Args) < 2 {
		fmt.Printf(`sftp_downloader, commit %s

usage: %s config-file [date-string]

config-file: the path to a JSON file containing configuration information
date-string [optional]: the date of a file to process, defaults to yesterday.
		format: YYYY-MM-DD (example: 2018-01-19)
See complete documentation at:
   https://github.com/FredHutch/sftp_downloader/blob/%s/README.md
`, gitCommit, os.Args[0], gitBranch)
		os.Exit(1) // TODO change to exit() ? No point if not testable?
	}

	config, err := GetConfig(os.Args[1])
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	sshConfig := &ssh.ClientConfig{
		User: config.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(config.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	fmt.Println("Connecting to SFTP server...")
	conn, err0 := ssh.Dial("tcp",
		fmt.Sprintf("%s:%d", config.Host, config.Port), sshConfig)
	if err0 != nil {

		fmt.Println("Failed to connect to ssh server.")
		os.Exit(1)
	}

	// open an SFTP session over an existing ssh connection.
	client, err := sftp.NewClient(conn)
	if err != nil {
		fmt.Println("Could not create sftp client.")
		os.Exit(1)
	}
	sftpclient := &SftpWrapper{Cl: client}

	var exitCodes []int

	for _, phase := range phases {

		fileDate, err := getDateString()
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
			return
		}

		downloadFolder := getDownloadFolder(phase, config)
		fmt.Printf("Downloading %s file...\n", getPhaseName(phase))

		rarFile, err := downloadFile(fileDate, config, sftpclient, phase)
		if err != nil {
			fmt.Printf("Could not download file: %s\n", err.Error())
			if phase == ClinicalPhase {
				os.Exit(1)
			} else {
				os.Exit(0)
			}
		}

		fmt.Printf("Unarchiving %s file...\n", getPhaseName(phase))
		err = UncompressFile(rarFile, fileDate, config, phase)
		if err != nil {
			fmt.Printf("Could not unrar file: %s\n", err.Error())
			os.Exit(1)
		}

		rundir := filepath.Join(downloadFolder, fileDate)

		if phase == ClinicalPhase {
			fmt.Println("Moving CSV and SAV files to top level...")
			err = moveFiles(rundir) // TODO needs to behave differently depending on phase??
			if err != nil {
				fmt.Printf("Error renaming files: %s.\n", err.Error())
				os.Exit(1)
			}
		} else if phase == LabPhase {
			fmt.Println("Consolidating lab files...")
			rawLabFileDir := filepath.Join(getDownloadFolder(LabPhase, config), fileDate)
			err = processLabFiles(config, rawLabFileDir)
			if err != nil {
				fmt.Println("Error processing lab files:", err.Error())
				os.Exit(1)
			}

		} else if phase == TNTPhase {
			fmt.Println("need to do more stuff here after extracting tnt files")
			/*
				TODOS:
				rename e.g.
				TNTstudies-Enrolamiento-20180719223543.csv
				to
				enr.TNT
				waiting for clarification on exactly what needs to be done
			*/

		}

		// if phase == ClinicalPhase {
		// 	fmt.Println("Renaming download directory...")
		// 	fileDate, err = renameDownloadDir(config, fileDate, phase)
		// 	if err != nil {
		// 		fmt.Printf("Error renaming download directory: %s\n", err.Error())
		// 		os.Exit(1)
		// 	}
		// }

		/*
			lab files (after processing) do not currently end up in a folder
			that has a date name. perhaps they should. this seems to cause
			the lab post-processing command to fail.
		*/

		rundir = filepath.Join(downloadFolder, fileDate)
		fmt.Printf("Running %s postprocessing command...\n", getPhaseName(phase))
		exitCode, _, _ := runScript(getScriptName(config, phase), rundir)
		exitCodes = append(exitCodes, exitCode)
		fmt.Printf("%s postprocessing command exited with code %d.\n", getPhaseName(phase), exitCode)

		fmt.Println("Done:") // TODO remove?
		fmt.Printf("(%s) Downloaded %s and unarchived files to %s/%s.\n", getPhaseName(phase), rarFile,
			downloadFolder, fileDate)

	} // end of for phase in phases

	var overallExitCode int
	for _, code := range exitCodes {
		if code > 0 {
			overallExitCode = code
		}
	}

	fmt.Printf("Exiting with exit code %d.\n", overallExitCode)
	os.Exit(overallExitCode)
}
