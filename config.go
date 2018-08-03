package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
)

// Config is a representation of the JSON config file
type Config struct {
	Host                          string `json:"host"`
	Port                          int    `json:"port"`
	User                          string `json:"user"`
	Password                      string `json:"password"`
	LocalDownloadFolderClinical   string `json:"local_download_folder_clinical"`
	LocalDownloadFolderLab        string `json:"local_download_folder_lab"`
	LocalDownloadFolderTNT        string `json:"local_download_folder_tnt"`
	RarDecryptionPassword         string `json:"rar_decryption_password"`
	PostProcessingCommandClinical string `json:"postprocessing_command_clinical"`
	PostProcessingCommandLab      string `json:"postprocessing_command_lab"`
	PhiZipPassword                string `json:"phi_zip_password"`
}

// Phase represents which phase of downloading we are doing (Clinical/Lab/TNT)
type Phase int

const (
	// ClinicalPhase is bla
	ClinicalPhase Phase = iota
	// TNTPhase is bla
	TNTPhase
	// LabPhase is bla
	LabPhase
)

var (
	// FIXME TODO
	// phases = []Phase{ClinicalPhase, LabPhase, TNTPhase}
	phases = []Phase{TNTPhase}
)

func getPhaseName(phase Phase) string {
	switch phase {
	case ClinicalPhase:
		return "clinical"
	case LabPhase:
		return "lab"
	}
	return "tnt"
}

// GetConfig populates the Config struct from a json file
func GetConfig(jsonFileName string) (Config, error) {
	// TODO check if jsonFileName's permissions are too open
	// (should not be readable by group or other)
	file, e := ioutil.ReadFile(jsonFileName)
	if e != nil {
		return Config{}, fmt.Errorf("unable to open file %s", jsonFileName)
	}
	var config Config
	err := json.Unmarshal(file, &config)
	if err != nil {
		return config, errors.New("could not unmarshal JSON file")
	}
	return config, nil
}
