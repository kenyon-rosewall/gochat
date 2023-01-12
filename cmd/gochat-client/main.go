package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strings"
	"time"

	"github.com/kenyon-rosewall/gochat/pkg/parser"
)

func randomUsername() string {
	rand.Seed(time.Now().UnixNano())
	nameLength := 5 + rand.Intn(5)
	username := make([]byte, nameLength)

	for i := 0; i < nameLength; i++ {
		username[i] = byte(97 + rand.Intn(25))
	}

	return string(username)
}

func sendMessage(conn net.Conn, inputReader bufio.Reader, chMsg chan string) {
	for {
		fmt.Print(">> ")
		firstByte, _ := inputReader.ReadByte()
		if firstByte == '\n' {
			continue
		}
		inputReader.UnreadByte()

		msg, _ := inputReader.ReadString('\n')
		msg = strings.TrimSpace(msg)

		fmt.Fprintf(conn, msg+"\n")
		chMsg <- msg
		break
	}
}
func receiveMessage(conn net.Conn) string {
	response, _ := bufio.NewReader(conn).ReadString('\n')
	fmt.Print(response)

	return string(response)
}

func main() {
	config, err := parser.GetConfig(os.Args[1:])
	if err != nil {
		fmt.Println(err)
		return
	}

	tcpAddr := config["host"] + ":" + config["port"]
	username := randomUsername()
	room := "default"

	tcpServer, err := net.ResolveTCPAddr("tcp", tcpAddr)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("Port: %v\n", tcpServer.Port)

	conn, err := net.DialTCP("tcp", nil, tcpServer)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	inputReader := bufio.NewReader(os.Stdin)
	fmt.Printf("Remote Addr: %v\n", conn.RemoteAddr())
	fmt.Println("====================")

	fmt.Print("Username: ")
	u_username, _ := inputReader.ReadString('\n')
	u_username = strings.TrimSpace(u_username)
	if u_username != "" {
		username = u_username
	}

	fmt.Print("Room: ")
	u_room, _ := inputReader.ReadString('\n')
	u_room = strings.TrimSpace(u_room)
	if u_room != "" {
		room = u_room
	}

	fmt.Println("====================")

	fmt.Fprintf(conn, username+"\n")
	fmt.Fprintf(conn, room+"\n")

	// sigChannel := make(chan os.Signal, 1)
	// signal.Notify(sigChannel, os.Interrupt, syscall.SIGTERM)
	msg := make(chan string)
	for {
		go sendMessage(conn, *inputReader, msg)
		receiveMessage(conn)

		if strings.TrimSpace(string(<-msg)) == "STOP" {
			fmt.Println("TCP client exiting...")
			return
		}
	}
}
