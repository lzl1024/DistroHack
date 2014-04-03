package main

import (
	"fmt"
	"msg"
	"net"
	"os"
	"strconv"
	"superNode"
	"util"
)

var ListenPortLocal = ":4213"
var ListenPortPeer = 4214
var ListenPortSuperNode = 4215

const rcvBufLen = 1024

var isSN = false

func main() {

	//parseArguments()

	// open database
	//util.DatabaseInit()

	//initMessagePasser()
	//go InitConnectionFromPeers()

	// open SN port when is needed
	//if isSN == true {
	//go superNode.SuperNodeThread(ListenPortSuperNode)
	superNode.SuperNodeThread(ListenPortSuperNode)
	//}

	//fmt.Print("aaa")

	// tests
	//tests()

	// open the listen port for local app
	//listenerLocal, errLocal := net.Listen("tcp", ListenPortLocal)

	/*if errLocal != nil {
		fmt.Println("Server: Listener port has been used:", errLocal.Error())
		return
	}*/
	// main routine: commmunication between server and app
	//HandleConnectionFromLocal(listenerLocal)
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
	//activeTest()
	// test for database
	//util.DBTest()
	go msg.TestMessagePasser()
}

func initMessagePasser() {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Println("error getting interfaces: ", err)
		os.Exit(-1)
	}

	var addr net.IP
	for i := range addrs {
		ip, _, _ := net.ParseCIDR(addrs[i].String())
		if ip.IsLoopback() || ip.IsInterfaceLocalMulticast() ||
			ip.IsLinkLocalMulticast() || ip.IsLinkLocalUnicast() ||
			ip.IsMulticast() {
			continue
		} else {
			addr = ip
			break
		}
	}

	msg.MsgPasser, err = msg.NewMsgPasser(addr.String(), ListenPortPeer)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	/* register handlers for all the types of messages */
	msg.Handlers[msg.STRING].Encode = msg.SendString
	msg.Handlers[msg.STRING].Decode = msg.RcvString

	msg.Handlers[msg.SN_RANK].Encode = msg.SendSnRank
	msg.Handlers[msg.SN_RANK].Decode = msg.RcvSnRank

	msg.Handlers[msg.SN_ON_SUBMIT].Encode = msg.SendSnOnSubmit
	msg.Handlers[msg.SN_ON_SUBMIT].Decode = msg.RcvSnOnSubmit

	msg.Handlers[msg.SN_ON_SIGNIN].Encode = msg.SendSnOnSignIn
	msg.Handlers[msg.SN_ON_SIGNIN].Decode = msg.RcvSnOnSignIn
}
