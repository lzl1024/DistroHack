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
	
	username := signUpMsg["username"]
	email := signUpMsg["email"]
	password := signUpMsg["password"]
	ONstatus := "failed"
	
	// check local database to see whether is has been registered
	DBstatus := util.DatabaseCheckUser(username, email)
	if  DBstatus == "success" {
		// open a thread to check first commit status for fixed time 
		commitStatusChan := make(chan string)
		go checkCommitStatus(commitStatusChan, username)
		
		// put user into request map
		signUp_commitLock.Lock()
		userStatus := new(SignUpCommitStatus)
		userStatus.NewSignUpCmitStatus()
		signUp_requestMap[username] = userStatus
		signUp_commitLock.Unlock()
		
		// send commit_ready to other SNs
		commitReadyMsg := new(Message)
		err = commitReadyMsg.NewMsgwithData("", SN_SN_COMMIT_RD, username)
		if err != nil {
			fmt.Println("In RcvSnSignUp:")
			signUp_commitLock.Lock()
			delete(signUp_requestMap, username)
			signUp_commitLock.Unlock()
			return nil, err
		}
		MulticastMsgInGroup(commitReadyMsg, true)
	
		// get result and send back commit result to SNs
		status := <- commitStatusChan
		
		// send commit_ready to other SNs
		commitResultMsg := new(Message)
		resultData := map[string]string{
			"username":     username,
			"email": 	email,
			"password":	password,
			"status":   status,
		}
		err = commitResultMsg.NewMsgwithData("", SN_SN_SIGNUP, resultData)
		if err != nil {
			fmt.Println("In RcvSnSignUp:")
			signUp_commitLock.Lock()
			delete(signUp_requestMap, username)
			signUp_commitLock.Unlock()
			return nil, err
		}
		MulticastMsgInGroup(commitResultMsg, true)
		
		// if success, update local information
		ONstatus = "failed"
		if status == "Commit" {
			userRecord := new(UserRecord)
			userRecord.UserName = username
			updateLocalInfoWithOneRecord(*userRecord)
			fmt.Printf("SuperNode: ordinary sign up %s,  status %s\n", username, status)
			ONstatus = "success"
		}
		
		// delete user request from request map			
		signUp_commitLock.Lock()
		delete(signUp_requestMap, username)
		signUp_commitLock.Unlock()
	} else {
		//DB check failed
		ONstatus = DBstatus
	}
	
	// send status back to ON
	backData := map[string]string{
		"user":     username,
		"status":   ONstatus,
		"question": "url",
	}

	sendoutMsg := new(Message)
	err = sendoutMsg.NewMsgwithData(msg.Src, SN_ON_SIGNUP_ACK, backData)
	if err != nil {
		fmt.Println("In RcvSnSignUp:")
		return nil, err
	}
	MsgPasser.Send(sendoutMsg)

	return signUpMsg, err
}


// rcv the multicast commit_ready from SN 
func RcvSnSignUpCommitReady(msg *Message) (interface{}, error) {
	if msg.Kind != SN_SN_COMMIT_RD {
		return nil, errors.New("message Kind indicates not a SN_SN_COMMIT_RD")
	}
	
	var username string
	if err := ParseRcvInterfaces(msg, &username); err != nil {
		fmt.Println("In RcvSnSignUpCommitReady:")
		return nil, err
	}
	
	status := "Ready"
	// check ready set to see whether it has been registered
	if _, exist := signUp_commit_readySet[username]; exist {
		status = "Abort"
	} else {
		// ready to commit this user
		signUp_commitLock.Lock()
		signUp_commit_readySet[username] = true
		signUp_commitLock.Unlock()
	}
	
	// send status to master
	commitACKMsg := new(Message)
	readyMsg := map[string]string {
		"status" 	: 	status,
		"username"	:	username,
	}
	
	err := commitACKMsg.NewMsgwithData(msg.Origin, SN_SN_COMMIT_RD_ACK, readyMsg)
	if err != nil {
		fmt.Println("In RcvSnSignUpCommitReady:")
		return nil, err
	}
	MsgPasser.Send(commitACKMsg)
	
	return username, nil
}


