alfred-slack
============

An Alfred workflow for Slacking

![Screenshot](doc/main_menu.png?raw=true)

Installation
------------

Download the latest workflow package from the [releases page](https://github.com/jason0x43/alfred-slack/releases) and double click it — Alfred will take care of the rest.


Usage
-----

The workflow provides one universal keyword, “slk”, from which all functions can be accessed. The first time you use the keyword, only one item will be available: `token`. Actioning `token` will pop up a dialog asking you to enter your API token ([test tokens](https://api.slack.com/docs/oauth-test-tokens) are from Slack if you're logged in).

The main functions also have individual keywords:

* `slc` - List subscribed channels
* `slc*` - List all channels
* `slu` - List users on your team
* `slp` - Show and update your presence


Channels
--------

The `channels` command (or `slc`) will list the channels you are subscribed to. An `slc*` command will list _all_ the channels available from your team.Actioning a channel will open it in the Slack app. Holding Cmd while actioning the channel will bring up a list of channel properties, currently “Pins...” and “Members...”, which can be actioned for more information.


Users
-----

The `users` command (or `slu`) will list the users in your team. Actioning a user will open a direct message to that user. Holding Cmd while actioning the user will open the user’s profile in the Slack app.


Presence
--------

The `presence` command (or `slp`) will show your current presence state, either “Active” or “Away”. Actioning the state item will toggle it.
