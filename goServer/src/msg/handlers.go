package msg

import (
	"encoding/json"
	"errors"
	"fmt"
	//"util"
)

var Handlers [NUMTYPES]Rcvhdlr

type Rcvhdlr func(*Message) (interface{}, error)

/*type Handler struct {
	Action Rcvhdlr
}*/

var SignInChan chan string
var SignUpChan chan string

var App_url = "http://localhost:8000/"

var Global_ranking = [GlobalRankSize]UserRecord{}
var Local_info = map[string]UserRecord{}

func RcvString(msg *Message) (interface{}, error) {
	if msg.Kind != STRING {
		return nil, errors.New("message Kind indicates not a STRING")
	}

	var rcvString string
	if err := ParseRcvInterfaces(msg, &rcvString); err != nil {
		return nil, err
	}

	return rcvString, nil
}

func RcvAskInfoAck(msg *Message) (interface{}, error) {
	if msg.Kind != ASKINFOACK {
		return nil, errors.New("message Kind indicates not a ASKINFOACK")
	}

	var signInMsg LocalInfo
	err := ParseRcvInterfaces(msg, &signInMsg)
	if err != nil {
		return nil, err
	}
	
	// update local info and ranking
	Local_info = signInMsg.Scoremap
	Global_ranking = signInMsg.Ranklist
	return signInMsg, err
}

func RcvNodeJoin(msg *Message) (interface{}, error) {
	if msg.Kind != SN_NODEJOIN {
		return nil, errors.New("message Kind indicates not a STRING")
	}

	return nil, nil
}

// received in SN
func RcvSignInAck(msg *Message) (interface{}, error) {
	if msg.Kind != SIGNINACK {
		return nil, errors.New("message Kind indicates not a SIGNINACK")
	}

	var signInAckMsg map[string]string
	if err := ParseRcvInterfaces(msg, &signInAckMsg); err != nil {
		return nil, err
	}
	
	// if success, send update request
	if signInAckMsg["status"] == "success" {
		// send map[string]string messages to SN
		sendoutMsg := new(Message)
		err := sendoutMsg.NewMsgwithData(SuperNodeIP, SN_ASKINFO, "")
		if err != nil {
			fmt.Println(err)
		}
		// send message to SN
		MsgPasser.Send(sendoutMsg)
	}
	
	//TODO: read the question url if needed

	// update signIn channel to stop the channel waiting
	SignInChan <- signInAckMsg["status"]

	return signInAckMsg, nil
}

// received in SN
func RcvSignUpAck(msg *Message) (interface{}, error) {
	if msg.Kind != SIGNUPACK {
		return nil, errors.New("message Kind indicates not a SIGNUPACK")
	}

	var signUpAckMsg map[string]string
	if err := ParseRcvInterfaces(msg, &signUpAckMsg); err != nil {
		return nil, err
	}

	// if success, send update request
	if signUpAckMsg["status"] == "success" {
		// send map[string]string messages to SN
		sendoutMsg := new(Message)
		err := sendoutMsg.NewMsgwithData(SuperNodeIP, SN_ASKINFO, "")
		if err != nil {
			fmt.Println(err)
		}
		// send message to SN
		MsgPasser.Send(sendoutMsg)
	}
	
	//TODO: read the question url if needed

	// update signIn channel to stop the channel waiting
	SignUpChan <- signUpAckMsg["status"]

	return signUpAckMsg, nil
}

//TODO: SN receive start or end, forward message to all ON and SN
// use NewMsgwithBytes will be efficient without en/decode msg
func RcvStartEnd_SN(msg *Message) (interface{}, error) {
	return nil, nil
}

// When ON receive start or end, call app url, to inform app
func RcvStartEnd_ON(msg *Message) (interface{}, error) {
	if msg.Kind != STARTEND_ON {
		return nil, errors.New("message Kind indicates not a STARTEND_ON")
	}

	var startEndMsg map[string]string
	if err := ParseRcvInterfaces(msg, &startEndMsg); err != nil {
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
		return nil, errors.New("STARTEND_ON message's inner type wrong")
	}
	return startEndMsg, nil
}
