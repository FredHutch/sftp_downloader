package iface

import "os"

// run this with:
// go generate ./...

//go:generate mockgen -destination=../mocks/mainmock.go -package=mocks github.com/FredHutch/sftp_downloader/iface Sftper,Filer,Walker

//go:generate mockgen -destination=../mocks/fileinfo.go -package=mocks os FileInfo

// Sftper helps make things testable
type Sftper interface {
	// return interface
	Walk(root string) Walker

	// return interface
	Create(path string) (Filer, error)

	// return concrete type since no methods on FileInfo are used by doTheWork
	Lstat(p string) (os.FileInfo, error)

	Close() error
}

// Filer interface has methods used by doTheWork
type Filer interface {
	Write([]byte) (int, error)
}

// Walker helps make things testable
type Walker interface {
	Step() bool
	Err() error
	Path() string
}
