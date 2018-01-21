package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
)

// Config is a representation of the JSON config file
type Config struct {
	Host                  string `json:"host"`
	Port                  int    `json:"port"`
	User                  string `json:"user"`
	Password              string `json:"password"`
	LocalDownloadFolder   string `json:"local_download_folder"`
	RarDecryptionPassword string `json:"rar_decryption_password"`
}

// GetConfig populates the Config struct from a json file
func GetConfig(jsonFileName string) (Config, error) {
	// TODO check if jsonFileName's permissions are too open
	// (should not be readable by group or other)
	file, e := ioutil.ReadFile(jsonFileName)
	if e != nil {
		return Config{}, fmt.Errorf("Unable to open file %s.", jsonFileName)
	}
	var config Config
	err := json.Unmarshal(file, &config)
	if err != nil {
		return config, errors.New("Could not unmarshal JSON file.")
	}
	return config, nil
}
