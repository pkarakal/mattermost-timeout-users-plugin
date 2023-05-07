package main

import (
	"github.com/mattermost/mattermost-server/server/v8/plugin"
)

func main() {
	plugin.ClientMain(&TimeoutUsersPlugin{})
}
