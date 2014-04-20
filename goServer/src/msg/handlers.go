package msg

import (
	"encoding/json"
	"errors"
	"fmt"
)

var Handlers [NUMTYPES]Rcvhdlr

type Rcvhdlr func(*Message) (interface{}, error)


var SignInChan chan string
var SignUpChan chan string

var App_url = "http://localhost:8000/"


// for test and debug
func RcvString(msg *Message) (interface{}, error) {
	if msg.Kind != STRING {
		return nil, errors.New("message Kind indicates not a STRING")
	}

	var rcvString string
	if err := ParseRcvInterfaces(msg, &rcvString); err != nil {
		fmt.Println("In RcvString:")
		return nil, err
	}

	return rcvString, nil
}


// rcv global_ranking, local_info from SN when  
func RcvAskInfoAck(msg *Message) (interface{}, error) {
	if msg.Kind != SN_ON_ASKINFO_ACK {
		return nil, errors.New("message Kind indicates not a SN_ON_ASKINFO_ACK")
	}

	var signInMsg LocalInfo
	err := ParseRcvInterfaces(msg, &signInMsg)
	if err != nil {
		fmt.Println("In RcvAskInfoAck:")
		return nil, err
	}

	// update local info and ranking
	Local_Info_Mutex.Lock()
	Global_ranking = signInMsg.Ranklist
	
	Local_map = signInMsg.Scoremap

	// update the app's ranking
	data, _ := json.Marshal(Global_ranking)
	Local_Info_Mutex.Unlock()

	// send data out
	SendtoApp(App_url+"hacks/update_rank/", string(data))

	return signInMsg, err
}


// rcv the sign in status msg from SN
func RcvSignInAck(msg *Message) (interface{}, error) {
	if msg.Kind != SN_ON_SIGNIN_ACK {
		return nil, errors.New("message Kind indicates not a SN_ON_SIGNIN_ACK")
	}

	var signInAckMsg map[string]string
	if err := ParseRcvInterfaces(msg, &signInAckMsg); err != nil {
		fmt.Println("In RcvSignInAck:")
		return nil, err
	}

	// if success, send update request
	if signInAckMsg["status"] == "success" {
		// update app question set
		SendtoApp(App_url+"hacks/updateq/", string(signInAckMsg["question"]))

		// send map[string]string messages to SN
		sendoutMsg := new(Message)
		err := sendoutMsg.NewMsgwithData(SuperNodeIP, ON_SN_ASKINFO, "")
		if err != nil {
			fmt.Println("In RcvSignInAck:")
			return nil, err
		}
		// send message to SN
		MsgPasser.Send(sendoutMsg)
	}

	// update signIn channel to stop the channel waiting
	SignInChan <- signInAckMsg["status"]

	return signInAckMsg, nil
}


// rcv the sign up status msg from SN
func RcvSignUpAck(msg *Message) (interface{}, error) {
	if msg.Kind != SN_ON_SIGNUP_ACK {
		return nil, errors.New("message Kind indicates not a SN_ON_SIGNUP_ACK")
	}

	var signUpAckMsg map[string]string
	if err := ParseRcvInterfaces(msg, &signUpAckMsg); err != nil {
		fmt.Println("In RcvSignUpAck:")
		return nil, err
	}

	// if success, send update request
	if signUpAckMsg["status"] == "success" {
		// update app question set
		SendtoApp(App_url+"hacks/updateq/", string(signUpAckMsg["question"]))

		// send map[string]string messages to SN
		sendoutMsg := new(Message)
		err := sendoutMsg.NewMsgwithData(SuperNodeIP, ON_SN_ASKINFO, "")
		if err != nil {
			fmt.Println("In RcvSignUpAck:")
			return nil, err
		}
		// send message to SN
		MsgPasser.Send(sendoutMsg)
	}

	// update signIn channel to stop the channel waiting
	SignUpChan <- signUpAckMsg["status"]

	return signUpAckMsg, nil
}

// When ON receive start or end, call app url, to inform app
func RcvStartEnd(msg *Message) (interface{}, error) {
	if msg.Kind != SN_ON_STARTEND {
		return nil, errors.New("message Kind indicates not a SN_ON_STARTEND")
	}

	var startEndMsg map[string]string
	if err := ParseRcvInterfaces(msg, &startEndMsg); err != nil {
		fmt.Println("In RcvStartEnd:")
		return nil, err
	}

	// end message
	if startEndMsg["type"] == "end_hack" {
		// send data out
		SendtoApp(App_url+"hacks/end_hack/", "")
	} else if startEndMsg["type"] == "start_hack" {
		data, _ := json.Marshal(startEndMsg)
		SendtoApp(App_url+"hacks/start_hack/", string(data))
	} else {
		fmt.Println("In RcvStartEnd:")
		return nil, errors.New("SN_ON_STARTEND message's inner type wrong")
	}

	return startEndMsg, nil
}


// SN or ON rcv ranking update
func RcvSnRankfromSN(msg *Message) (interface{}, error) {
	if msg.Kind != SN_ON_RANK {
		return nil, errors.New("message Kind indicates not a SN_ON_RANK")
	}

	var newRankList [GlobalRankSize]UserRecord
	if err := ParseRcvInterfaces(msg, &newRankList); err != nil {
		fmt.Println("In RcvSnRankfromSN:")
		return nil, err
	}
	
	//update the global rank list in local
	Local_Info_Mutex.Lock()
	Global_ranking = newRankList
	Local_Info_Mutex.Unlock()

	return newRankList, nil
}