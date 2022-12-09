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
	fmt.Printf("Sending message: %v\n", msg)

	_, err := conn.Write([]byte(msg))
	if err != nil {
		fmt.Println("Write data failed: ", err.Error())
	}
}

func receiveMessage(conn net.Conn) string {
	response := make([]byte, 1024)
	_, err := conn.Read(response)
	if err != nil {
		fmt.Println("Read data failed: ", err.Error())
	}

	return string(response)
}

func main() {
	config, err := parser.GetConfig(os.Args[1:])
	if err != nil {
		panic(err)
	}

	tcpAddr := config["host"] + ":" + config["port"]
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
	fmt.Println("====================")

	for {
		var reqMsg string
		fmt.Print("Your message: ")
		inputReader := bufio.NewReader(os.Stdin)
		reqMsg, _ = inputReader.ReadString('\n')
		reqMsg = strings.TrimSuffix(reqMsg, "\n")
		sendMessage(conn, reqMsg)

		received := receiveMessage(conn)
		fmt.Printf("Response: %v\n", string(received))
	}
}
