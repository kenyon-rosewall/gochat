package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/kenyon-rosewall/gochat/pkg/parser"
)

func randomUsername() string {
	//nameLength := 5 + rand.Intn(5)
	username := "kenyon"
	//username := GetRandomString(nameLength)

	return username
}

func sendMessage(conn net.Conn, inputReader bufio.Reader, key string, chMsg chan string) {
	for {
		fmt.Print(">> ")
		firstByte, _ := inputReader.ReadByte()
		if firstByte == '\n' {
			continue
		}
		inputReader.UnreadByte()

		msg, _ := inputReader.ReadString('\n')
		msg = strings.TrimSpace(msg)

		ciphermsg, nonce, mac := Encrypt(msg, key)

		fmt.Fprintf(conn, "%s:%s:%s\n", nonce, mac, ciphermsg)
		chMsg <- msg
		break
	}
}
func receiveMessage(conn net.Conn, key string) string {
	response, _ := bufio.NewReader(conn).ReadString('\n')

	resArr := strings.Split(response, ">")
	metaArr := strings.Split(resArr[0], ":")
	cipherArr := strings.Split(resArr[1], ":")
	fmt.Println(resArr[1])

	msg := Decrypt(cipherArr[2], cipherArr[0], cipherArr[1], key)

	if msg != "" {
		msgLine := fmt.Sprintf("[%s:%s] %s", metaArr[0], metaArr[1], msg)
		fmt.Print(msgLine)
	}

	return string(msg)
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
	key := config["key"]

	tcpServer, err := net.ResolveTCPAddr("tcp", tcpAddr)
	if err != nil {
		fmt.Println(err)
		return
	}
	conn, err := net.DialTCP("tcp", nil, tcpServer)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	fmt.Printf("Remote Addr: %v\n", conn.RemoteAddr())
	fmt.Println("====================")

	inputReader := bufio.NewReader(os.Stdin)
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
		go sendMessage(conn, *inputReader, key, msg)
		receiveMessage(conn, key)

		if strings.TrimSpace(string(<-msg)) == "STOP" {
			fmt.Println("TCP client exiting...")
			return
		}
	}
}
