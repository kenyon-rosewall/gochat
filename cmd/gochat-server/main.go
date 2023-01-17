package main

import (
	"os"

	"bitbucket.org/krosewall/gochat/pkg/parser"
)

func main() {
	config, err := parser.GetConfig(os.Args[1:])
	if err != nil {
		panic(err)
	}

	err = StartServer(config)
	if err != nil {
		panic(err)
	}
}
