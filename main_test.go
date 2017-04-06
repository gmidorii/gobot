package main

import "testing"

func TestReadConfig(t *testing.T) {
	config, err := readConfig("test/config.toml")
	if err != nil {
		t.Error(err)
	}
	if config.Token != "test-token" || config.UserHash != "<@test>" {
		t.Error("Not Expected Config")
		t.Log("Config.Token: " + config.Token)
		t.Log("Config.UserHash: " + config.UserHash)
	}
}
