package main

import (
	"io/ioutil"
	"testing"

	"time"

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

func TestCalcBusinessDay_BeforeTime(t *testing.T) {
	past, _ := time.Parse("2006/01/02", "2000/04/03")
	_, err := calcBusinessDay(past)
	if err.Error() != "arg time must be after now" {
		t.Error("Expected error is not output")
	}
}

func TestCalcBusinessDay(t *testing.T) {
	now, _ = time.Parse("2006/01/02", "2017/04/03")
	date, _ := time.Parse("2006/01/02", "2017/04/28")
	count, err := calcBusinessDay(date)
	if err != nil {
		t.Error(err)
	}
	if count != 19 {
		t.Error("Count business day failed")
		t.Log(count)
	}
}

func TestIsHoliday(t *testing.T) {
	holiday, _ := time.Parse("2006/01/02", "2017/04/02")
	if isHoliday(holiday) != true {
		t.Error("Holiday judge is failed")
	}
	businessday, _ := time.Parse("2006/01/02", "2017/04/03")
	if isHoliday(businessday) != false {
		t.Error("Businessdaay judge is failed")
	}
}

func TestReadRelease(t *testing.T) {
	release, err := readRelease("test/args.toml")
	if err != nil {
		t.Error(err)
	}
	if release.Date != "2017/04/03" || release.Day != "10" {
		t.Error("Not Expected Relase")
		t.Log("Config.Date: " + release.Date)
		t.Log("Config.Day: " + release.Day)
	}
}
