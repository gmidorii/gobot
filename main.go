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
					sendSlack(rtm, ev.Channel, post)
				}

				if strings.Contains(ev.Text, "mocha-change-day") {
					args := strings.Split(ev.Text, " ")
					if len(args) != 2 {
						sendSlack(rtm, ev.Channel, "mocha-change-day [date:yyyy/MM/dd]")
						continue
					}
					date := args[1]
					t, err := time.Parse(layout, date)
					if err != nil {
						sendSlack(rtm, ev.Channel, "日付の形式まちがいです yyyy/MM/dd")
						continue
					}
					if isHoliday(t) {
						sendSlack(rtm, ev.Channel, "休日の指定はなしで..:darkness:")
						continue
					}
					count, err := calcBusinessDay(t)
					if err != nil {
						sendSlack(rtm, ev.Channel, err.Error())
					}

					release := Release{
						Date: t.Format(layout),
						Day:  strconv.Itoa(count),
					}
					update(release)

					p, _ := post()
					sendSlack(rtm, ev.Channel, p)
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

	time, err := time.Parse(layout, release.Date)
	if err != nil {
		return "", err
	}
	count, err := calcBusinessDay(time)
	if err != nil {
		return "", err
	}
	release.Day = strconv.Itoa(count)

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

func update(release Release) error {
	file, err := os.Create("args.toml")
	if err != nil {
		return err
	}
	w := bufio.NewWriter(file)
	encoder := toml.NewEncoder(w)
	return encoder.Encode(release)
}

func sendSlack(rtm *slack.RTM, ch, post string) {
	rtm.SendMessage(rtm.NewOutgoingMessage(post, ch))
}
