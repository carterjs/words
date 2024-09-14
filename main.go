package main

import (
	"fmt"
	"github.com/carterjs/words/cmd/cli"
	"github.com/carterjs/words/cmd/server"
	"os"
)

func main() {
	cmd := cli.Command()
	cmd.AddCommand(server.Command())

	err := cmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