// rcv commit_ready ack from workers
func RcvSnSignUpCommitReadyACK(msg *Message) (interface{}, error) {
	if msg.Kind != SN_SN_COMMIT_RD_ACK {
		return nil, errors.New("message Kind indicates not a SN_SN_COMMIT_RD_ACK")
	}
	
	var readyReply map[string]string
	if err := ParseRcvInterfaces(msg, &readyReply); err != nil {
		fmt.Println("In RcvSnSignUpCommitReadyACK:")
		return nil, err
	}
	
	username := readyReply["username"]
	status := readyReply["status"]
	
	// check ready set to see whether it has conflict
	if _, exist := signUp_requestMap[username]; exist {
		signUp_commitLock.Lock()
		if status == "Abort" {
			signUp_requestMap[username].HasAbort = true
		} else {
			signUp_requestMap[username].ReadySNIP[msg.Src] = true
		}
		signUp_commitLock.Unlock()
	}
	
	return readyReply, nil
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

	fmt.Println("SIGNUP GET MSG:" , mSignUpMsg)
	if mSignUpMsg["status"] == "Commit" {
		status := util.DatabaseSignUp(mSignUpMsg["username"], mSignUpMsg["password"], mSignUpMsg["email"])
		if status != "success" {
			return nil, err
		}
		
		// TODO: should update global rank and SN_RANK to his SNs?
	}
	
	// delete it from request map
	signUp_commitLock.Lock()
	delete(signUp_commit_readySet, mSignUpMsg["username"])
	signUp_commitLock.Unlock()

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
	MulticastMsgInGroup(sendoutMsg, false)

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

	//update the global rank list in local
	mergeGlobalRankingWithList(newRankList)
	//Global_ranking = newRankList

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
	MulticastMsgInGroup(newMessage, false)

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

	MulticastMsgInGroup(sendoutMsg, false)

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

	MulticastMsgInGroup(newMessage, true)

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

	MulticastMsgInGroup(newMessage, false)

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



// merge global ranking with a new list
func mergeGlobalRankingWithList(newRankList [GlobalRankSize]UserRecord){
	newList := [GlobalRankSize]UserRecord{}
	indexGlobalR := 0
	indexNewR := 0
	index := 0
	userNameSet := map[string]bool{}
	endGlobalR := getEmptyPos(Global_ranking)
	endNewR := getEmptyPos(newRankList)
	
	// main merge
	for index < GlobalRankSize && indexGlobalR < endGlobalR && indexNewR < endNewR {
		if Global_ranking[indexGlobalR].CompareTo(newRankList[indexNewR]) {
			// only when username not in set, add them
			if _, exist := userNameSet[Global_ranking[indexGlobalR].UserName]; !exist {
				newList[index] = Global_ranking[indexGlobalR]
				// put username into map
				userNameSet[Global_ranking[indexGlobalR].UserName] = true
				index++
			}
			indexGlobalR++
		} else {
			// only when username not in set, add them
			if _, exist := userNameSet[newRankList[indexNewR].UserName]; !exist {
				newList[index] = newRankList[indexNewR]
				userNameSet[newRankList[indexNewR].UserName] = true
				index++
			}
			indexNewR++
		}
	}
	
	for index < GlobalRankSize && indexGlobalR < endGlobalR {
		if _, exist := userNameSet[Global_ranking[indexGlobalR].UserName]; !exist {
			newList[index] = Global_ranking[indexGlobalR]
			userNameSet[Global_ranking[indexGlobalR].UserName] = true
			index++
		}
		indexGlobalR++
	}
	
	for index < GlobalRankSize && indexNewR < endNewR {
		if _, exist := userNameSet[newRankList[indexNewR].UserName]; !exist {
			newList[index] = newRankList[indexNewR]
			userNameSet[newRankList[indexNewR].UserName] = true
			index++
		}
		indexNewR++
	}
	
	Global_ranking = newList
	
	//fmt.Println("NEW GLOBAL RANKING!!!!")
	//for i := range Global_ranking {
	//	fmt.Println(Global_ranking[i].String())
	//}
}


// find the first empty slot of a ranklist
func getEmptyPos(rankList [GlobalRankSize]UserRecord) int {
	index := 0
	for  ; index < GlobalRankSize; index++ {
		if len(rankList[index].UserName) == 0 {
			return index
		}
	}
	return index
}


// is Super: true: send to other SNs, false: send to ON in group
func MulticastMsgInGroup(m *Message, isSuper bool) {
	newMCastMsg := new(MultiCastMessage)
	tmpMsg := &newMCastMsg.Message
	tmpMsg.CopyMsg(m)
	newMCastMsg.Origin = MsgPasser.ServerIP
	newMCastMsg.Seqnum = atomic.AddInt32(&MsgPasser.SeqNum, 1)

	newMCastMsg.HostList = make(map[string]string)

	if isSuper {
		fmt.Printf("SuperNodeHandler: multicastMsgInGroup SNHostList %d\n", len(MsgPasser.SNHostlist))

		for k,_ := range MsgPasser.SNHostlist {
			newMCastMsg.HostList[k] = MsgPasser.SNHostlist[k]
		}

	} else {
		fmt.Printf("SuperNodeHandler: multicastMsgInGroup ONHostList %d\n", len(MsgPasser.ONHostlist))

		for k,_ := range MsgPasser.ONHostlist {
			newMCastMsg.HostList[k] = MsgPasser.ONHostlist[k]
		}
	}
	MsgPasser.SendMCast(newMCastMsg)
}


func multicastGlobalRankToSNs() {
	newMCastMsg := new(MultiCastMessage)
	newMCastMsg.NewMCastMsgwithData("", SN_SN_RANK, Global_ranking)
	newMCastMsg.Origin = MsgPasser.ServerIP
	newMCastMsg.Seqnum = atomic.AddInt32(&MsgPasser.SeqNum, 1)

	newMCastMsg.HostList = make(map[string]string)
	for k,_ := range MsgPasser.SNHostlist {
		newMCastMsg.HostList[k] = MsgPasser.SNHostlist[k]
	}
	
	MsgPasser.SendMCast(newMCastMsg)
}



