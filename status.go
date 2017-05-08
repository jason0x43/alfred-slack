package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jason0x43/go-alfred"
)

// StatusCommand resets all stored API tokens
type StatusCommand struct{}

// About returns information about a command
func (c StatusCommand) About() alfred.CommandDef {
	return alfred.CommandDef{
		Keyword:     "status",
		Description: "Show and update your status",
		IsEnabled:   config.APIToken != "",
	}
}

// Items returns a list of filter items
func (c StatusCommand) Items(arg, data string) (items []alfred.Item, err error) {
	var cfg statusConfig
	if data != "" {
		if err := json.Unmarshal([]byte(data), &cfg); err != nil {
			dlog.Printf("Invalid channel config")
		}
	}

	var emoji string
	var emojiName string

	if cfg.StatusEmoji != nil {
		emoji = *cfg.StatusEmoji
		emojiName = strings.TrimSuffix(strings.TrimPrefix(emoji, ":"), ":")
	}

	if cfg.StatusText != nil {
		for i := range cache.Emoji {
			name := cache.Emoji[i].Name

			if name == emojiName {
				continue
			}

			if alfred.FuzzyMatches(name, arg) {
				ename := ":" + name + ":"

				item := alfred.Item{
					Title: name,
					Arg: &alfred.ItemArg{
						Keyword: "status",
						Mode:    alfred.ModeDo,
						Data:    alfred.Stringify(statusConfig{NewState: cfg.NewState, StatusText: cfg.StatusText, StatusEmoji: &ename}),
					},
				}

				if emojiFile, err := getEmojiFromSlack(name); err == nil {
					item.Icon = emojiFile
				}

				items = append(items, item)
			}
		}

		if spriteEmoji, err := getAllSpriteEmoji(); err == nil {
			for _, name := range spriteEmoji {
				if name == emojiName {
					continue
				}

				if alfred.FuzzyMatches(name, arg) {
					ename := ":" + name + ":"

					item := alfred.Item{
						Title: name,
						Arg: &alfred.ItemArg{
							Keyword: "status",
							Mode:    alfred.ModeDo,
							Data:    alfred.Stringify(statusConfig{NewState: cfg.NewState, StatusText: cfg.StatusText, StatusEmoji: &ename}),
						},
					}

					if emojiFile, err := getEmojiFromSprite(name); err == nil {
						item.Icon = emojiFile
					}

					items = append(items, item)
				}
			}
		} else {
			dlog.Print("Unable to load sprite icons: ", err)
		}

		alfred.FuzzySort(items, arg)

		if arg == "" {
			emoji := ""
			item := alfred.Item{
				Title: "No status icon",
				Icon:  "none",
				Arg: &alfred.ItemArg{
					Keyword: "status",
					Mode:    alfred.ModeDo,
					Data:    alfred.Stringify(statusConfig{NewState: cfg.NewState, StatusText: cfg.StatusText, StatusEmoji: &emoji}),
				},
			}

			items = append([]alfred.Item{item}, items...)
		}

		dlog.Print("status emoji: ", emoji)
		if emoji != "" && alfred.FuzzyMatches(*cfg.StatusEmoji, arg) {
			item := alfred.Item{
				Title: emojiName,
				Arg: &alfred.ItemArg{
					Keyword: "status",
					Mode:    alfred.ModeDo,
					Data:    alfred.Stringify(statusConfig{NewState: cfg.NewState, StatusText: cfg.StatusText, StatusEmoji: &emoji}),
				},
			}

			if emojiFile, err := getEmojiFromSlack(emoji); err == nil {
				item.Icon = emojiFile
			}

			items = append([]alfred.Item{item}, items...)
		}
	} else {
		i := indexOfUserByID(cache.Auth.UserID)
		if i == -1 {
			err = fmt.Errorf("The user cache is empty")
			return
		}

		if time.Now().Sub(cache.PresenceTime).Minutes() > 1.0 || cache.Users[i].Presence == "" {
			s := OpenSession(config.APIToken)
			if cache.Users[i].Presence, err = s.GetPresence(cache.Auth.UserID); err != nil {
				return
			}
			cache.PresenceTime = time.Now()
			if err = alfred.SaveJSON(cacheFile, &cache); err != nil {
				return
			}
		}

		// There are two properties of interest, 'presence' and 'status'. Presence
		// is whether a user is active or away, and status is some message
		// indicating what they're doing.

		var title string
		var subtitle string

		user := cache.Users[i]
		presence := user.Presence

		if arg == "" {
			title = user.Profile.StatusText
			if user.Profile.StatusText == "" {
				subtitle = "No status message"
			} else {
				subtitle = "Clear existing status message"
			}
		} else {
			title = arg
			subtitle = "Update status message"
		}

		item := alfred.Item{
			Title:    title,
			Subtitle: subtitle,
			Arg: &alfred.ItemArg{
				Keyword: "status",
				Data:    alfred.Stringify(statusConfig{StatusText: &arg, StatusEmoji: &user.Profile.StatusEmoji}),
			},
		}

		if presence == PresenceAway {
			item.Title = fmt.Sprintf("%s %s", AwayMarker, item.Title)
		} else {
			item.Title = fmt.Sprintf("%s %s", ActiveMarker, item.Title)
		}

		var modSubtitle string
		if subtitle == "" {
			modSubtitle = "Set presence to Active"
		} else {
			modSubtitle = subtitle + ", set presence to Active"
		}

		item.AddMod(alfred.ModCmd, alfred.ItemMod{
			Subtitle: modSubtitle,
			Arg: &alfred.ItemArg{
				Keyword: "status",
				Data:    alfred.Stringify(statusConfig{NewState: PresenceActive, StatusText: &arg, StatusEmoji: &user.Profile.StatusEmoji}),
			},
		})

		if subtitle == "" {
			modSubtitle = "Set presence to Away"
		} else {
			modSubtitle = subtitle + ", set presence to Away"
		}

		item.AddMod(alfred.ModAlt, alfred.ItemMod{
			Subtitle: modSubtitle,
			Arg: &alfred.ItemArg{
				Keyword: "status",
				Data:    alfred.Stringify(statusConfig{NewState: PresenceAway, StatusText: &arg, StatusEmoji: &user.Profile.StatusEmoji}),
			},
		})

		if user.Profile.StatusEmoji != "" {
			var emojiFile string
			emojiFile, err = getEmojiFromSprite(user.Profile.StatusEmoji)
			if err != nil {
				emojiFile, err = getEmojiFromSlack(user.Profile.StatusEmoji)
			}

			if err == nil {
				item.Icon = emojiFile
			}
		}

		items = append(items, item)
	}

	return
}

