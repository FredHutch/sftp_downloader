package main

import (
	"io"
	"os"

	"github.com/FredHutch/sftp_downloader/iface"
	"github.com/pkg/sftp"
)

// Concrete implementations of interfaces, for use by the actual running
// program (not tests which mock out interfaces).

// SftpWrapper is a thin wrapper around sftp, implements Sftper
type SftpWrapper struct {
	Cl *sftp.Client
}

// Walk method of concrete implementation, can't test these...
func (w *SftpWrapper) Walk(root string) iface.Walker {
	return w.Cl.Walk(root)
}

// Create method of concrete implementation
func (w *SftpWrapper) Create(path string) (iface.Filer, error) {
	return w.Cl.Create(path)
}

// Lstat method of concrete implementation
func (w *SftpWrapper) Lstat(p string) (os.FileInfo, error) {
	return w.Cl.Lstat(p)
}

// Close method of concrete implementation
func (w *SftpWrapper) Close() error {
	return w.Cl.Close()
}

// ReadDir method of concrete implementation
func (w *SftpWrapper) ReadDir(p string) ([]os.FileInfo, error) {
	return w.Cl.ReadDir(p)
}

// Open method of concrete implementation
func (w *SftpWrapper) Open(path string) (io.Reader, error) {
	return w.Cl.Open(path)
}
