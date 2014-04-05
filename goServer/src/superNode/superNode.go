package superNode

import (
	"encoding/gob"
	"fmt"
	"msg"
	"net"
	//"reflect"
	"sync"
	//"time"
	"container/list"
	"util"
)

// TODO superNode Thread,
// function 1: when receiving supernode "global_ranking" change, merge lists, update and send to all its peers
// function 2: when receiving on "submit success", update global ranking list and send to its peers and other SN if needed
// function 3: when recieve "sign_in" from ON check available SN and send msg to it, to let him register the on or send back sign in fail.
// function 4: connect with an ON and send all msg. userlist, local_info, global ranking.

type localInfo struct {
	rankList [msg.GlobalRankSize]msg.UserRecord
	scoreMap map[string]msg.UserRecord
}

var mu sync.Mutex
var mp = msg.MsgPasser
var rankList [msg.GlobalRankSize]msg.UserRecord
var scoreMap map[string]msg.UserRecord

var groupList *list.List
var snList *list.List

var Global_ranking = []msg.UserRecord{}
var Local_info = map[string]msg.UserRecord{}

func SuperNodeThreadTest() {
	scoreMap = make(map[string]msg.UserRecord)

	tmpLocalInfo := new(localInfo)

	fmt.Printf("%p\n", &tmpLocalInfo.scoreMap)

	tmpLocalInfo.rankList = rankList
	record := new(msg.UserRecord)
	record.NewUserRecord("aaa", 1, rankList[0].Ctime)
	scoreMap["aaa"] = *record
	tmpLocalInfo.scoreMap = scoreMap

	fmt.Printf("%p\n", &scoreMap)
	fmt.Printf("%p\n", &tmpLocalInfo.scoreMap)
	fmt.Println(tmpLocalInfo.scoreMap["aaa"].String())

	/*record := new(msg.UserRecord)
	fmt.Println(reflect.TypeOf(record))
	record.NewUserRecord("aaa", 1, rankList[0].Ctime)
	fmt.Printf("%p\n", record)

	record1 := *record
	fmt.Printf("%p\n", &record1)
	record1.Score = 2

	rankList[0] = record1
	fmt.Printf("%p\n", &rankList[0])
	rankList[1] = *record
	fmt.Printf("%p\n", &rankList[1])

	var tmprankList [msg.GlobalRankSize]msg.UserRecord

	record1.Score = 3
	tmprankList[0] = record1
	tmprankList[1] = *record

	updateGlobalRankList(tmprankList)*/
}

func SuperNodeThread(serverPort int) {
	channel := make(chan error)
	go superNodeListenThread(serverPort, channel)
	value := <-channel
	fmt.Println(value)
}

func superNodeListenThread(serverPort int, c chan error) {
	// Get the list of other super node
	parseConfigFile()

	// Get the tcpAddr
	fmt.Println("SuperNode: Started superNode server thread")
	service := fmt.Sprint(":", serverPort)
	tcpAddr, err := net.ResolveTCPAddr("tcp", service)
	if err != nil {
		fmt.Println("SuperNode: Unrecoverable error trying to start go supernode server")
		c <- err
		return
	}

	// Create Listener
	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		fmt.Println("SuperNode: Unrecoverable error trying to start listening on server ", err)
		c <- err
		return
	}

	// Create new receive thread when a new msg arrives
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("SuperNode: error accepting connection...continuing")
			continue
		}
		go superNodeRcvThread(mp, conn)
	}

}

func superNodeRcvThread(mp *msg.Messagepasser, conn net.Conn) {
	fmt.Println("SuperNode: Started super node recevier thread\n")
	var tcpconn *net.TCPConn
	var ok bool
	var err error
	var msg msg.Message

	tcpconn, ok = conn.(*net.TCPConn)
	if ok {
		err = tcpconn.SetLinger(0)
		if err != nil {
			fmt.Println("SuperNode: cannot set linger options")
		}
	}

	decoder := gob.NewDecoder(conn)
	for {
		err := decoder.Decode(&msg)
		if err != nil {
			fmt.Println("SuperNode: error while decoding: ", err)
			conn.Close()
			break
		}
		superNodeMsgDoAction(&msg)
	}
}

