package main

import (
	"fmt"
	"net"
)

const ListenPortLocal = ":4213"
const ListenPortPeer = ":4214"
const rcvBufLen = 1024

func main() {
		// open the listen port for peers
	listenerPeer, errPeer := net.Listen("tcp", ListenPortPeer)

	if errPeer != nil {
		fmt.Println("Listener port has been used:", errPeer.Error())
		return
	}

	go handleConnectionFromPeers(listenerPeer)

	// active connect to application
	activeTest()
	
	// open the listen port for local app
	listenerLocal, errLocal := net.Listen("tcp", ListenPortLocal)

	if errLocal != nil {
		fmt.Println("Listener port has been used:", errLocal.Error())
		return
	}
	
	// main routine: commmunication between server and app
	handleConnectionFromLocal(listenerLocal)

}
