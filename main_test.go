package main

import (
	"io/ioutil"
	"testing"

	"github.com/BurntSushi/toml"
)

func TestReadConfig(t *testing.T) {
	config, err := readConfig("test/config-test.toml")
	if err != nil {
		t.Error(err)
	}
	if config.Token != "test-token" || config.UserHash != "<@test>" {
		t.Error("Not Expected Config")
		t.Log("Config.Token: " + config.Token)
		t.Log("Config.UserHash: " + config.UserHash)
	}
}

func TestUpdate(t *testing.T) {
	expected := Release{
		Date: "2017/04/03",
		Day:  "10",
	}
	update(expected, "test/args.toml")

	args, _ := ioutil.ReadFile("test/args.toml")
	var actual Release
	_, err := toml.Decode(string(args), &actual)
	if err != nil {
		t.Error(err)
	}
	if expected.Date != actual.Date || expected.Day != actual.Day {
		t.Error("Not Expected Release")
		t.Log("Release.Date: " + actual.Date)
		t.Log("Release.Day: " + actual.Day)
	}
}
