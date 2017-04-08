package main

import (
	"log"
	"os"
	"text/template"

	"io/ioutil"

	"bytes"

	"strings"

	"time"

	"bufio"
	"strconv"

	"errors"

	"github.com/BurntSushi/toml"
	"github.com/nlopes/slack"
)

const (
	layout       = "2006/01/02"
	changeDate   = "change-date"
	argsFile     = "resources/args.toml"
	configFile   = "resources/config.toml"
	templateFile = "resources/template.txt"
)

type Release struct {
	Date string
	Day  string
}

type Config struct {
	Token    string
	UserHash string
}

func main() {
	config, err := readConfig(configFile)
	if err != nil {
		log.Fatalln(err)
	}
	api := slack.New(config.Token)
	os.Exit(run(api, config.UserHash))
}

func run(api *slack.Client, user string) int {
	rtm := api.NewRTM()
	go rtm.ManageConnection()
	log.Println("gobot start up")

	for {
		select {
		case msg := <-rtm.IncomingEvents:
			log.Println(msg.Type)
			log.Println(msg.Data)
			switch ev := msg.Data.(type) {
			case *slack.MessageEvent:
				if !strings.Contains(ev.Text, user) {
					continue
				}

				if c := strings.Index(ev.Text, changeDate); c != -1 {
					args := strings.Split(ev.Text[c:], " ")
					if len(args) != 2 {
						sendSlack(rtm, ev.Channel, changeDate+" [date:yyyy/MM/dd]")
						continue
					}
					date := args[1]
					log.Println(date)
					t, err := time.Parse(layout, date)
					if err != nil {
						sendSlack(rtm, ev.Channel, "日付の形式まちがいです yyyy/MM/dd")
						continue
					}
					if isHoliday(t) {
						sendSlack(rtm, ev.Channel, "休日の指定はなしで..:darkness:")
						continue
					}
					count, err := calcBusinessDay(t, time.Now())
					if err != nil {
						sendSlack(rtm, ev.Channel, err.Error())
						continue
					}

					release := Release{
						Date: t.Format(layout),
						Day:  strconv.Itoa(count),
					}
					err = update(release, argsFile)
					if err != nil {
						sendSlack(rtm, ev.Channel, "update failed")
						log.Println(err)
						continue
					}
				}
				p, err := createText(argsFile, templateFile)
				if err != nil {
					sendSlack(rtm, ev.Channel, "createText data failed")
				}
				sendSlack(rtm, ev.Channel, p)
			case *slack.InvalidAuthEvent:
				log.Println("Error")
				return 1
			}
		}
	}
}

func createText(args, temp string) (string, error) {
	release, err := readRelease(args)
	if err != nil {
		return "", err
	}

	date, err := time.Parse(layout, release.Date)
	if err != nil {
		return "", err
	}
	count, err := calcBusinessDay(date, time.Now())
	if err != nil {
		return "", err
	}
	release.Day = strconv.Itoa(count)

	t, err := template.ParseFiles(temp)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	t.Execute(&buf, release)
	return buf.String(), nil
}

func isHoliday(t time.Time) bool {
	if t.Weekday() == time.Sunday || t.Weekday() == time.Saturday {
		return true
	}
	return false
}

func calcBusinessDay(t time.Time, now time.Time) (int, error) {
	if t.Before(now) {
		return 0, errors.New("arg time must be after now")
	}
	count := 0
	for t.Format(layout) != now.Format(layout) {
		if !isHoliday(now) {
			count += 1
		}
		now = now.AddDate(0, 0, 1)
	}

	return count, nil
}

func update(release Release, file string) error {
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	w := bufio.NewWriter(f)
	encoder := toml.NewEncoder(w)
	return encoder.Encode(release)
}

func sendSlack(rtm *slack.RTM, ch, post string) {
	rtm.SendMessage(rtm.NewOutgoingMessage(post, ch))
}

func readConfig(file string) (Config, error) {
	f, err := ioutil.ReadFile(file)
	if err != nil {
		return Config{}, err
	}
	var config Config
	if _, err := toml.Decode(string(f), &config); err != nil {
		return Config{}, err
	}
	return config, nil
}

func readRelease(file string) (Release, error) {
	args, err := ioutil.ReadFile(file)
	if err != nil {
		return Release{}, err
	}
	var release Release
	if _, err := toml.Decode(string(args), &release); err != nil {
		return Release{}, err
	}
	return release, nil
}
