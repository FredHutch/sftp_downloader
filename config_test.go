package main

import (
	"testing"
)

func TestGetConfig(t *testing.T) {

	t.Run("validconfig", func(t *testing.T) {
		config, err := GetConfig("testdata/testconfig.json")
		if err != nil {
			t.Error("Did not expect error in getConfig()")
		}
		if config.Host != "somehost" {
			t.Error("Expected config.Host to be somehost, got ", config.Host)
		}

	})

	t.Run("badfile", func(t *testing.T) {
		_, err := GetConfig("testdata/nonexistentfile.json")
		if err == nil {
			t.Error("Expected error with getConfig() and nonexistent json file")
		}
	})

	t.Run("invalidjson", func(t *testing.T) {
		_, err := GetConfig("testdata/invalidjson.json")
		if err == nil {
			t.Error("Expected error in getConfig with invalid json file")
		}
	})

}
