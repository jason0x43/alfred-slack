package main

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

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

	var user User
	if cfg.User != "" {
		if u, found := getUser(cfg.User); found {
			user = u
		}
	}

	if user.ID != "" {
		if alfred.FuzzyMatches("username:", arg) {
			items = append(items, alfred.Item{
				Title: fmt.Sprintf("Username: %s", user.Name),
			})
		}

		if alfred.FuzzyMatches("status:", arg) {
			item := alfred.Item{
				Title: fmt.Sprintf("Status: %s", user.Profile.StatusText),
			}

			if user.Profile.StatusEmoji != "" {
				var emojiFile string
				emojiFile, err = getEmojiFromSprite(user.Profile.StatusEmoji)
				if err != nil {
					emojiFile, err = getEmojiFromSlack(user.Profile.StatusEmoji)
				}

				if err == nil {
					dlog.Printf("Setting icon to", emojiFile)
					item.Icon = emojiFile
				}
			}

			items = append(items, item)
		}

		if alfred.FuzzyMatches("presence:", arg) {
			items = append(items, alfred.Item{
				Title: fmt.Sprintf("Presence: %s", user.Presence),
			})
		}

		if alfred.FuzzyMatches("id:", arg) {
			items = append(items, alfred.Item{
				Title: fmt.Sprintf("ID: %s", user.ID),
			})
		}

		if alfred.FuzzyMatches("name:", arg) {
			items = append(items, alfred.Item{
				Title: fmt.Sprintf("Name: %s", user.Profile.RealName),
			})
		}

		if alfred.FuzzyMatches("email:", arg) {
			items = append(items, alfred.Item{
				Title: fmt.Sprintf("Email: %s", user.Profile.Email),
			})
		}
	} else {
		for _, user := range cache.Users {
			if user.Deleted {
				dlog.Print("Skipping deleted user ", user.Name)
				continue
			}

			if user.Profile.Email == "" {
				dlog.Print("Skipping fake user ", user.Name)
				continue
			}

			if channel != nil && !isInChannel(user.ID, channel) {
				continue
			}

			if alfred.FuzzyMatches(user.ID, arg) || alfred.FuzzyMatches(user.Profile.RealName, arg) {
				item := alfred.Item{
					Title:        user.Name,
					Subtitle:     user.Profile.StatusText,
					Autocomplete: user.Name,
					Arg: &alfred.ItemArg{
						Keyword: "users",
						Data:    alfred.Stringify(&userConfig{User: user.ID}),
					},
				}

				if user.Presence == PresenceAway {
					item.Title = fmt.Sprintf("%s %s", AwayMarker, item.Title)
				} else {
					item.Title = fmt.Sprintf("%s %s", ActiveMarker, item.Title)
				}

				// Show a user's status icon if they have one set
				if user.Profile.StatusEmoji != "" {
					var emojiFile string
					emojiFile, err = getEmojiFromSprite(user.Profile.StatusEmoji)
					if err != nil {
						emojiFile, err = getEmojiFromSlack(user.Profile.StatusEmoji)
					}

					if err == nil {
						dlog.Printf("Setting icon to", emojiFile)
						item.Icon = emojiFile
					}
				}

				item.AddMod(alfred.ModCmd, alfred.ItemMod{
					Subtitle: "Chat with user",
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
				})

				item.AddMod(alfred.ModAlt, alfred.ItemMod{
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
		sort.Stable(byStatus(items))
	}

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
		time.Sleep(1 * time.Second)
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
	User      string
	ToMessage *dmID
	ToOpen    *dmID
	Channel   *string
}

type byStatus alfred.Items

func (b byStatus) Len() int {
	return len(b)
}

func (b byStatus) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

func (b byStatus) Less(i, j int) bool {
	dlog.Printf("comparing %s to %s", b[i].Title, b[j].Title)
	// TODO Actually look at the user, not just item data
	return strings.HasPrefix(b[i].Title, string(ActiveMarker)) && strings.HasPrefix(b[j].Title, string(AwayMarker))
}
