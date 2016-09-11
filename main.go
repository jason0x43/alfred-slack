package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"time"

	"github.com/jason0x43/go-alfred"
)

var cacheFile string
var configFile string
var config configStruct
var cache cacheStruct
var workflow alfred.Workflow

type configStruct struct {
	APIToken string `json:"api_key"`
}

type cacheStruct struct {
	Time     time.Time
	Auth     Auth
	Channels []Channel
	Users    []User
}

var dlog = log.New(os.Stderr, "[redmine] ", log.LstdFlags)

func main() {
	var err error

	dlog.Printf("Args: %#v", os.Args)

	workflow, err = alfred.OpenWorkflow(".", true)
	if err != nil {
		fmt.Printf("Error: %s", err)
		os.Exit(1)
	}

	configFile = path.Join(workflow.DataDir(), "config.json")
	cacheFile = path.Join(workflow.CacheDir(), "cache.json")

	dlog.Println("Using config file", configFile)
	dlog.Println("Using cache file", cacheFile)

	err = alfred.LoadJSON(configFile, &config)
	if err != nil {
		dlog.Println("Error loading config:", err)
	}
	dlog.Println("loaded config:", config)

	err = alfred.LoadJSON(cacheFile, &cache)
	dlog.Println("loaded cache")

	commands := []alfred.Command{
		TokenCommand{},
		ChannelsCommand{},
		UsersCommand{},
		PresenceCommand{},
		ResetCommand{},
	}

	workflow.Run(commands)
}
