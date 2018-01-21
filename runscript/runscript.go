package runscript

import (
	"fmt"
	"os"
)

func runScript(cmdline string, rundir string) (int, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return 0, fmt.Errorf("Could not change to script directory %s", rundir)
	}
	defer os.Chdir(currentDir)
	os.Chdir(rundir)
	return 0, nil
}
