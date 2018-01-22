package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
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
		fileDate = t.Format("02-01-2006")
	} else {
		now := currentTimeFunction()
		yesterday := now.AddDate(0, 0, -1)
		fileDate = yesterday.Format("02-01-2006")
	}
	return fileDate, nil
}

func exit(exitCode int, msg string) int { // TODO remove - only called from main() which is not tested (??)
	if _, ok := os.LookupEnv("TESTING_SFTP_DOWNLOADER"); ok {
		os.Setenv("SFTP_DOWNLOADER_EXIT_CODE", strconv.Itoa(exitCode))
		os.Setenv("SFTP_EXIT_MESSAGE", msg)
	} else { // These four lines can't be covered by
		fmt.Println(msg)  // tests
		os.Exit(exitCode) // without a lot of
	} // hassle.
	return exitCode
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
		fmt.Printf(`
usage: %s config-file [date-string]

config-file: the path to a JSON file containing configuration information
date-string [optional]: the date of a file to process, defaults to yesterday.
		format: YYYY-MM-DD (example: 2018-01-19)
See complete documentation at:
   https://github.com/FredHutch/sftp_downloader/tree/%s
`, os.Args[0], gitBranch)
		os.Exit(1) // TODO change to exit() ? No point if not testable?
	}

	fileDate, err := getDateString()
	if err != nil {
		exit(1, err.Error())
		return
	}

	config, err := GetConfig(os.Args[1])
	if err != nil {
		exit(1, err.Error())
		return
	}

	sshConfig := &ssh.ClientConfig{
		User: config.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(config.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	conn, err0 := ssh.Dial("tcp",
		fmt.Sprintf("%s:%d", config.Host, config.Port), sshConfig)
	if err0 != nil {
		exit(1, "Failed to connect to ssh server.")
		return
	}

	// open an SFTP session over an existing ssh connection.
	client, err := sftp.NewClient(conn)
	if err != nil {
		exit(1, "Could not create sftp client.")
		return
	}
	sftpclient := &SftpWrapper{Cl: client}
	rarFile, err := downloadFile(fileDate, config, sftpclient)
	if err != nil {
		exit(1, fmt.Sprintf("Could not download file: %s", err.Error()))
	}
	err = UncompressFile(rarFile, fileDate, config)
	if err != nil {
		exit(1, fmt.Sprintf("Could not unrar file: %s", err.Error()))
	}
	fmt.Println("Done:")
	fmt.Printf("Downloaded %s and unarchived files to %s/%s.\n", rarFile,
		config.LocalDownloadFolder, fileDate)
	// TODO run a custom script in the directory where files were archived
}