func superNodeMsgDoAction(m *msg.Message) {
	data, err := msg.Handlers[m.Kind](m)
	if err != nil {
		return
	}
	fmt.Println((*m).String())

	switch m.Kind {
	case msg.SN_RANK:
		rcvGlobalRankList(m, data.([msg.GlobalRankSize]msg.UserRecord))
	case msg.SN_PBLSUCCESS:
		rcvPblSuccessFromON(data.(msg.UserRecord))
	case msg.SN_SIGNIN:
		rcvSignInFromON(m, data.(map[string]string))
	case msg.SN_NODEJOIN:
		rcvConnectWithON(m)
	}
}

func rcvGlobalRankList(m *msg.Message, tmpRankList [msg.GlobalRankSize]msg.UserRecord) {
	// Update global rank list
	mu.Lock()

	//update the global rank list in local
	rankList = tmpRankList

	// Update the local info map
	for _, userRecord := range tmpRankList {
		if _, present := scoreMap[userRecord.UserName]; present {
			if scoreMap[userRecord.UserName].Ctime.Before(userRecord.Ctime) {
				scoreMap[userRecord.UserName] = userRecord
			}
		}
	}
	mu.Unlock()

	//multicast the global rank in group
	multicastMsgInGroup(m)
}

func rcvPblSuccessFromON(userRecord msg.UserRecord) {
	//Update the local Info
	globalRankChanged := updateLocalInfoWithOneRecord(userRecord)

	//multicast the new grobal rank to Sns
	if globalRankChanged {
		multicastGlobalRankToSNs()
	}
}

func rcvSignInFromON(m *msg.Message, signInMsg map[string]string) {
	// check database and send back SignInAck
	// TODO: update number of node register, send by heartbeat

	backMsg := util.DatabaseSignIn(signInMsg["username"], signInMsg["password"])

	backData := map[string]string{
		"user":   signInMsg["username"],
		"status": backMsg,
	}

	sendoutMsg := new(msg.Message)
	err := sendoutMsg.NewMsgwithData(m.Src, msg.SIGNINACK, backData)
	if err != nil {
		fmt.Println(err)
	}

	// send message to SN
	mp.Send(sendoutMsg, false)
}

func rcvConnectWithON(m *msg.Message) {
	userRecord := new(msg.UserRecord)
	userRecord.UserName = m.Src

	globalRankChanged := updateLocalInfoWithOneRecord(*userRecord)

	//Multicast the userRecord
	multicastMsgInGroup(m)

	//multicast global rank to SNs
	if globalRankChanged {
		multicastGlobalRankToSNs()
	}

	//unicast local info to the new node
	unicastLocalInfoToNode(m.Src)
}

func updateLocalInfoWithOneRecord(userRecord msg.UserRecord) bool {
	mu.Lock()

	if _, present := scoreMap[userRecord.UserName]; present {
		if scoreMap[userRecord.UserName].Ctime.After(userRecord.Ctime) {
			mu.Unlock()
			return false
		}
		scoreMap[userRecord.UserName] = userRecord
	} else {
		fmt.Println("SuperNode Error: New ON Put in Local List")
		scoreMap[userRecord.UserName] = userRecord
		mu.Unlock()
		return false
	}

	//Update the Global Rank
	i := msg.GlobalRankSize - 1
	if i >= 0 && (userRecord.Score > rankList[i].Score || len(rankList[i].UserName) == 0) {
		tmpRank := rankList[i]
		rankList[i] = userRecord
		if i < msg.GlobalRankSize-1 {
			rankList[i+1] = tmpRank
		}
		i--
	}
	mu.Unlock()

	return true

}

func multicastMsgInGroup(m *msg.Message) {
	newMCastMsg := new(msg.MultiCastMessage)
	tmpMsg := &newMCastMsg.Message
	tmpMsg.CopyMsg(m)
	newMCastMsg.Origin = mp.ServerIP
	mp.SendMCast(newMCastMsg, groupList)
}

func multicastGlobalRankToSNs() {
	newMCastMsg := new(msg.MultiCastMessage)
	newMCastMsg.NewMCastMsgwithData("", msg.SN_RANK, rankList)
	newMCastMsg.Origin = mp.ServerIP
	mp.SendMCast(newMCastMsg, snList)
}

func unicastLocalInfoToNode(dest string) {
	tmpLocalInfo := new(localInfo)
	tmpLocalInfo.rankList = rankList
	tmpLocalInfo.scoreMap = scoreMap

	message := new(msg.Message)
	message.NewMsgwithData(dest, msg.GROUPINFO, *tmpLocalInfo)
	mp.Send(message, false)
}

func parseConfigFile() {

}
