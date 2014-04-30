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
	if DBstatus == "success" {
		// put user into request map
		SignUp_commitLock.Lock()
		userStatus := new(SignUpCommitStatus)
		userStatus.NewSignUpCmitStatus()
		signUp_requestMap[username] = userStatus
		SignUp_commitLock.Unlock()

		// open a thread to check first commit status for fixed time
		commitStatusChan := make(chan string)
		go checkCommitStatus(commitStatusChan, username)

		// send commit_ready to other SNs
		commitReadyMsg := new(Message)
		err = commitReadyMsg.NewMsgwithData("", SN_SN_COMMIT_RD, username)
		if err != nil {
			fmt.Println("In RcvSnSignUp:")
			SignUp_commitLock.Lock()
			delete(signUp_requestMap, username)
			SignUp_commitLock.Unlock()
			return nil, err
		}
		MulticastMsgInGroup(commitReadyMsg, true)

		// get result and send back commit result to SNs
		status := <-commitStatusChan

		// send commit_ready to other SNs
		commitResultMsg := new(Message)
		resultData := map[string]string{
			"username": username,
			"email":    email,
			"password": password,
			"status":   status,
		}
		err = commitResultMsg.NewMsgwithData("", SN_SN_SIGNUP, resultData)
		if err != nil {
			fmt.Println("In RcvSnSignUp:")
			SignUp_commitLock.Lock()
			delete(signUp_requestMap, username)
			SignUp_commitLock.Unlock()
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
		SignUp_commitLock.Lock()
		delete(signUp_requestMap, username)
		SignUp_commitLock.Unlock()
	} else {
		//DB check failed
		ONstatus = DBstatus
	}

	StartEnd_Lock.Lock()
	// send status back to ON
	backData := map[string]string{
		"user":      username,
		"status":    ONstatus,
		"question":  Question_URI,
		"startTime": StartTime,
	}
	StartEnd_Lock.Unlock()

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
	if _, exist := SignUp_commit_readySet[username]; exist {
		status = "Abort"
	} else {
		// ready to commit this user
		SignUp_commitLock.Lock()
		SignUp_commit_readySet[username] = msg.Origin
		SignUp_commitLock.Unlock()
	}

	// send status to master
	commitACKMsg := new(Message)
	readyMsg := map[string]string{
		"status":   status,
		"username": username,
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
		SignUp_commitLock.Lock()
		if status == "Abort" {
			signUp_requestMap[username].HasAbort = true
		} else {
			signUp_requestMap[username].ReadySNIP[msg.Src] = true
		}
		SignUp_commitLock.Unlock()
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

	if mSignUpMsg["status"] == "Commit" {
		status := util.DatabaseSignUp(mSignUpMsg["username"], mSignUpMsg["password"], mSignUpMsg["email"])
		if status != "success" {
			return nil, err
		}

	}

	// delete it from request map
	SignUp_commitLock.Lock()
	delete(SignUp_commit_readySet, mSignUpMsg["username"])
	SignUp_commitLock.Unlock()

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

	StartEnd_Lock.Lock()
	backData := map[string]string{
		"user":      signInMsg["username"],
		"status":    backMsg,
		"question":  Question_URI,
		"startTime": StartTime,
	}
	StartEnd_Lock.Unlock()

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
		globalRankMsg := new(Message)
		err := globalRankMsg.NewMsgwithData("", SN_SN_RANK, Global_ranking)
		if err != nil {
			fmt.Println("In RcvSnPblSuccess:")
			return nil, err
		}

		MulticastMsgInGroup(globalRankMsg, true)
		//multicastGlobalRankToSNs()
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
	//for i := range Global_ranking {
	//	fmt.Println(Global_ranking[i].String())
	//}

	Local_Info_Mutex.Unlock()
	return rankChanged
}

// merge global ranking with a new list
func mergeGlobalRankingWithList(newRankList [GlobalRankSize]UserRecord) {
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
}

// find the first empty slot of a ranklist
func getEmptyPos(rankList [GlobalRankSize]UserRecord) int {
	index := 0
	for ; index < GlobalRankSize; index++ {
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
		//fmt.Printf("SuperNodeHandler: multicastMsgInGroup SNHostList %d\n", len(MsgPasser.SNHostlist))
		SNHostlistMutex.Lock()
		newMCastMsg.HostList = MsgPasser.SNHostlist
		MsgPasser.SendMCast(newMCastMsg)
		SNHostlistMutex.Unlock()
	} else {
		//fmt.Printf("SuperNodeHandler: multicastMsgInGroup ONHostList %d\n", len(MsgPasser.ONHostlist))
		ONHostlistMutex.Lock()
		newMCastMsg.HostList = MsgPasser.ONHostlist
		MsgPasser.SendMCast(newMCastMsg)
		ONHostlistMutex.Unlock()
	}
}
