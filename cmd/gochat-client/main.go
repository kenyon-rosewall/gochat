package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"math/rand"
	"net"
	"os"
	"strings"

	"bitbucket.org/krosewall/gochat/pkg/parser"
	"bitbucket.org/krosewall/gochat/pkg/utils"
)

const headerLength = 18
const (
	datatype_Macsum = iota
	datatype_Nonce
	datatype_Message
	datatype_Quit
)

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

func createHeader(sessionID uint64, dataType uint16, data string) []byte {
	header := make([]byte, headerLength)
	binary.BigEndian.PutUint64(header[:8], sessionID)
	binary.BigEndian.PutUint16(header[8:10], dataType)
	binary.BigEndian.PutUint64(header[10:], uint64(len(data)))

	return header
}

func unpackHeader(conn net.Conn) (uint64, uint16, []byte) {
	header := make([]byte, headerLength)
	_, err := conn.Read(header)
	if err != nil {
		if err == io.EOF {
			// connection closed by the other side
			return 0, 0, nil
		}
		fmt.Println("Could not receive message header: ", err)
	}
	sessionID := binary.BigEndian.Uint64(header[:8])
	dataType := binary.BigEndian.Uint16(header[8:10])
	dataLength := binary.BigEndian.Uint64(header[10:])
	data := make([]byte, dataLength)
	_, err = conn.Read(data)
	if err != nil {
		fmt.Printf("Could not read data for sessionID %d and data type %d\n", sessionID, dataType)
	}

	return sessionID, dataType, data
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

		if msg == "/stop" {
			quitheader := createHeader(c.sessionID, datatype_Quit, "")
			c.conn.Write(quitheader)

			writeMessage(fmt.Sprintf("You have left the room %s", c.room))
		} else {
			sendMsg := fmt.Sprintf("[%s] %s", c.username, msg)
			ciphermsg, nonce, mac := Encrypt(sendMsg, c.key)

			cipherheader := createHeader(c.sessionID, datatype_Message, ciphermsg)
			c.conn.Write(append(cipherheader, ciphermsg...))
			nonceheader := createHeader(c.sessionID, datatype_Nonce, nonce)
			c.conn.Write(append(nonceheader, nonce...))
			macheader := createHeader(c.sessionID, datatype_Macsum, mac)
			c.conn.Write(append(macheader, mac...))

			writeMessage(fmt.Sprintf("[you] %s", msg))
		}

		chMsg <- msg
	}
}

// TODO: Implement a header message that states what type of message is coming
func receiveMessage(c *client, chMsg chan string) {
	msg := "/stop"
	cipherSession, cipherType, cipher := unpackHeader(c.conn)
	if cipherSession > 0 && cipherType > 0 && len(cipher) > 0 {
		nonceSession, nonceType, nonce := unpackHeader(c.conn)
		macsumSession, macsumType, macsum := unpackHeader(c.conn)

		if cipherSession != nonceSession || cipherSession != macsumSession {
			fmt.Println("Session ids do not match")
			return
		}
		if cipherType != datatype_Message || nonceType != datatype_Nonce || macsumType != datatype_Macsum {
			fmt.Println("Data types do not line up")
			return
		}

		msg := Decrypt(cipher, nonce, macsum, c.key)
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

		if strings.TrimSpace(string(<-msg)) == "/stop" {
			fmt.Println("TCP client exiting...")
			return
		}
	}
}
