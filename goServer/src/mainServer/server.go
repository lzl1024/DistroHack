package main

import (
	"encoding/gob"
	"fmt"
	"msg"
	"net"
	"os"
	"util"
)

const rcvBufLen = 1024

var ipArg = ""

func main() {
	gob.Register(msg.Message{})
	gob.Register(msg.MultiCastMessage{})

	parseArguments()
	msg.ReadConfig()
	fmt.Println("Dial Configuration list done")

	/* Initialize single distributed lock */
	msg.DLock = new(msg.DsLock)
	msg.DLock.Init()

	util.DatabaseInit(msg.IsSN)
	fmt.Println("Datavase Initiate")

	initMessagePasser()
	fmt.Println("Message Passer initiated")
	go InitListenerForPeers()
	go msg.DoBootStrap()
	InitListenerLocal()
}

// parse the go argument isSN ipAddress
func parseArguments() {
	argLen := len(os.Args)

	if argLen > 1 {
		if os.Args[1] == "True" {
			msg.IsSN = true
		}

		if argLen > 2 {
			ipArg = os.Args[2]
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
			fmt.Println("Local IP address: ", addrs[i])
			addr = ip
			break
		}
	}

	if ipArg == "" {
		ipArg = addr.String()
	}

	msg.MsgPasser, err = msg.NewMsgPasser(ipArg, msg.ListenPortPeer,
		msg.ListenPortSuperNode)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	// add itself to its ON list
	msg.MsgPasser.ONHostlist[msg.MsgPasser.ServerIP] = msg.MsgPasser.ServerIP
	if msg.IsSN {
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
	msg.Handlers[msg.SN_SN_JOIN_ACK] = msg.RcvSnJoinAck
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
	msg.Handlers[msg.SN_ON_CHANGEONLIST] = msg.RcvSNChangeONList

	// ON to SN
	msg.Handlers[msg.ON_SN_SIGNUP] = msg.RcvSnSignUp
	msg.Handlers[msg.ON_SN_SIGNIN] = msg.RcvSnSignIn
	msg.Handlers[msg.ON_SN_PBLSUCCESS] = msg.RcvSnPblSuccess
	msg.Handlers[msg.ON_SN_ASKINFO] = msg.RcvSnAskInfo
	msg.Handlers[msg.ON_SN_STARTEND] = msg.RcvSnStartEndFromON
	msg.Handlers[msg.ON_SN_JOIN] = msg.RcvOnJoin
	msg.Handlers[msg.ON_SN_REGISTER] = msg.RcvSnOnRegister

	// ON to ON: SN election
	msg.Handlers[msg.ON_ON_ELECTION] = msg.RcvONElection
	msg.Handlers[msg.ON_ON_LEADER] = msg.RcvONLeader
	msg.Handlers[msg.ON_ON_ELECTION_ACK] = msg.RcvElectionACK

	// DS_LOCK
	msg.Handlers[msg.SN_SNLOCKREQ] = msg.RcvSnLockReq
	msg.Handlers[msg.SN_SNLOCKREL] = msg.RcvSnLockRel
	msg.Handlers[msg.SN_SNLOCKACK] = msg.RcvSnLockAck
	msg.Handlers[msg.SN_SNACKRENEG] = msg.RcvSnAckReneg

}
