package parser

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func getValidKeys() (validKeys []string) {
	return []string{"host", "port", "key"}
}

func getFilename(args []string) (filename string) {
	filename = "./gochat.cfg"
	if len(args) > 0 {
		filename = args[0]
	}

	return filename
}

func isKeyValid(needle string) (exists bool) {
	exists = false
	validKeys := getValidKeys()

	for _, key := range validKeys {
		if needle == key {
			exists = true
		}
	}

	return exists
}

func GetConfig(args []string) (config map[string]string, err error) {
	config = make(map[string]string, 10)
	filename := getFilename(args)

	file, err := os.Open(filename)
	if err != nil {
		return config, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	atLine := 0
	for scanner.Scan() {
		line := scanner.Text()
		atLine++

		kvp := strings.Split(line, "=")

		if len(kvp) != 2 {
			return config, fmt.Errorf("malformed line in config %s at line %d", filename, atLine)
		}

		if !isKeyValid(kvp[0]) {
			continue
		}

		config[kvp[0]] = kvp[1]
	}

	err = scanner.Err()

	return config, err
}
