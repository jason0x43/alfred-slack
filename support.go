package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/jason0x43/go-alfred"
)

func checkRefresh() error {
	if time.Now().Sub(cache.Time).Minutes() < 5.0 {
		return nil
	}

	dlog.Println("Refreshing cache...")
	err := refresh()
	if err != nil {
		dlog.Println("Error refreshing cache:", err)
	}
	return err
}

func refresh() (err error) {
	s := OpenSession(config.APIToken)
	cache.Time = time.Now()

	dataChan := make(chan interface{})
	errorChan := make(chan error)

	go func() {
		if auth, err := s.GetAuth(); err != nil {
			errorChan <- err
		} else {
			dataChan <- auth
		}
	}()

	go func() {
		if channels, err := s.GetChannels(); err != nil {
			errorChan <- err
		} else {
			dataChan <- channels
		}
	}()

	go func() {
		if users, err := s.GetUsers(); err != nil {
			errorChan <- err
		} else {
			dataChan <- users
		}
	}()

	// wait for all functions to complete
	for i := 0; i < 3; i++ {
		select {
		case data := <-dataChan:
			switch value := data.(type) {
			case Auth:
				cache.Auth = value
				dlog.Println("Got auth")
			case []Channel:
				cache.Channels = value
				dlog.Println("Got channels")
			case []User:
				cache.Users = value
				dlog.Println("Got users")
			}
		case err := <-errorChan:
			return err
		}
	}

	for i := range cache.Users {
		go func(i int) {
			uid := cache.Users[i].ID
			up := userPresence{ID: uid}
			if presence, err := s.GetPresence(uid); err != nil {
				errorChan <- err
			} else {
				up.Presence = presence
				dataChan <- up
			}
		}(i)
	}

	for i := 0; i < len(cache.Users); i++ {
		select {
		case data := <-dataChan:
			switch value := data.(type) {
			case userPresence:
				ui := indexOfUserByID(value.ID)
				cache.Users[ui].Presence = value.Presence
				dlog.Printf("Got presence for %s", cache.Users[ui].Name)
			}
		case err := <-errorChan:
			return err
		}
	}

	return alfred.SaveJSON(cacheFile, &cache)
}

type userPresence struct {
	ID       string
	Presence string
}

func getChannel(id string) (c Channel, found bool) {
	for i := range cache.Channels {
		if cache.Channels[i].ID == id {
			return cache.Channels[i], true
		}
	}
	return
}

func indexOfUserByID(id string) (i int) {
	for i := range cache.Users {
		if cache.Users[i].ID == id {
			return i
		}
	}
	return -1
}

func getUser(id string) (u User, found bool) {
	i := indexOfUserByID(id)
	if i != -1 {
		return cache.Users[i], true
	}
	return
}

func getFile(url string, filename string) (outFile string, err error) {
	if filename == "" {
		filename = path.Base(url)
	}
	outFile = path.Join(workflow.CacheDir(), filename)
	if _, err = os.Stat(outFile); err == nil {
		return
	}

	var req *http.Request
	if req, err = http.NewRequest("GET", url, nil); err != nil {
		return
	}

	var resp *http.Response
	if resp, err = client.Do(req); err != nil {
		return
	}
	defer resp.Body.Close()

	var content []byte
	if content, err = ioutil.ReadAll(resp.Body); err != nil {
		return
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return "", fmt.Errorf(resp.Status)
	}

	err = ioutil.WriteFile(outFile, content, 0600)
	return
}
