alfred-slack
============

An Alfred workflow for Slacking

![Screenshot](doc/main_menu.png?raw=true)

Installation
------------

Download the latest workflow package from the [releases
page](https://github.com/jason0x43/alfred-slack/releases) and double click it —
Alfred will take care of the rest.


Usage
-----

The workflow provides one universal keyword, “slk”, from which all functions
can be accessed. The first time you use the keyword, only one item will be
available: `token`. Actioning `token` will pop up a dialog asking you to enter
your API token ([test tokens](https://api.slack.com/docs/oauth-test-tokens) are
from Slack if you're logged in).

The main functions also have individual keywords:

* `slc` - List subscribed channels
* `slu` - List users on your team
* `sls` - Show and update your status and presence

Initially, `slk` will ask for an API token (Oauth support is coming). To get a
token, browse to https://api.slack.com/custom-integrations/legacy-tokens
(you’ll need to login) and generate one. Then run `slk`, action the
“Manually enter a token” item, and paste the token value into the resulting
dialog.


Channels
--------

The `channels` command (or `slc`) will list the channels available to your
team. Channels you are not subscribed to will have a faded icon.

* Actioning the channel will bring up a list of channel properties, currently
  “Pins...” and “Members...”, which can be actioned for more information.
* Holding Cmd while actioning a channel will open it in the Slack app.


Users
-----

The `users` command (or `slu`) will list the users in your team. The user’s
presence is represented by a filled (active) or empty (away) disc next to the
name. If the user has a status emoji set, that emoji will be used as the item
icon. If the user has a status message set, that will be the item’s subtitle.

* Actioning a user will open more information about the user.
* Holding Cmd while actioning the user will open a chat with the user.
* Holding Alt while actioning the user will open the user’s profile in the
  Slack app.


Status
------

The `status` command (or `sls`) will show your current presence state, either
“Active” or “Away”.

* Entering a status message (or not) and actioning the item will allow the user
  to select a status icon. Actioning an icon (or the “No status icon” option)
  will update your status message and status icon.
* Holding Cmd while entering a status message (or not) and actioning the item
  will follow the same steps, but will set your status to ‘active’.
* Holding Alt while entering a status message (or not) and actioning the item
  will follow the same steps, but will set your status to ‘away’.
