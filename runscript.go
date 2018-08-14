package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

func getScriptName(config Config, phase Phase) string {
	if phase == ClinicalPhase {
		return config.PostProcessingCommandClinical
	} else if phase == LabPhase {
		return config.PostProcessingCommandLab
	}
	return config.PostProcessingCommandTnt
}

func runScript(cmdline string, rundir string) (int, string, error) {
	currentDir, _ := os.Getwd()
	defer os.Chdir(currentDir)
	err := os.Chdir(rundir)
	if err != nil {
		return 1, "", fmt.Errorf("Could not change to script directory %s: %s",
			rundir, err.Error())
	}
	// wrap command in bash -c so that environment variables can be used
	cmd := exec.Command("bash", "-c", fmt.Sprintf("%s", cmdline))
	stdoutStderr, err := cmd.CombinedOutput()
	outputStr := string(stdoutStderr)
	ws := cmd.ProcessState.Sys().(syscall.WaitStatus)
	exitCode := ws.ExitStatus()
	fmt.Println("Command output:\n=======")
	fmt.Printf("%s\n", strings.TrimSpace(outputStr))
	fmt.Println("=======")

	if exitCode == 0 {
		fmt.Println("Post-processing script succeeded with exit code 0")
	} else {
		fmt.Printf("Post-processing script failed! Exit code: %d\n", exitCode)
	}
	if err != nil {
		return exitCode, outputStr, errors.New("error getting combined output of command")
	}
	return exitCode, outputStr, nil
}
