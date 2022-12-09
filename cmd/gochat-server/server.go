package main

import (
	"fmt"
	"log"
	"net"
	"time"
)

func handleConnection(conn net.Conn) {
	buffer := make([]byte, 1024)
	_, err := conn.Read(buffer)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("Request: %v\n", string(buffer))

	if len(buffer) > 0 {
		time := time.Now().Format(time.ANSIC)
		resp := fmt.Sprintf("Your message is: %v. Received time: %v", string(buffer[:]), time)
		conn.Write([]byte(resp))
	}

	// conn.Close()
}

func StartServer(config map[string]string) error {
	ln, err := net.Listen("tcp", config["host"]+":"+config["port"])
	if err != nil {
		return err
	}
	defer ln.Close()

	if ln != nil {
		fmt.Println("Server listening on port " + config["port"])
		for {
			conn, err := ln.Accept()
			if err != nil {
				log.Println(err)
			}

			go func(conn net.Conn) {
				handleConnection(conn)
			}(conn)
		}
	}

	return nil
}
