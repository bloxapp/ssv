package main

import "github.com/bloxapp/ssv/cli"

var (
	// AppName is the application name
	AppName = "exporter-Node"

	// Version is the app version
	Version = "latest"
)

func main() {
	cli.Execute(AppName, Version)
}
