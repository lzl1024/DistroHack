package main

import (
	"encoding/gob"
	"fmt"
	"msg"
	"net"
	"strconv"
	"util"
	"os"
)

const rcvBufLen = 1024

var isSN = true

func main() {
	gob.Register(msg.Message{})
	gob.Register(msg.MultiCastMessage{})
	
	msg.ReadConfig()
	parseArguments()
	// open database
	util.DatabaseInit()

	initMessagePasser()
	/*if isSN {
		msg.BootStrapSN()
	}*/

	go InitListenerForPeers()

	tests()
	// open the listen port for local app
	listenerLocal, errLocal := net.Listen("tcp", fmt.Sprint(":", msg.ListenPortLocal))

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
				msg.ListenPortLocal,_ = strconv.Atoi(os.Args[2])
				if argLen > 3 {
					msg.ListenPortPeer, _ = strconv.Atoi(os.Args[3])
				}
			}
		}
	}
}


func tests() {
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

	msg.MsgPasser, err = msg.NewMsgPasser(addr.String(), msg.ListenPortPeer,
		msg.ListenPortSuperNode)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	/* register handlers for all the types of messages */
	msg.Handlers[msg.STRING] = msg.RcvString
	
	// SN to SN
	msg.Handlers[msg.SN_SN_SIGNUP] = msg.RcvSnMSignUp		
	msg.Handlers[msg.SN_SN_STARTEND] = msg.RcvSnStartEndFromSN
	msg.Handlers[msg.SN_SN_RANK] = msg.RcvSnRankfromOrigin
	
	// SN to ON
	msg.Handlers[msg.SN_ON_SIGNIN_ACK] = msg.RcvSignInAck
	msg.Handlers[msg.SN_ON_ASKINFO_ACK] = msg.RcvAskInfoAck
	msg.Handlers[msg.SN_ON_SIGNUP_ACK] = msg.RcvSignUpAck
	msg.Handlers[msg.SN_ON_STARTEND] = msg.RcvStartEnd
	msg.Handlers[msg.SN_ON_RANK] = msg.RcvSnRankfromSN
	
	// ON to SN
	msg.Handlers[msg.ON_SN_SIGNUP] = msg.RcvSnSignUp
	msg.Handlers[msg.ON_SN_SIGNIN] = msg.RcvSnSignIn
	msg.Handlers[msg.ON_SN_PBLSUCCESS] = msg.RcvSnPblSuccess
	msg.Handlers[msg.ON_SN_ASKINFO] = msg.RcvSnAskInfo
	msg.Handlers[msg.ON_SN_STARTEND] = msg.RcvSnStartEndFromON
	
	// TO BE IMPLE
	msg.Handlers[msg.SN_NODEJOIN] = msg.RcvNodeJoin
	msg.Handlers[msg.SN_JOIN] = msg.RcvSnJoin	
}
