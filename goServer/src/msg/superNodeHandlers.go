package msg

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"util"
)


type LocalInfo struct {
	Ranklist [GlobalRankSize]UserRecord
	Scoremap map[string]UserRecord
}

var Local_Info_Mutex sync.Mutex

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
	//newMCastMsg.NewMCastMsgwithData("", msg.SN_SNON_RANK, rankList)
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


// rcv sign up msg from ON
func RcvSnSignUp(msg *Message) (interface{}, error) {
	// register user and send back SignUpAck
	if msg.Kind != ON_SN_SIGNUP {
		return nil, errors.New("message Kind indicates not a ON_SN_SIGNUP")
	}

	var signUpMsg map[string]string
	err := ParseRcvInterfaces(msg, &signUpMsg)
	if err != nil {
		fmt.Println("In RcvSnSignUp:")
		return nil, err
	}

	// TODO: change to 2 phase commit
	backMsg := util.DatabaseSignUp(signUpMsg["username"], signUpMsg["password"], signUpMsg["email"])

	backData := map[string]string{
		"user":     signUpMsg["username"],
		"status":   backMsg,
		"question": "url",
	}

	fmt.Printf("SuperNode: ordinary sign up %s,  status %s\n", signUpMsg["username"], backMsg)

	// if success update local information and multicast messages 
	if backMsg == "success" {
		userRecord := new(UserRecord)
		userRecord.UserName = signUpMsg["username"]
		updateLocalInfoWithOneRecord(*userRecord)
		
		//Multicast the new user to other supernodes
		sendCastMsg := new(Message)
		sendCastMsg.CopyMsg(msg)
		sendCastMsg.Kind = SN_SN_SIGNUP
		multicastMsgInGroup(sendCastMsg, true)
	}

	sendoutMsg := new(Message)
	err = sendoutMsg.NewMsgwithData(msg.Src, SN_ON_SIGNUP_ACK, backData)
	if err != nil {
		fmt.Println("In RcvSnSignUp:")
		return nil, err
	}
	
	// send message to ON
	MsgPasser.Send(sendoutMsg)

	return signUpMsg, err
}


// rcv the multicast sign up message from other SN
func RcvSnMSignUp(msg *Message) (interface{}, error) {
	// register user and send back SignUpAck
	if msg.Kind != SN_SN_SIGNUP {
		return nil, errors.New("message Kind indicates not a SN_SN_SIGNUP")
	}

	var mSignUpMsg map[string]string
	err := ParseRcvInterfaces(msg, &mSignUpMsg)
	if err != nil {
		fmt.Println("In SN_SN_SIGNUP:")
		return nil, err
	}

	backMsg := util.DatabaseSignUp(mSignUpMsg["username"], mSignUpMsg["password"], mSignUpMsg["email"])

	fmt.Println("SuperNodeandlers RcvSnMSignUp: backMsg ", backMsg)

	return mSignUpMsg, err
}


// rcv sign In msg from ON
func RcvSnSignIn(msg *Message) (interface{}, error) {
	if msg.Kind != ON_SN_SIGNIN {
		return nil, errors.New("message Kind indicates not a ON_SN_SIGNIN")
	}

	var signInMsg map[string]string
	err := ParseRcvInterfaces(msg, &signInMsg)
	if err != nil {
		fmt.Println("In RcvSnSignIn")
		return nil, err
	}

	backMsg := util.DatabaseSignIn(signInMsg["username"], signInMsg["password"])

	backData := map[string]string{
		"user":   signInMsg["username"],
		"status": backMsg,
	}

	fmt.Printf("SuperNode: ordinary sign in %s,  status %s\n", signInMsg["username"], backMsg)

	// if success, update local information
	if backMsg == "success" {
		userRecord := new(UserRecord)
		userRecord.UserName = signInMsg["username"]
		updateLocalInfoWithOneRecord(*userRecord)
	}

	// create and send status message
	sendoutMsg := new(Message)
	err = sendoutMsg.NewMsgwithData(msg.Src, SN_ON_SIGNIN_ACK, backData)
	if err != nil {
		fmt.Println("In RcvSnSignIn:")
		return nil, err
	}

	// send message to ON
	MsgPasser.Send(sendoutMsg)

	return signInMsg, err
}


