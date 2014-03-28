package main

import (
	"fmt"
	"net"
	"os"
)

var ListenPortLocal = ":4213"
var ListenPortPeer = ":4214"

var localIP = "127.0.0.1"

const rcvBufLen = 1024

var isSN = false

func maina() {
	parseArguments()

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
		fmt.Println("Server: Listener port has been used:", errLocal.Error())
		return
	}

	// main routine: commmunication between server and app
	handleConnectionFromLocal(listenerLocal)
}

// parse the go argument [locaPort, peerPort] isSN
func parseArguments() {
	argLen := len(os.Args)

	if argLen > 1 {
		if os.Args[1] == "True" {
			isSN = true
			if argLen > 2 {
				ListenPortLocal = ":" + os.Args[2]
				if argLen > 3 {
					ListenPortPeer = ":" + os.Args[3]
				}
			}
		}
	}
}
