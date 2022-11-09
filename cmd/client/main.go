package main

import (
	"fmt"
	"net"
	"os"
)

func main() {
	config, err := GetConfig(os.Args)

	tcpServer, err := net.ResolveTCPAddr("tcp", config["host"]+":"+config["port"])
	conn, err := net.DialTCP("tcp", nil, tcpServer)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	_, err = conn.Write([]byte("This is a message"))
	if err != nil {
		fmt.Println("Write data failed: ", err.Error())
	}

	received := make([]byte, 1024)
	_, err = conn.Read(received)
	if err != nil {
		fmt.Println("Read data failed: ", err.Error())
	}
}
