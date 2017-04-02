package main

import (
	"log"
	"os"
	"text/template"

	"io/ioutil"

	"github.com/BurntSushi/toml"
)

type Release struct {
	Date string
	Day  string
}

func main() {
	args, err := ioutil.ReadFile("args.toml")
	if err != nil {
		log.Fatalln(err)
	}
	var release Release
	if _, err := toml.Decode(string(args), &release); err != nil {
		log.Fatalln(err)
	}

	t, _ := template.ParseFiles("template.txt")
	t.Execute(os.Stdout, release)
}
