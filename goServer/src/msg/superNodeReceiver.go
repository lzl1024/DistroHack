package msg

import (
	"container/list"
	"fmt"
	"sync"
	"util"
)

// TODO superNode Thread,
// function 1: when receiving supernode "global_ranking" change, merge lists, update and send to all its peers
// function 2: when receiving on "submit success", update global ranking list and send to its peers and other SN if needed
// function 3: when recieve "sign_in" from ON check available SN and send msg to it, to let him register the on or send back sign in fail.
// function 4: connect with an ON and send all msg. userlist, local_info, global ranking.

type localInfo struct {
	rankList [GlobalRankSize]UserRecord
	scoreMap map[string]UserRecord
}

var mu sync.Mutex
var rankList [GlobalRankSize]UserRecord
var scoreMap map[string]UserRecord = make(map[string]UserRecord)

func SuperNodeThreadTest() {
	scoreMap = make(map[string]UserRecord)

	tmpLocalInfo := new(localInfo)

	fmt.Printf("%p\n", &tmpLocalInfo.scoreMap)

	tmpLocalInfo.rankList = rankList
	record := new(UserRecord)
	record.NewUserRecord("aaa", 1, rankList[0].Ctime)
	scoreMap["aaa"] = *record
	tmpLocalInfo.scoreMap = scoreMap

	fmt.Printf("%p\n", &scoreMap)
	fmt.Printf("%p\n", &tmpLocalInfo.scoreMap)
	fmt.Println(tmpLocalInfo.scoreMap["aaa"].String())

	a := list.New()
	fmt.Printf("%p\n", a)

	b := a
	fmt.Printf("%p\n", b)

	newMCastMsg := new(MultiCastMessage)
	//newMCastMsg.NewMCastMsgwithData("", msg.SN_RANK, rankList)
	//newMCastMsg.Origin = mp.ServerIP
	fmt.Printf("%p  %p\n ", &newMCastMsg.HostList, newMCastMsg.HostList)
	fmt.Println(newMCastMsg.HostList[0])
	newMCastMsg.HostList = make([]string, 0)
	fmt.Printf("%p %p\n", &newMCastMsg.HostList, newMCastMsg.HostList)

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

func SuperNodeMsgDoAction(m *Message) {
	data, err := Handlers[m.Kind](m)
	if err != nil {
		return
	}

	switch m.Kind {
	case SN_ONSIGNUP:
		rcvSignUpFromON(m, data.(map[string]string))
	case SN_ONSIGNIN:
		rcvSignInFromON(m, data.(map[string]string))
	case SN_RANK:
		rcvGlobalRankList(m, data.([GlobalRankSize]UserRecord))
	case SN_PBLSUCCESS:
		rcvPblSuccessFromON(data.(UserRecord))
	case SN_NODEJOIN:
		rcvConnectWithON(m)
	}
}

func rcvSignUpFromON(m *Message, signUpMsg map[string]string) {
	// register user and send back SignUpAck

	backMsg := util.DatabaseSignUp(signUpMsg["username"], signUpMsg["password"], signUpMsg["email"])

	backData := map[string]string{
		"user":   signUpMsg["username"],
		"status": backMsg,
	}

	fmt.Printf("SuperNode: ordinary sign up %s,  status %s\n", signUpMsg["username"], backMsg)

	sendoutMsg := new(Message)
	err := sendoutMsg.NewMsgwithData(m.Src, SIGNUPACK, backData)
	if err != nil {
		fmt.Println(err)
	}

	// send message to SN
	MsgPasser.Send(sendoutMsg)
}

func rcvSignInFromON(m *Message, signInMsg map[string]string) {
	// check database and send back SignInAck
	// TODO: update number of node register, send by heartbeat

	backMsg := util.DatabaseSignIn(signInMsg["username"], signInMsg["password"])

	backData := map[string]string{
		"user":   signInMsg["username"],
		"status": backMsg,
	}

	fmt.Printf("SuperNode: ordinary sign in %s,  status %s\n", signInMsg["username"], backMsg)

	sendoutMsg := new(Message)
	err := sendoutMsg.NewMsgwithData(m.Src, SIGNINACK, backData)
	if err != nil {
		fmt.Println(err)
	}

	// send message to SN
	MsgPasser.Send(sendoutMsg)
}

func rcvGlobalRankList(m *Message, tmpRankList [GlobalRankSize]UserRecord) {
	// Update global rank list
	mu.Lock()

	fmt.Println("SuperNodeReceiver: SN_RANK Message Received")

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

func rcvPblSuccessFromON(userRecord UserRecord) {
	//Update the local Info
	globalRankChanged := updateLocalInfoWithOneRecord(userRecord)

	//multicast the new grobal rank to Sns
	if globalRankChanged {
		multicastGlobalRankToSNs()
	}
}

func rcvConnectWithON(m *Message) {
	userRecord := new(UserRecord)
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

func updateLocalInfoWithOneRecord(userRecord UserRecord) bool {
	mu.Lock()

	rankChanged := false

	if _, present := scoreMap[userRecord.UserName]; present {
		if scoreMap[userRecord.UserName].Ctime.After(userRecord.Ctime) {
			mu.Unlock()
			return rankChanged
		}
		scoreMap[userRecord.UserName] = userRecord
	} else {
		fmt.Println("SuperNode Info: New ON Put in Local List")
		scoreMap[userRecord.UserName] = userRecord
	}

	//Update the Global Rank
	for i := GlobalRankSize - 1; i >= 0 && (userRecord.Score > rankList[i].Score || len(rankList[i].UserName) == 0); i-- {
		rankChanged = true
		tmpRank := rankList[i]
		rankList[i] = userRecord
		if i < GlobalRankSize-1 {
			rankList[i+1] = tmpRank
		}
	}

	for i := range rankList {
		fmt.Println(rankList[i].String())
	}

	mu.Unlock()
	return rankChanged
}

func multicastMsgInGroup(m *Message) {
	newMCastMsg := new(MultiCastMessage)
	tmpMsg := &newMCastMsg.Message
	tmpMsg.CopyMsg(m)
	newMCastMsg.Origin = MsgPasser.ServerIP

	newMCastMsg.HostList = make([]string, 0)
	for e := MsgPasser.ONHostlist.Front(); e != nil; e = e.Next() {
		newMCastMsg.HostList = append(newMCastMsg.HostList, e.Value.(string))
	}
	MsgPasser.SendMCast(newMCastMsg)
}

func multicastGlobalRankToSNs() {
	newMCastMsg := new(MultiCastMessage)
	newMCastMsg.NewMCastMsgwithData("", SN_RANK, rankList)
	newMCastMsg.Origin = MsgPasser.ServerIP

	newMCastMsg.HostList = make([]string, 0)
	for e := MsgPasser.SNHostlist.Front(); e != nil; e = e.Next() {
		newMCastMsg.HostList = append(newMCastMsg.HostList, e.Value.(string))
	}
	MsgPasser.SendMCast(newMCastMsg)
}

func unicastLocalInfoToNode(dest string) {
	tmpLocalInfo := new(localInfo)
	tmpLocalInfo.rankList = rankList
	tmpLocalInfo.scoreMap = scoreMap

	message := new(Message)
	message.NewMsgwithData(dest, GROUPINFO, *tmpLocalInfo)
	MsgPasser.Send(message)
}

func parseConfigFile() {

}
