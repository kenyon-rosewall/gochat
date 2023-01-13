package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
)

type client struct {
	conn     net.Conn
	reader   *bufio.Reader
	room     string
	username string
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
		conn:   conn,
		reader: bufio.NewReader(conn),
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

	for {
		message, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			fmt.Println(err)
			return
		}

		// Figure out another way (with a new variable) that we can communicate a STOP message
		if strings.TrimSpace(message) == "STOP" {
			break
		}

		for c := range r.clients {
			if c != cl {
				messageline := fmt.Sprintf("%s:%s>%s\n", cl.username, cl.room, message)
				c.conn.Write([]byte(messageline))
			} else {
				messageline := fmt.Sprintf("you:%s>%s\n", cl.room, message)
				cl.conn.Write([]byte(messageline))
			}
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
