package main

import (
	"encoding/json"
	"fmt"

	"github.com/jason0x43/go-alfred"
	"github.com/pkg/browser"
)

// UsersCommand allows an API token to be specified
type UsersCommand struct{}

// About returns information about a command
func (c UsersCommand) About() alfred.CommandDef {
	return alfred.CommandDef{
		Keyword:     "users",
		Description: "List users",
		IsEnabled:   config.APIToken != "",
		Arg: &alfred.ItemArg{
			Keyword: "users",
		},
	}
}

// Items returns the items for the command
func (c UsersCommand) Items(arg, data string) (items []alfred.Item, err error) {
	if err = checkRefresh(); err != nil {
		return
	}

	var cfg userConfig
	if data != "" {
		if err = json.Unmarshal([]byte(data), &cfg); err != nil {
			return
		}
	}

	var channel *Channel
	if cfg.Channel != nil {
		if c, found := getChannel(*cfg.Channel); found {
			channel = &c
		}
	}

	for _, user := range cache.Users {
		if user.Deleted {
			continue
		}

		if !cfg.ShowAll && user.Presence == "away" {
			continue
		}

		if channel != nil && !isInChannel(user.ID, channel) {
			continue
		}

		if alfred.FuzzyMatches(user.Name, arg) || alfred.FuzzyMatches(user.Profile.RealName, arg) {
			item := alfred.Item{
				UID:          user.ID,
				Title:        user.Name,
				Subtitle:     fmt.Sprintf("%s %s, %s", user.Profile.FirstName, user.Profile.LastName, user.Profile.Email),
				Autocomplete: user.Name,
				Arg: &alfred.ItemArg{
					Keyword: "users",
					Mode:    alfred.ModeDo,
					Data: alfred.Stringify(&userConfig{
						ToMessage: &dmID{
							User: user.ID,
							Team: cache.Auth.TeamID,
						},
					}),
				},
			}

			if user.Presence == "away" {
				item.Icon = "icon_faded.png"
			}

			item.AddMod(alfred.ModCmd, alfred.ItemMod{
				Subtitle: "Open profile",
				Arg: &alfred.ItemArg{
					Keyword: "users",
					Mode:    alfred.ModeDo,
					Data: alfred.Stringify(&userConfig{
						ToOpen: &dmID{
							User: user.ID,
							Team: cache.Auth.TeamID,
						},
					}),
				},
			})

			items = append(items, item)
		}
	}

	alfred.FuzzySort(items, arg)

	return
}

// Do implements the command
func (c UsersCommand) Do(data string) (out string, err error) {
	var cfg userConfig
	if data != "" {
		if err := json.Unmarshal([]byte(data), &cfg); err != nil {
			return "", fmt.Errorf("Invalid open message")
		}
	}

	if cfg.ToMessage != nil {
		var channel string
		s := OpenSession(config.APIToken)
		if channel, err = s.OpenDirectMessage(cfg.ToMessage.User); err != nil {
			return
		}
		browser.OpenURL(fmt.Sprintf("slack://channel?team=%s&id=%s", cfg.ToMessage.Team, channel))
	}

	if cfg.ToOpen != nil {
		browser.OpenURL(fmt.Sprintf("slack://user?team=%s&id=%s", cfg.ToOpen.Team, cfg.ToOpen.User))
	}

	return
}

func isInChannel(userID string, channel *Channel) bool {
	for _, uid := range channel.Members {
		if uid == userID {
			return true
		}
	}
	return false
}

type dmID struct {
	User string
	Team string
}

type userConfig struct {
	ToMessage *dmID
	ToOpen    *dmID
	Channel   *string
	ShowAll   bool
}
