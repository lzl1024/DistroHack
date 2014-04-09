package msg

import (
	//"container/list"
	"errors"
	"fmt"
	"sync"
	"util"
)

// TODO superNode Thread,
// function 1: when receiving supernode "global_ranking" change, merge lists, update and send to all its peers
// function 2: when receiving on "submit success", update global ranking list and send to its peers and other SN if needed
// function 3: when recieve "sign_in" from ON check available SN and send msg to it, to let him register the on or send back sign in fail.
// function 4: connect with an ON and send all msg. userlist, local_info, global ranking.

type LocalInfo struct {
	Ranklist [GlobalRankSize]UserRecord
	Scoremap map[string]UserRecord
}

var mu sync.Mutex
var rankList [GlobalRankSize]UserRecord
var scoreMap map[string]UserRecord = make(map[string]UserRecord)

func SuperNodeThreadTest() {
	/*scoreMap = make(map[string]UserRecord)

	tmpLocalInfo := new(localInfo)

	fmt.Printf("%p\n", &tmpLocalInfo.scoremap)

	tmpLocalInfo.ranklist = rankList
	record := new(UserRecord)
	record.NewUserRecord("aaa", 1, rankList[0].Ctime)
	scoreMap["aaa"] = *record
	tmpLocalInfo.scoremap = scoreMap

	fmt.Printf("%p\n", &scoreMap)
	fmt.Printf("%p\n", &tmpLocalInfo.scoremap)
	fmt.Println(tmpLocalInfo.scoremap["aaa"].String())

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

func RcvSnSignUp(msg *Message) (interface{}, error) {
	// register user and send back SignUpAck
	if msg.Kind != SN_ONSIGNUP {
		return nil, errors.New("message Kind indicates not a SN_ONSIGNUP")
	}

	var signUpMsg map[string]string
	err := ParseRcvInterfaces(msg, &signUpMsg)
	if err != nil {
		return nil, err
	}

	backMsg := util.DatabaseSignUp(signUpMsg["username"], signUpMsg["password"], signUpMsg["email"])

	backData := map[string]string{
		"user":     signUpMsg["username"],
		"status":   backMsg,
		"question": "url",
	}

	fmt.Printf("SuperNode: ordinary sign up %s,  status %s\n", signUpMsg["username"], backMsg)

	if backMsg == "success" {
		userRecord := new(UserRecord)
		userRecord.UserName = signUpMsg["username"]
		updateLocalInfoWithOneRecord(*userRecord)
	}

	//Multicast the new user to other supernodes
	sendCastMsg := new(Message)
	sendCastMsg.CopyMsg(msg)
	sendCastMsg.Kind = SN_MSIGNUP
	multicastMsgInGroup(sendCastMsg, true)

	sendoutMsg := new(Message)
	err = sendoutMsg.NewMsgwithData(msg.Src, SIGNUPACK, backData)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	// send message to ON
	MsgPasser.Send(sendoutMsg)

	return signUpMsg, err
}

func RcvSnMSignUp(msg *Message) (interface{}, error) {
	// register user and send back SignUpAck
	if msg.Kind != SN_MSIGNUP {
		return nil, errors.New("message Kind indicates not a SN_MSIGNUP")
	}

	var mSignUpMsg map[string]string
	err := ParseRcvInterfaces(msg, &mSignUpMsg)
	if err != nil {
		return nil, err
	}

	backMsg := util.DatabaseSignUp(mSignUpMsg["username"], mSignUpMsg["password"], mSignUpMsg["email"])

	fmt.Println("SuperNodeandlers RcvSnMSignUp: backMsg ", backMsg)

	return msg, err
}

func RcvSnSignIn(msg *Message) (interface{}, error) {
	if msg.Kind != SN_ONSIGNIN {
		return nil, errors.New("message Kind indicates not a SN_ONSIGNIN")
	}

	var signInMsg map[string]string
	err := ParseRcvInterfaces(msg, &signInMsg)
	if err != nil {
		return nil, err
	}

	backMsg := util.DatabaseSignIn(signInMsg["username"], signInMsg["password"])

	backData := map[string]string{
		"user":   signInMsg["username"],
		"status": backMsg,
	}

	fmt.Printf("SuperNode: ordinary sign in %s,  status %s\n", signInMsg["username"], backMsg)

	if backMsg == "success" {
		userRecord := new(UserRecord)
		userRecord.UserName = signInMsg["username"]
		updateLocalInfoWithOneRecord(*userRecord)
	}

	sendoutMsg := new(Message)
	err = sendoutMsg.NewMsgwithData(msg.Src, SIGNINACK, backData)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	// send message to SN
	MsgPasser.Send(sendoutMsg)

	return signInMsg, err
}

// Received in SN
func RcvPblSuccess(msg *Message) (interface{}, error) {
	if msg.Kind != SN_PBLSUCCESS {
		return nil, errors.New("message Kind indicates not a PBLSUCCESS")
	}

	// TODO: for SN, it should merge to global ranking, and send back if needed
	var userRecord UserRecord
	if err := ParseRcvInterfaces(msg, &userRecord); err != nil {
		return nil, err
	}

	globalRankChanged := updateLocalInfoWithOneRecord(userRecord)

	//multicast the new grobal rank to Sns
	if globalRankChanged {
		multicastGlobalRankToSNs()
	}

	return userRecord, nil
}

func RcvSnRank(msg *Message) (interface{}, error) {
	if msg.Kind != SN_RANK {
		return nil, errors.New("message Kind indicates not a PBLSUCCESS")
	}

	// TODO: for SN, it should merge to global ranking, and send back if needed
	var newRankList [GlobalRankSize]UserRecord
	if err := ParseRcvInterfaces(msg, &newRankList); err != nil {
		return nil, err
	}

	mu.Lock()

	fmt.Println("SuperNodeReceiver: SN_RANK Message Received")

	//update the global rank list in local
	rankList = newRankList

	// Update the local info map
	for _, userRecord := range newRankList {
		if _, present := scoreMap[userRecord.UserName]; present {
			if scoreMap[userRecord.UserName].Ctime.Before(userRecord.Ctime) {
				scoreMap[userRecord.UserName] = userRecord
			}
		}
	}
	mu.Unlock()

	//multicast the global rank in group
	multicastMsgInGroup(msg, false)

	return newRankList, nil
}

// Request from ordinary node to ask info from supernode
func RcvSnAskInfo(msg *Message) (interface{}, error) {
	if msg.Kind != SN_ASKINFO {
		return nil, errors.New("message Kind indicates not a SN_ASKINFO")
	}

	backData := new(LocalInfo)
	backData.Ranklist = rankList
	backData.Scoremap = make(map[string]UserRecord)

	mu.Lock()
	for k, v := range scoreMap {
		backData.Scoremap[k] = v
	}
	mu.Unlock()

	sendoutMsg := new(Message)
	err := sendoutMsg.NewMsgwithData(msg.Src, ASKINFOACK, *backData)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	// send message to SN
	MsgPasser.Send(sendoutMsg)

	return nil, nil
}

func RcvSnStartEndFromON(msg *Message) (interface{}, error) {
	if msg.Kind != SN_STARTENDON {
		return nil, errors.New("message Kind indicates not a SN_STARTENDON")
	}

	newMessage := new(Message)
	newMessage.CopyMsg(msg)
	newMessage.Kind = SN_STARTEND

	multicastMsgInGroup(newMessage, true)

	return nil, nil
}

func RcvSnStartEnd(msg *Message) (interface{}, error) {
	if msg.Kind != SN_STARTEND {
		return nil, errors.New("message Kind indicates not a SN_STARTEND")
	}

	newMessage := new(Message)
	newMessage.CopyMsg(msg)
	newMessage.Kind = STARTEND

	multicastMsgInGroup(newMessage, false)

	return nil, nil
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

	var tmpUserRecord UserRecord
	replaced := false
	for i := 0; i < GlobalRankSize; i++ {
		if !replaced {
			if userRecord.UserName != rankList[i].UserName {
				if rankList[i].CompareTo(userRecord) {
					continue
				} else {
					tmpUserRecord = rankList[i]
					rankList[i] = userRecord
					replaced = true
					if len(tmpUserRecord.UserName) == 0 {
						break
					}
				}
			} else {
				if userRecord.CompareTo(rankList[i]) {
					rankList[i] = userRecord
				}
				break
			}
		} else {
			if userRecord.UserName != rankList[i].UserName {
				tmpUserRecord1 := rankList[i]
				rankList[i] = tmpUserRecord
				tmpUserRecord = tmpUserRecord1
				if len(tmpUserRecord.UserName) == 0 {
					break
				}
			} else {
				rankList[i] = tmpUserRecord
				break
			}
		}
	}

	for i := range rankList {
		fmt.Println(rankList[i].String())
	}

	mu.Unlock()
	return rankChanged
}

func multicastMsgInGroup(m *Message, isSuper bool) {
	newMCastMsg := new(MultiCastMessage)
	tmpMsg := &newMCastMsg.Message
	tmpMsg.CopyMsg(m)
	newMCastMsg.Origin = MsgPasser.ServerIP

	newMCastMsg.HostList = make([]string, 0)

	if isSuper {
		for e := MsgPasser.SNHostlist.Front(); e != nil; e = e.Next() {
			newMCastMsg.HostList = append(newMCastMsg.HostList, e.Value.(string))
		}

	} else {
		for e := MsgPasser.ONHostlist.Front(); e != nil; e = e.Next() {
			newMCastMsg.HostList = append(newMCastMsg.HostList, e.Value.(string))
		}
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

func parseConfigFile() {

}
