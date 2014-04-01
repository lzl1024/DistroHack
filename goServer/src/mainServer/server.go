package main

import (
	"fmt"
	"net"
	"os"
	"util"
	"superNode"
	"msg"
	"strconv"
)

var ListenPortLocal = ":4213"
var ListenPortPeer = 4214

var localIP = "127.0.0.1"

const rcvBufLen = 1024

var isSN = false

func main() {
	parseArguments()
	
	// open database
	util.DatabaseInit()

	initMessagePasser()
	initConnectionFromPeers()
	
	// open SN port when is needed
	if isSN == true {
		go superNode.SuperNodeThread()
	}

	// tests
	tests()
	
	// open the listen port for local app
	listenerLocal, errLocal := net.Listen("tcp", ListenPortLocal)

	if errLocal != nil {
		fmt.Println("Server: Listener port has been used:", errLocal.Error())
		return
	}
	// main routine: commmunication between server and app
	HandleConnectionFromLocal(listenerLocal)
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
					ListenPortPeer, _ = strconv.Atoi(os.Args[3])
				}
			}
		}
	}
}

// tests
func tests() {
	// active connect to application
	activeTest()
	// test for database
	util.DBTest()
	msg.TestMessagePasser()
}

func initMessagePasser() {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Println("error getting interfaces: ", err)
		os.Exit(-1)
	}
	
	var addr net.IP
	for i := range addrs {
		ip,_,_ := net.ParseCIDR(addrs[i].String())
		if ip.IsLoopback() || ip.IsInterfaceLocalMulticast() || 
		ip.IsLinkLocalMulticast() || ip.IsLinkLocalUnicast() || 
		ip.IsMulticast() {
			continue
		} else {
			addr = ip;
			break;
		}
	}

	msg.MsgPasser,err = msg.NewMsgPasser(addr.String(), ListenPortPeer)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	
	/* register handlers for all the types of messages */
	msg.Handlers[msg.STRING].Encode = msg.SendString
	msg.Handlers[msg.STRING].Decode = msg.RcvString
}