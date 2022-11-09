package main

import (
	"os"
)

func main() {
	config, err := GetConfig(os.Args[1:])
	if err != nil {
		panic(err)
	}

	err = StartServer(config)
	if err != nil {
		panic(err)
	}
}
