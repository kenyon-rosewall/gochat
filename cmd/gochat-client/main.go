package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strings"

	"bitbucket.org/krosewall/gochat/pkg/parser"
	"bitbucket.org/krosewall/gochat/pkg/transfer"
	"bitbucket.org/krosewall/gochat/pkg/utils"
)

const quitMessage = "/quit"

type client struct {
	conn      *net.TCPConn
	username  string
	room      string
	key       string
	sessionID uint64
}

func randomUsername() string {
	nameLength := 5 + rand.Intn(5)
	username := utils.GetRandomString(nameLength)

	return username
}

func writeMessage(msg string) {
	fmt.Println(msg)
}

func sendMessage(c *client, inputReader bufio.Reader, chMsg chan string) {
	for {
		firstByte, _ := inputReader.ReadByte()
		if firstByte == '\n' {
			fmt.Print(">> ")
			continue
		}
		inputReader.UnreadByte()

		msg, _ := inputReader.ReadString('\n')
		msg = strings.TrimSpace(msg)

		if msg == quitMessage {
			quitTransfer := transfer.PackTransfer(c.sessionID, transfer.Quit, "", c.key)
			c.conn.Write(quitTransfer)

			writeMessage(fmt.Sprintf("You have left the room %s", c.room))
		} else {
			sendMsg := fmt.Sprintf("[%s] %s", c.username, msg)

			t := transfer.PackTransfer(c.sessionID, transfer.Message, sendMsg, c.key)
			c.conn.Write(t)

			writeMessage(fmt.Sprintf("[you] %s", msg))
		}

		chMsg <- msg
	}
}

func receiveMessage(c *client, chMsg chan string) {
	t := transfer.UnpackTransfer(c.conn, c.key)
	msg := t.Body

	if t.MessageType == transfer.Quit {
		msg = quitMessage
	} else if t.MessageType == transfer.Command {

	} else {
		writeMessage(fmt.Sprintf("\n%s", msg))
	}

	chMsg <- msg
}

func connect(args []string) (*client, error) {
	// Defaults
	username := randomUsername()
	room := "default"
	key := ""

	config, err := parser.GetConfig(args)
	if err != nil {
		return nil, err
	} else {
		tcpAddr := config["host"] + ":" + config["port"]
		key = config["key"]

		tcpServer, err := net.ResolveTCPAddr("tcp", tcpAddr)
		if err != nil {
			return nil, err
		} else {
			conn, err := net.DialTCP("tcp", nil, tcpServer)
			if err != nil {
				return nil, err
			}

			c := &client{
				conn:     conn,
				username: username,
				room:     room,
				key:      key,
			}

			return c, nil
		}
	}
}

func main() {
	c, err := connect(os.Args[1:])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer c.conn.Close()

	fmt.Printf("Remote Addr: %v\n", c.conn.RemoteAddr())
	fmt.Println("====================")

	inputReader := bufio.NewReader(os.Stdin)
	fmt.Print("Username: ")
	u_username, _ := inputReader.ReadString('\n')
	u_username = strings.TrimSpace(u_username)
	if u_username != "" {
		c.username = u_username
	}

	fmt.Print("Room: ")
	u_room, _ := inputReader.ReadString('\n')
	u_room = strings.TrimSpace(u_room)
	if u_room != "" {
		c.room = u_room
	}

	fmt.Println("====================")

	fmt.Fprintf(c.conn, c.username+"\n")
	fmt.Fprintf(c.conn, c.room+"\n")

	fmt.Fscanf(c.conn, "%d\n", &c.sessionID)

	fmt.Printf("You have been given session id %d\n", c.sessionID)

	// sigChannel := make(chan os.Signal, 1)
	// signal.Notify(sigChannel, os.Interrupt, syscall.SIGTERM)
	msg := make(chan string)

	for {
		fmt.Print(">> ")

		go sendMessage(c, *inputReader, msg)
		go receiveMessage(c, msg)

		if strings.TrimSpace(string(<-msg)) == quitMessage {
			fmt.Println("TCP client exiting...")
			return
		}
	}
}
