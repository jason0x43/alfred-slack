package main

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"

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
			if alfred.FuzzyMatches("open", arg) {
				item := alfred.Item{
					UID:          fmt.Sprintf("%s.channels.open", workflow.BundleID()),
					Title:        "Open",
					Subtitle:     "Open this channel in the Slack app",
					Autocomplete: "Open",
					Arg: &alfred.ItemArg{
						Keyword: "channels",
						Mode:    alfred.ModeDo,
						Data: alfred.Stringify(&channelConfig{
							ToOpen: &channelID{
								Channel: cid,
								Team:    cache.Auth.TeamID,
							},
						}),
					},
				}
				items = append(items, item)
			}

			if alfred.FuzzyMatches("pins", arg) {
				property = "pins"
				item := alfred.Item{
					UID:          fmt.Sprintf("%s.channels.pins", workflow.BundleID()),
					Title:        "Pins",
					Subtitle:     "List the pins for this channel",
					Autocomplete: "Pins",
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

			if alfred.FuzzyMatches("members", arg) {
				item := alfred.Item{
					UID:          fmt.Sprintf("%s.channels.members", workflow.BundleID()),
					Title:        "Members",
					Subtitle:     "List the members in this channel",
					Autocomplete: "Members",
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
			if alfred.FuzzyMatches(channel.Name, arg) {
				item := alfred.Item{
					Title:        channel.Name,
					Autocomplete: channel.Name,
					Arg: &alfred.ItemArg{
						Keyword: "channels",
						Data:    alfred.Stringify(&channelConfig{Channel: &channel.ID}),
					},
				}

				if !isInChannel(cache.Auth.UserID, &channel) {
					item.Icon = "icon_faded.png"

					// If the user isn't subscribed to the channel, take away
					// its UID so that Alfred will leave it after the
					// subscribed channels
					item.UID = ""
				}

				item.AddMod(alfred.ModCmd, alfred.ItemMod{
					Subtitle: "Open this channel",
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
				})

				items = append(items, item)
			}
		}

		alfred.FuzzySort(items, arg)
		sort.Stable(bySubscription(items))
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
		err = browser.OpenURL(fmt.Sprintf("slack://channel?id=%s&team=%s", cfg.ToOpen.Channel, cfg.ToOpen.Team))
	}

	if cfg.ToBrowse != nil {
		err = browser.OpenURL(*cfg.ToBrowse)
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
}

type bySubscription alfred.Items

func (b bySubscription) Len() int {
	return len(b)
}

func (b bySubscription) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

func (b bySubscription) Less(i, j int) bool {
	return b[i].Icon != "icon_faded.png" && b[j].Icon == "icon_faded.png"
}