// rcv pbl Success msg from ON
func RcvSnPblSuccess(msg *Message) (interface{}, error) {
	if msg.Kind != ON_SN_PBLSUCCESS {
		return nil, errors.New("message Kind indicates not a ON_SN_PBLSUCCESS")
	}

	var userRecord UserRecord
	if err := ParseRcvInterfaces(msg, &userRecord); err != nil {
		fmt.Println("In RcvSnPblSuccess:")
		return nil, err
	}

	globalRankChanged := updateLocalInfoWithOneRecord(userRecord)
	
	fmt.Println("CHange?, ", globalRankChanged)
	
	//multicast the new grobal rank to SNs
	if globalRankChanged {
		multicastGlobalRankToSNs()
	}

	// create and send new local info to other ONs
	backData := new(LocalInfo)
	backData.Ranklist = Global_ranking
	backData.Scoremap = make(map[string]UserRecord)

	Local_Info_Mutex.Lock()
	for k, v := range Local_map {
		backData.Scoremap[k] = v
	}
	Local_Info_Mutex.Unlock()

	sendoutMsg := new(Message)
	err := sendoutMsg.NewMsgwithData(msg.Src, SN_ON_ASKINFO_ACK, *backData)
	if err != nil {
		fmt.Println("In RcvSnPblSuccess:")
		return nil, err
	}
	
	// send message to ONs
	multicastMsgInGroup(sendoutMsg, false)

	return userRecord, nil
}


// SN or ON rcv ranking update
func RcvSnRankfromOrigin(msg *Message) (interface{}, error) {
	if msg.Kind != SN_SN_RANK {
		return nil, errors.New("message Kind indicates not a SN_SN_RANK")
	}

	var newRankList [GlobalRankSize]UserRecord
	if err := ParseRcvInterfaces(msg, &newRankList); err != nil {
		fmt.Println("In RcvSnRankfromOrigin:")
		return nil, err
	}
	
	Local_Info_Mutex.Lock()

	// TODO: should merge
	//update the global rank list in local
	Global_ranking = newRankList

	// Update the local info map
	/*for _, userRecord := range newRankList {
		if _, present := Local_map[userRecord.UserName]; present {
			if Local_map[userRecord.UserName].Ctime.Before(userRecord.Ctime) {
				Local_map[userRecord.UserName] = userRecord
			}
		}
	}*/
	Local_Info_Mutex.Unlock()

	// multicast the global rank in group
	newMessage := new(Message)
	newMessage.CopyMsg(msg)
	newMessage.Kind = SN_ON_RANK
	multicastMsgInGroup(newMessage, false)

	return newRankList, nil
}


// Request from ordinary node to ask info from supernode
func RcvSnAskInfo(msg *Message) (interface{}, error) {
	if msg.Kind != ON_SN_ASKINFO {
		return nil, errors.New("message Kind indicates not a ON_SN_ASKINFO")
	}

	backData := new(LocalInfo)
	backData.Ranklist = Global_ranking
	backData.Scoremap = make(map[string]UserRecord)

	Local_Info_Mutex.Lock()
	for k, v := range Local_map {
		backData.Scoremap[k] = v
	}
	Local_Info_Mutex.Unlock()

	sendoutMsg := new(Message)
	err := sendoutMsg.NewMsgwithData(msg.Src, SN_ON_ASKINFO_ACK, *backData)
	if err != nil {
		fmt.Println("In RcvSnAskInfo:")
		return nil, err
	}

	multicastMsgInGroup(sendoutMsg, false)

	return nil, nil
}


// rcv start/end msg from ON
func RcvSnStartEndFromON(msg *Message) (interface{}, error) {
	if msg.Kind != ON_SN_STARTEND {
		return nil, errors.New("message Kind indicates not a ON_SN_STARTEND")
	}

	newMessage := new(Message)
	newMessage.CopyMsg(msg)
	newMessage.Kind = SN_SN_STARTEND

	multicastMsgInGroup(newMessage, true)

	return nil, nil
}


