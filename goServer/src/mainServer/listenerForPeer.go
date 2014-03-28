package main

import (
	"fmt"
	"net"
	"encoding/json"
)

func handleConnectionFromPeers(listener net.Listener) {
	for {
		conn, err := listener.Accept()

		if err != nil {
			fmt.Println("Accept Error:", err.Error())
			return
		}

		// new thread to handle request from peers
		go handleConnectionFromPeersThread(conn)
	}
}

func handleConnectionFromPeersThread(conn net.Conn) {
	rcvbuf := make([]byte, rcvBufLen)

	// receive messages
	size, err := conn.Read(rcvbuf)

	if err != nil {
		fmt.Println("Read Error:", err.Error())
		return
	}

	var msg map[string]string

	err = json.Unmarshal(rcvbuf[:size], &msg)

	if err != nil {
		fmt.Println("Unmarshal Error:", err.Error())
		return
	}

	msgType, exist := msg["type"]

	if exist == false {
		fmt.Println("No Message Type!")
		return
	}

	fmt.Println(msg)

	// handle different type of message
	switch msgType {
	default:
		fmt.Println("Messge Type Undefined")
	}
}
