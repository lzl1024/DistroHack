package main

import (
	"encoding/gob"
	"fmt"
	"msg"
	"net"
	"os"
	"strconv"
	"util"
)

var ListenPortLocal = ":4213"
var ListenPortPeer = 4214

var ListenPortSuperNode = 4214

// TODO: change it!
var SNListenPort = 4214

const rcvBufLen = 1024

var isSN = true

//var isSN = false

func main() {
	gob.Register(msg.Message{})
	gob.Register(msg.MultiCastMessage{})

	parseArguments()
	// open database
	util.DatabaseInit()

	initMessagePasser()
	go InitListenerForPeers()

	// tests
	//tests()

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

	msg.MsgPasser, err = msg.NewMsgPasser(addr.String(), ListenPortPeer,
		SNListenPort)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	/* register handlers for all the types of messages */
	msg.Handlers[msg.SN_ONSIGNUP] = msg.RcvSnSignUp
	msg.Handlers[msg.SN_ONSIGNIN] = msg.RcvSnSignIn
	msg.Handlers[msg.SN_ASKINFO] = msg.RcvSnAskInfo
	msg.Handlers[msg.SN_RANK] = msg.RcvSnRank
	msg.Handlers[msg.SN_PBLSUCCESS] = msg.RcvPblSuccess
	msg.Handlers[msg.SN_NODEJOIN] = msg.RcvNodeJoin

	msg.Handlers[msg.STRING] = msg.RcvString
	msg.Handlers[msg.PBLSUCCESS] = msg.RcvPblSuccess
	msg.Handlers[msg.SIGNINACK] = msg.RcvSignInAck
	msg.Handlers[msg.SIGNUPACK] = msg.RcvSignUpAck
	msg.Handlers[msg.ASKINFOACK] = msg.RcvAskInfoAck

	msg.Handlers[msg.STARTEND_SN] = msg.RcvStartEnd_SN
	msg.Handlers[msg.STARTEND_ON] = msg.RcvStartEnd_ON
}