// rcv start/end msg from other sn
func RcvSnStartEndFromSN(msg *Message) (interface{}, error) {
	if msg.Kind != SN_SN_STARTEND {
		return nil, errors.New("message Kind indicates not a SN_SN_STARTEND")
	}

	// multicast to ONs
	newMessage := new(Message)
	newMessage.CopyMsg(msg)
	newMessage.Kind = SN_ON_STARTEND

	multicastMsgInGroup(newMessage, false)

	return nil, nil
}


func updateLocalInfoWithOneRecord(userRecord UserRecord) bool {
	Local_Info_Mutex.Lock()
	rankChanged := false

	// Update local info
	if _, present := Local_map[userRecord.UserName]; present {
		// if local map data is better or equal, return
		if !userRecord.CompareTo(Local_map[userRecord.UserName]) {
			Local_Info_Mutex.Unlock()
			return rankChanged
		}
	}
	
	// new or better record
	Local_map[userRecord.UserName] = userRecord

	// Update the Global Rank
	var tmpUserRecord UserRecord
	for i := 0; i < GlobalRankSize; i++ {
		if !rankChanged {
			if userRecord.UserName != Global_ranking[i].UserName {
				if Global_ranking[i].CompareTo(userRecord) {
					continue
				} else {
					tmpUserRecord = Global_ranking[i]
					Global_ranking[i] = userRecord
					rankChanged = true
					if len(tmpUserRecord.UserName) == 0 {
						break
					}
				}
			} else {
				if userRecord.CompareTo(Global_ranking[i]) {
					Global_ranking[i] = userRecord
					rankChanged = true
				}
				break
			}
		} else {
			if userRecord.UserName != Global_ranking[i].UserName {
				tmpUserRecord1 := Global_ranking[i]
				Global_ranking[i] = tmpUserRecord
				tmpUserRecord = tmpUserRecord1
				if len(tmpUserRecord.UserName) == 0 {
					break
				}
			} else {
				Global_ranking[i] = tmpUserRecord
				break
			}
		}
	}

	// print out new global ranking
	for i := range Global_ranking {
		fmt.Println(Global_ranking[i].String())
	}

	Local_Info_Mutex.Unlock()
	return rankChanged
}


// is Super: true: send to other SNs, false: send to ON in group
func multicastMsgInGroup(m *Message, isSuper bool) {
	newMCastMsg := new(MultiCastMessage)
	tmpMsg := &newMCastMsg.Message
	tmpMsg.CopyMsg(m)
	newMCastMsg.Origin = MsgPasser.ServerIP
	newMCastMsg.Seqnum = atomic.AddInt32(&MsgPasser.SeqNum, 1)

	newMCastMsg.HostList = make([]string, 0)

	if isSuper {
		fmt.Printf("SuperNOdeHandler: multicastMsgInGroup SNHostList %d\n", MsgPasser.SNHostlist.Len())

		for e := MsgPasser.SNHostlist.Front(); e != nil; e = e.Next() {
			newMCastMsg.HostList = append(newMCastMsg.HostList, e.Value.(string))
		}

	} else {
		fmt.Printf("SuperNOdeHandler: multicastMsgInGroup ONHostList %d\n", MsgPasser.ONHostlist.Len())

		for e := MsgPasser.ONHostlist.Front(); e != nil; e = e.Next() {
			newMCastMsg.HostList = append(newMCastMsg.HostList, e.Value.(string))
		}
	}
	MsgPasser.SendMCast(newMCastMsg)
}


func multicastGlobalRankToSNs() {
	newMCastMsg := new(MultiCastMessage)
	newMCastMsg.NewMCastMsgwithData("", SN_SN_RANK, Global_ranking)
	newMCastMsg.Origin = MsgPasser.ServerIP
	newMCastMsg.Seqnum = atomic.AddInt32(&MsgPasser.SeqNum, 1)

	newMCastMsg.HostList = make([]string, 0)
	for e := MsgPasser.SNHostlist.Front(); e != nil; e = e.Next() {
		newMCastMsg.HostList = append(newMCastMsg.HostList, e.Value.(string))
	}
	
	MsgPasser.SendMCast(newMCastMsg)
}

func parseConfigFile() {

}
