package main

import (
	"fmt"
	"net"
)

const ListenPort = ":4213"
const rcvBufLen = 1024

func main() {
	// open the listen port for local app
	listenerLocal, errLocal := net.Listen("tcp", ListenPortLocal)

	if errLocal != nil {
		fmt.Println("Listener port has been used:", errLocal.Error())
		return
	}

	// active connect to application
	go activeThread()

	go handleConnectionFromLocal(ListenerLocal)

	// open the listen port for peers
	listenerPeer, errPeer := net.Listen("tcp", ListenPortPeer)

	if errPeer != nil {
		fmt.Println("Listener port has been used:", errPeer.Error())
		return
	}

	go handleConnectionFromPeers(ListenerPeer)
}
