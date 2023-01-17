package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"log"
	"math/rand"
	"net"
	"strings"
)

const headerLength = 18
const (
	datatype_Macsum = iota
	datatype_Nonce
	datatype_Message
	datatype_Quit
)

type client struct {
	conn      net.Conn
	reader    *bufio.Reader
	room      string
	username  string
	sessionID uint64
}

type room struct {
	name    string
	clients map[*client]struct{}
}

func newRoom(name string) *room {
	return &room{
		name:    name,
		clients: make(map[*client]struct{}),
	}
}

func handleConnection(conn net.Conn, rooms map[string]*room) {
	cl := &client{
		conn:      conn,
		reader:    bufio.NewReader(conn),
		sessionID: uint64(rand.Int63()),
	}

	// Encrypt these packets
	username, err := cl.reader.ReadString('\n')
	if err != nil {
		fmt.Println(err)
		return
	}
	cl.username = strings.TrimSpace(username)

	roomname, err := cl.reader.ReadString('\n')
	if err != nil {
		fmt.Println(err)
		return
	}
	cl.room = strings.TrimSpace(roomname)

	r, ok := rooms[cl.room]
	if !ok {
		r = newRoom(cl.room)
		rooms[cl.room] = r
	}
	r.clients[cl] = struct{}{}

	fmt.Printf("%s has joined the room %s\n", cl.username, cl.room)

	fmt.Fprintf(conn, "%d\n", cl.sessionID)

	wantToQuit := false
	for {
		var packets [3][]byte
		for i := 0; i < 3; i++ {
			header := make([]byte, headerLength)
			_, err := conn.Read(header)
			if err != nil {
				fmt.Println("Could not read header from packet")
			}

			dataType := binary.BigEndian.Uint16(header[8:10])
			if dataType == datatype_Quit {
				wantToQuit = true
				break
			}

			dataLength := binary.BigEndian.Uint64(header[10:])
			fmt.Println(dataLength)
			data := make([]byte, dataLength)
			_, err = conn.Read(data)
			if err != nil {
				fmt.Println("Could not read data from packet")
			}

			packets[i] = append(packets[i], append(header, data...)...)
		}

		for c := range r.clients {
			if c != cl {
				for i := 0; i < 3; i++ {
					if _, err := c.conn.Write(packets[i]); err != nil {
						fmt.Println("Error when sending message to the client:", err)
					}
					packets[i] = []byte{}
				}
			}
		}

		if wantToQuit {
			break
		}
	}

	delete(r.clients, cl)
	fmt.Printf("%s has left the room %s\n", cl.username, cl.room)

	if len(r.clients) == 0 {
		delete(rooms, r.name)
		fmt.Printf("Room %s has been closed\n", r.name)
	}

	conn.Close()
}

func StartServer(config map[string]string) error {
	ln, err := net.Listen("tcp", config["host"]+":"+config["port"])
	if err != nil {
		return err
	}
	defer ln.Close()

	if ln != nil {
		fmt.Println("Server listening on port " + config["port"])

		rooms := make(map[string]*room)
		rooms["default"] = &room{name: "default"}

		for {
			conn, err := ln.Accept()
			if err != nil {
				log.Println(err)
				continue
			}

			go handleConnection(conn, rooms)
		}
	}

	return nil
}
