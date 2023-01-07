package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/kenyon-rosewall/gochat/pkg/parser"
)

func sendMessage(conn net.Conn, msg string) {
	fmt.Fprintf(conn, msg+"\n")
}

func receiveMessage(conn net.Conn) string {
	response, _ := bufio.NewReader(conn).ReadString('\n')
	fmt.Print("~>: " + response)

	return string(response)
}

func main() {
	config, err := parser.GetConfig(os.Args[1:])
	if err != nil {
		panic(err)
	}

	tcpAddr := config["host"] + ":" + config["port"]
	username := config["username"]
	room := "default"
	if len(config["room"]) > 3 {
		room = config["room"]
	}

	tcpServer, err := net.ResolveTCPAddr("tcp", tcpAddr)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Port: %v\n", tcpServer.Port)

	conn, err := net.DialTCP("tcp", nil, tcpServer)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	fmt.Printf("Remote Addr: %v\n", conn.RemoteAddr())
	fmt.Printf("Username: %s\tRoom: %s\n", username, room)
	fmt.Println("====================")

	fmt.Fprintf(conn, username+"\n")
	fmt.Fprintf(conn, room+"\n")

	for {
		inputReader := bufio.NewReader(os.Stdin)
		fmt.Print(">> ")
		reqMsg, _ := inputReader.ReadString('\n')
		reqMsg = strings.TrimSuffix(reqMsg, "\n")

		sendMessage(conn, reqMsg)
		resp := receiveMessage(conn)

		if strings.TrimSpace(string(resp)) == "STOP" {
			fmt.Println("TCP client exiting...")
			return
		}
	}
}
