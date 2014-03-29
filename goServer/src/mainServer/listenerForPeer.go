package main

import (
	"fmt"
	"net"
)

func HandleConnectionFromPeers(listener net.Listener) {
	for {
		conn, err := listener.Accept()

		if err != nil {
			fmt.Println("Server: Accept Error:", err.Error())
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
		fmt.Println("Server: Read Error:", err.Error())
		return
	}

	fmt.Printf("Server: Connection constructed from peer,  %d bytes have been received.\n", size)
}
