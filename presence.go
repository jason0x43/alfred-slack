package main

import (
	"encoding/json"
	"fmt"

	"github.com/jason0x43/go-alfred"
)

// PresenceCommand resets all stored API tokens
type PresenceCommand struct{}

// About returns information about a command
func (c PresenceCommand) About() alfred.CommandDef {
	return alfred.CommandDef{
		Keyword:     "presence",
		Description: "Show and update your presence",
		IsEnabled:   config.APIToken != "",
	}
}

// Items returns a list of filter items
func (c PresenceCommand) Items(arg, data string) (items []alfred.Item, err error) {
	s := OpenSession(config.APIToken)

	var presence string
	if presence, err = s.GetPresence(cache.Auth.UserID); err != nil {
		return
	}

	var title string
	var subtitle string
	var newState Presence

	if presence == "away" {
		title = "Away"
		subtitle = "Press Enter to change to Active"
		newState = PresenceAuto
	} else {
		title = "Active"
		subtitle = "Press Enter to change to Away"
		newState = PresenceAway
	}

	item := alfred.Item{
		Title:    title,
		Subtitle: subtitle,
		Arg: &alfred.ItemArg{
			Keyword: "presence",
			Mode:    alfred.ModeDo,
			Data:    alfred.Stringify(presenceConfig{NewState: newState}),
		},
	}

	if presence == "away" {
		item.Icon = "icon_faded.png"
	}

	items = append(items, item)

	return
}

// Do implements the command
func (c PresenceCommand) Do(data string) (out string, err error) {
	var cfg presenceConfig
	if data != "" {
		if err := json.Unmarshal([]byte(data), &cfg); err != nil {
			return "", fmt.Errorf("Error unmarshalling data: %v", err)
		}
	}

	s := OpenSession(config.APIToken)
	if err = s.SetPresence(cfg.NewState); err != nil {
		return
	}

	i := indexOfUserByID(cache.Auth.UserID)
	if cfg.NewState == PresenceAuto {
		cache.Users[i].Presence = "active"
	} else {
		cache.Users[i].Presence = "away"
	}

	out = fmt.Sprintf("Presence set to %s", cfg.NewState)
	return
}

type presenceConfig struct {
	NewState Presence
}
