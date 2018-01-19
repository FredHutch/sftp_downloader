package main

import (
	"fmt"
	"log"
	"os"

	"github.com/kr/fs"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// Sftper helps make things testable
type Sftper interface {
	Close() error
	Create(path string) (*sftp.File, error)
	Lstat(p string) (os.FileInfo, error)
	Walk(root string) *fs.Walker
}

// Walker helps make things testable
type Walker interface {
	Step() bool
	Err() error
	Path() string
}

var sftpclient Sftper
var w Walker

func doTheWork(sftpclient Sftper) {
	defer sftpclient.Close()

	// walk a directory
	w = sftpclient.Walk("/tmp/")
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
	sftpclient, err = sftp.NewClient(conn)
	if err != nil {
		log.Fatal(err)
	}
	doTheWork(sftpclient)
}
