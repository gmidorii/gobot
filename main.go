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

const layout = "2006/01/02"

type Release struct {
	Date string
	Day  string
}

type Config struct {
	Token string
}

func main() {
	f, err := ioutil.ReadFile("config.toml")
	if err != nil {
		log.Fatalln(err)
	}
	var config Config
	if _, err := toml.Decode(string(f), &config); err != nil {
		log.Fatalln(err)
	}
	api := slack.New(config.Token)
	os.Exit(run(api))
}

func run(api *slack.Client) int {
	rtm := api.NewRTM()
	go rtm.ManageConnection()

	for {
		select {
		case msg := <-rtm.IncomingEvents:
			switch ev := msg.Data.(type) {
			case *slack.MessageEvent:
				if strings.Contains(ev.Text, "mocha-release") {
					post, err := post()
					if err != nil {
						log.Println(err)
					}
					rtm.SendMessage(rtm.NewOutgoingMessage(post, ev.Channel))
				}

				if strings.Contains(ev.Text, "mocha-change-day") {
					args := strings.Split(ev.Text, " ")
					if len(args) != 2 {
						rtm.SendMessage(rtm.NewOutgoingMessage("mocha-change-day [date:yyyy/MM/dd]", ev.Channel))
						continue
					}
					date := args[1]
					t, err := time.Parse(layout, date)
					if err != nil {
						rtm.SendMessage(rtm.NewOutgoingMessage("日付の形式まちがいです yyyy/MM/dd", ev.Channel))
						continue
					}
					if isHoliday(t) {
						rtm.SendMessage(rtm.NewOutgoingMessage("休日の指定はなしで..:darkness:", ev.Channel))
						continue
					}
					count, err := calcBusinessDay(t)
					if err != nil {
						rtm.SendMessage(rtm.NewOutgoingMessage(err.Error(), ev.Channel))
					}

					release := Release{
						Date: t.Format(layout),
						Day:  strconv.Itoa(count),
					}
					file, err := os.Create("args.toml")
					if err != nil {
						continue
					}
					w := bufio.NewWriter(file)
					encoder := toml.NewEncoder(w)
					encoder.Encode(release)

					p, _ := post()
					rtm.SendMessage(rtm.NewOutgoingMessage(p, ev.Channel))
				}
			case *slack.InvalidAuthEvent:
				log.Println("Error")
				return 1
			}
		}
	}
}

func post() (string, error) {
	args, err := ioutil.ReadFile("args.toml")
	if err != nil {
		return "", err
	}
	var release Release
	if _, err := toml.Decode(string(args), &release); err != nil {
		return "", err
	}

	t, err := template.ParseFiles("template.txt")
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

func calcBusinessDay(t time.Time) (int, error) {
	now := time.Now()
	count := 0
	if t.Before(now) {
		return count, errors.New("arg time must be after now")
	}
	for t.Format(layout) != now.Format(layout) {
		if !isHoliday(now) {
			count += 1
		}
		now = now.AddDate(0, 0, 1)
	}

	return count, nil
}
