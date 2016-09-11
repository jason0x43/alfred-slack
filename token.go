package main

import "github.com/jason0x43/go-alfred"

// TokenCommand allows an API token to be specified
type TokenCommand struct{}

// About returns information about a command
func (c TokenCommand) About() alfred.CommandDef {
	return alfred.CommandDef{
		Keyword:     "token",
		Description: "Manually enter a Slack API token",
		IsEnabled:   config.APIToken == "",
		Arg: &alfred.ItemArg{
			Keyword: "token",
			Mode:    alfred.ModeDo,
		},
	}
}

// Do implements the command
func (c TokenCommand) Do(data string) (out string, err error) {
	dlog.Println("Getting token...")

	var btn string
	var token string

	if btn, token, err = workflow.GetInput("API token", "", false); err != nil {
		return
	}

	if btn != "Ok" {
		dlog.Println("User didn't click OK")
		return
	}

	dlog.Printf("token: %s", token)

	config.APIToken = token
	if err = alfred.SaveJSON(configFile, &config); err != nil {
		return
	}

	workflow.ShowMessage("Token saved!")
	return
}
