package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func TestRunScript(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "sftpdownloader-testing")
	if err != nil {
		t.Fail()
	}
	defer os.RemoveAll(tempDir)
	pwd, err := os.Getwd()
	if err != nil {
		t.Fail()
	}
	t.Run("exit0", func(t *testing.T) {
		// func runScript(cmdline string, rundir string) (int, error)
		cmdLine := fmt.Sprintf("%s/testdata/testprint.sh", pwd)
		exitCode, output, err := runScript(cmdLine, tempDir)
		if err != nil {
			t.Error("unexpected error")
		}
		if exitCode != 0 {
			t.Errorf("Got exitcode %d, expected 0", exitCode)
		}
		exp := "line 1 to stderr\nline 2 to stdout"
		if !strings.Contains(output, exp) {
			t.Errorf("expected output to contain:\n%s", exp)
		}
		newPwd, err := os.Getwd()
		if err != nil {
			t.Fail()
		}
		if pwd != newPwd {
			t.Errorf("Directory has changed from %s to %s", pwd, newPwd)
		}
	})

	t.Run("exit1", func(t *testing.T) {
		cmdLine := fmt.Sprintf("%s/testdata/exit1.sh", pwd)
		exitCode, output, err := runScript(cmdLine, tempDir)

		if err == nil {
			t.Error("expected an error")
		}
		exp := "error getting combined output of command"
		if err.Error() != exp {
			t.Errorf("Expected error '%s', got %s", exp, err.Error())
		}
		if exitCode != 1 {
			t.Errorf("Expected exit code 1, got %d", exitCode)
		}
		exp = "error\nhaha\n"
		if output != exp {
			t.Errorf("Expected:\n%s\nGot:\n%s\n", exp, output)
		}
	})

	t.Run("bad_dir", func(t *testing.T) {
		_, _, err := runScript("echo bobcat", "/nonexistentdirectory")
		exp := "Could not change to script directory /nonexistentdirectory: chdir /nonexistentdirectory: no such file or directory"
		if err.Error() != exp {
			t.Errorf("Expected error:\n%s\nGot error:%s\n", exp, err.Error())
		}
	})

}
