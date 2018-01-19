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
var currentTimeFunction = time.Now // use currentTimeFunction instead of time.Now() throughout...

// SftpWrapper is a thin wrapper around sftp, implements Sftper
type SftpWrapper struct {
	cl *sftp.Client
}

// Config is a representation of the JSON config file
type Config struct {
	Host                  string `json:"host"`
	Port                  int    `json:"port"`
	User                  string `json:"user"`
	Password              string `json:"password"`
	LocalDownloadFolder   string `json:"local_download_folder"`
	RarDecryptionPassword string `json:"rar_decryption_password"`
}

// Walk method of concrete implementation, can't test these...
func (w *SftpWrapper) Walk(root string) iface.Walker {
	return w.cl.Walk(root)
}

// Create method of concrete implementation
func (w *SftpWrapper) Create(path string) (iface.Filer, error) {
	return w.cl.Create(path)
}

// Lstat method of concrete implementation
func (w *SftpWrapper) Lstat(p string) (os.FileInfo, error) {
	return w.cl.Lstat(p)
}

// Close method of concrete implementation
func (w *SftpWrapper) Close() error {
	return w.cl.Close()
}

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

func getDateString() string {
	var fileDate string
	if len(os.Args) == 3 {
		t, err := time.Parse("2006-01-02", os.Args[2])
		if err != nil {
			exit(1, fmt.Sprintf("Error formatting date %s, must be in YYYY-MM-DD format.", os.Args[2]))
		}
		fileDate = t.Format("02-01-2006")
	} else {
		now := currentTimeFunction()
		yesterday := now.AddDate(0, 0, -1)
		fileDate = yesterday.Format("02-01-2006")
	}
	return fileDate
}

func exit(exitCode int, msg string) int {
	if _, ok := os.LookupEnv("TESTING_SFTP_DOWNLOADER"); ok {
		os.Setenv("SFTP_DOWNLOADER_EXIT_CODE", strconv.Itoa(exitCode))
	} else { // These four lines can't be covered by
		fmt.Println(msg)  // tests
		os.Exit(exitCode) // without a lot of
	} // hassle.
	return exitCode
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
		os.Exit(1)
	}

	fileDate := getDateString()
	fmt.Println("fileDate is", fileDate)

	config := &ssh.ClientConfig{
		User: "dtenenba",
		Auth: []ssh.AuthMethod{
			ssh.Password(os.Args[1]),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	conn, err0 := ssh.Dial("tcp", "localhost:22", config)
	if err0 != nil {
		log.Fatal("Failed to dial: ", err0)
	}

	// open an SFTP session over an existing ssh connection.
	var err error
	client, err := sftp.NewClient(conn)
	if err != nil {
		log.Fatal(err)
	}
	sftpclient := &SftpWrapper{client}
	doTheWork(sftpclient)
}
