package main

import "github.com/jason0x43/go-alfred"

// ResetCommand resets all stored API tokens
type ResetCommand struct{}

// About returns information about a command
func (c ResetCommand) About() alfred.CommandDef {
	return alfred.CommandDef{
		Keyword:     "reset",
		Description: "Reset the workflow, erasing all local data",
		IsEnabled:   config.APIToken != "",
		Arg: &alfred.ItemArg{
			Keyword: "reset",
			Mode:    alfred.ModeDo,
		},
	}
}

// Do implements the command
func (c ResetCommand) Do(data string) (out string, err error) {
	config = configStruct{}
	if err = alfred.SaveJSON(configFile, &config); err != nil {
		return
	}

	cache = cacheStruct{}
	err = alfred.SaveJSON(cacheFile, &cache)

	workflow.ShowMessage("The Slack workflow has been reset")
	return
}
