package main

import "github.com/mar1mo-41414/ore2ca/cmd"

// version is set by ldflags at build time (e.g. -X main.version=v1.1)
var version = "dev"

func main() {
	cmd.Execute(version)
}