// Do implements the command
func (c StatusCommand) Do(data string) (out string, err error) {
	var cfg statusConfig
	if data != "" {
		if err := json.Unmarshal([]byte(data), &cfg); err != nil {
			return "", fmt.Errorf("Error unmarshalling data: %v", err)
		}
	}

	s := OpenSession(config.APIToken)
	var errPresence error
	var errStatus error

	if cfg.NewState != "" {
		if errPresence = s.SetPresence(cfg.NewState); errPresence == nil {
			i := indexOfUserByID(cache.Auth.UserID)
			if i == -1 {
				dlog.Printf("The user cache is empty")
				errPresence = fmt.Errorf("The user cache is empty")
			} else {
				if cfg.NewState == PresenceActive {
					cache.Users[i].Presence = "active"
				} else {
					cache.Users[i].Presence = "away"
				}

				out = fmt.Sprintf("Presence set to %s", cfg.NewState)
				alfred.SaveJSON(cacheFile, &cache)
			}
		}
	}

	if cfg.StatusText != nil {
		statusText := *cfg.StatusText
		statusEmoji := *cfg.StatusEmoji

		if errStatus = s.SetStatus(statusText, statusEmoji); errStatus == nil {
			i := indexOfUserByID(cache.Auth.UserID)
			if i == -1 {
				dlog.Printf("The user cache is empty")
				errStatus = fmt.Errorf("The user cache is empty")
			} else {
				cache.Users[i].Profile.StatusText = statusText
				cache.Users[i].Profile.StatusEmoji = statusEmoji
				if out != "" {
					out += ", "
				}
				if statusText != "" {
					out += fmt.Sprintf("Status set to %s", statusText)
				} else {
					out += "Status message cleared, emoji set to " + statusEmoji
				}
				alfred.SaveJSON(cacheFile, &cache)
			}
		}
	}

	if errPresence != nil {
		err = errPresence
	} else {
		err = errStatus
	}

	return
}

type statusConfig struct {
	NewState    Presence
	StatusText  *string
	StatusEmoji *string
}
