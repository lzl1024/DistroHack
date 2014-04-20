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

var isSN = false

func main() {
	gob.Register(msg.Message{})
	gob.Register(msg.MultiCastMessage{})

	parseArguments()
	msg.ReadConfig()

	/* Initialize single distributed lock */
	msg.DLock = new(msg.DsLock)
	msg.DLock.Init()
	
	util.DatabaseInit(isSN)

	initMessagePasser()
	if isSN {
		msg.BootStrapSN()
	} else {
		msg.BootStrapON()
	}

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
		fmt.Println(addrs[i])
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
	
	// add itself to its ON list 
	msg.MsgPasser.ONHostlist[msg.MsgPasser.ServerIP] = msg.MsgPasser.ServerIP
	if isSN {
		msg.MsgPasser.SNHostlist[msg.MsgPasser.ServerIP] = msg.MsgPasser.ServerIP
		msg.MsgPasser.SNLoadlist[msg.MsgPasser.ServerIP] = len(msg.MsgPasser.ONHostlist)
	}

	/* register handlers for all the types of messages */
	msg.Handlers[msg.STRING] = msg.RcvString
	
	// SN to SN
	msg.Handlers[msg.SN_SN_SIGNUP] = msg.RcvSnMSignUp		
	msg.Handlers[msg.SN_SN_STARTEND] = msg.RcvSnStartEndFromSN
	msg.Handlers[msg.SN_SN_RANK] = msg.RcvSnRankfromOrigin
	msg.Handlers[msg.SN_SN_COMMIT_RD] = msg.RcvSnSignUpCommitReady
	msg.Handlers[msg.SN_SN_COMMIT_RD_ACK] = msg.RcvSnSignUpCommitReadyACK
	msg.Handlers[msg.SN_SN_JOIN] = msg.RcvSnJoin
	msg.Handlers[msg.SN_SN_LOADUPDATE] = msg.RcvSnLoadUpdate
	msg.Handlers[msg.SN_SN_LOADMERGE] = msg.RcvSnLoadMerge
	msg.Handlers[msg.SN_SN_LISTMERGE] = msg.RcvSnListMerge
	msg.Handlers[msg.SN_SN_LISTUPDATE] = msg.RcvSnListUpdate
	
	// SN to ON
	msg.Handlers[msg.SN_ON_SIGNIN_ACK] = msg.RcvSignInAck
	msg.Handlers[msg.SN_ON_ASKINFO_ACK] = msg.RcvAskInfoAck
	msg.Handlers[msg.SN_ON_SIGNUP_ACK] = msg.RcvSignUpAck
	msg.Handlers[msg.SN_ON_STARTEND] = msg.RcvStartEnd
	msg.Handlers[msg.SN_ON_RANK] = msg.RcvSnRankfromSN
	msg.Handlers[msg.SN_ON_JOIN_ACK] = msg.RcvOnJoinAck
	
	// ON to SN
	msg.Handlers[msg.ON_SN_SIGNUP] = msg.RcvSnSignUp
	msg.Handlers[msg.ON_SN_SIGNIN] = msg.RcvSnSignIn
	msg.Handlers[msg.ON_SN_PBLSUCCESS] = msg.RcvSnPblSuccess
	msg.Handlers[msg.ON_SN_ASKINFO] = msg.RcvSnAskInfo
	msg.Handlers[msg.ON_SN_STARTEND] = msg.RcvSnStartEndFromON
	msg.Handlers[msg.ON_SN_JOIN] = msg.RcvOnJoin
	msg.Handlers[msg.ON_SN_REGISTER] = msg.RcvSnOnRegister
	
	// DS_LOCK
	msg.Handlers[msg.SN_SNLOCKREQ] = msg.RcvSnLockReq
	msg.Handlers[msg.SN_SNLOCKREL] = msg.RcvSnLockRel
	msg.Handlers[msg.SN_SNLOCKACK] = msg.RcvSnLockAck
	msg.Handlers[msg.SN_SNACKRENEG] = msg.RcvSnAckReneg
	
}
