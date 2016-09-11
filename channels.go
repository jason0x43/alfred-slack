package main

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/jason0x43/go-alfred"
	"github.com/pkg/browser"
)

// ChannelsCommand allows an API token to be specified
type ChannelsCommand struct{}

// About returns information about a command
func (c ChannelsCommand) About() alfred.CommandDef {
	return alfred.CommandDef{
		Keyword:     "channels",
		Description: "List channels",
		IsEnabled:   config.APIToken != "",
		Arg: &alfred.ItemArg{
			Keyword: "channels",
		},
	}
}

// Items returns the items for the command
func (c ChannelsCommand) Items(arg, data string) (items []alfred.Item, err error) {
	if err = checkRefresh(); err != nil {
		return
	}

	var cfg channelConfig
	if data != "" {
		if err := json.Unmarshal([]byte(data), &cfg); err != nil {
			dlog.Printf("Invalid channel config")
		}
	}

	var cid string
	if cfg.Channel != nil {
		cid = *cfg.Channel
	}

	if cid != "" {
		var property string
		if cfg.Property != nil {
			property = *cfg.Property
		}

		if property != "" {
			if property == "pins" {
				var pins []Pin
				s := OpenSession(config.APIToken)
				if pins, err = s.GetPins(cid); err == nil {
					urlMatcher := regexp.MustCompile(`<?(https?://\S+)>?`)

					for _, pin := range pins {
						title := pin.Title()

						item := alfred.Item{
							Title: pin.Title(),
						}

						if urlMatcher.MatchString(title) {
							url := urlMatcher.FindStringSubmatch(title)[1]

							item.Arg = &alfred.ItemArg{
								Keyword: "channels",
								Mode:    alfred.ModeDo,
								Data: alfred.Stringify(&channelConfig{
									ToBrowse: &url,
								}),
							}
						} else if pin.File != nil {
							if icon, err := getFile(pin.File.Thumb64, ""); err == nil {
								item.Icon = icon
							}

							item.Arg = &alfred.ItemArg{
								Keyword: "channels",
								Mode:    alfred.ModeDo,
								Data: alfred.Stringify(&channelConfig{
									ToBrowse: &pin.File.PrivateURL,
								}),
							}
						}

						items = append(items, item)
					}

					alfred.FuzzySort(items, arg)
				}
			}
		} else {
			if alfred.FuzzyMatches("Pins...", arg) {
				property = "pins"
				item := alfred.Item{
					// UID:          c.Workflow().BundleID,
					Title:        "Pins...",
					Subtitle:     "List the pins for this channel",
					Autocomplete: "Pins...",
					Arg: &alfred.ItemArg{
						Keyword: "channels",
						Data: alfred.Stringify(&channelConfig{
							Channel:  &cid,
							Property: &property,
						}),
					},
				}
				items = append(items, item)
			}

			if alfred.FuzzyMatches("Members...", arg) {
				property = "members"
				item := alfred.Item{
					Title:        "Members...",
					Subtitle:     "List the members in this channel",
					Autocomplete: "Members...",
					Arg: &alfred.ItemArg{
						Keyword: "users",
						Data: alfred.Stringify(&userConfig{
							Channel: &cid,
						}),
					},
				}
				items = append(items, item)
			}
		}
	} else {
		for _, channel := range cache.Channels {
			if !cfg.ShowAll && !isInChannel(cache.Auth.UserID, &channel) {
				continue
			}

			if alfred.FuzzyMatches(channel.Name, arg) {
				item := alfred.Item{
					Title:        channel.Name,
					Autocomplete: channel.Name,
					UID:          channel.ID,
					Arg: &alfred.ItemArg{
						Keyword: "channels",
						Mode:    alfred.ModeDo,
						Data: alfred.Stringify(&channelConfig{
							ToOpen: &channelID{
								Channel: channel.ID,
								Team:    cache.Auth.TeamID,
							},
						}),
					},
				}

				if !isInChannel(cache.Auth.UserID, &channel) {
					item.Icon = "icon_faded.png"
				}

				item.AddMod(alfred.ModCmd, alfred.ItemMod{
					Subtitle: "Details...",
					Arg: &alfred.ItemArg{
						Keyword: "channels",
						Data:    alfred.Stringify(&channelConfig{Channel: &channel.ID}),
					},
				})

				items = append(items, item)
			}
		}

		alfred.FuzzySort(items, arg)
	}

	return
}

// Do implements the command
func (c ChannelsCommand) Do(data string) (out string, err error) {
	var cfg channelConfig
	if data != "" {
		if err := json.Unmarshal([]byte(data), &cfg); err != nil {
			return "", fmt.Errorf("Invalid open message")
		}
	}

	if cfg.ToOpen != nil {
		browser.OpenURL(fmt.Sprintf("slack://channel?id=%s&team=%s", cfg.ToOpen.Channel, cfg.ToOpen.Team))
	}

	if cfg.ToBrowse != nil {
		browser.OpenURL(*cfg.ToBrowse)
	}

	return
}

type channelID struct {
	Channel string
	Team    string
}

type channelConfig struct {
	Channel  *string
	ToOpen   *channelID
	Property *string
	ToBrowse *string
	ShowAll  bool
}
