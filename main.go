package main

import (
	"fmt"
	"log"
	"os"

	"github.com/FredHutch/sftp_downloader/iface"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// SftpWrapper is a thin wrapper around sftp, implements Sftper
type SftpWrapper struct {
	cl *sftp.Client
}

// Walk method of concrete implementation
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

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("usage: %s [password]\n", os.Args[0])
		os.Exit(1)
	}
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
