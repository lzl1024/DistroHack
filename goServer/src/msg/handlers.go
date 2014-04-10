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

	// update the app's ranking
	data, _ := json.Marshal(Global_ranking)

	// send data out
	SendtoApp(App_url+"hacks/update_rank/", string(data))

	return signInMsg, err
}

func RcvNodeJoin(msg *Message) (interface{}, error) {
	if msg.Kind != SN_NODEJOIN {
		return nil, errors.New("message Kind indicates not a STRING")
	}

	return nil, nil
}

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
		// update app question set
		SendtoApp(App_url+"hacks/updateq/", string(signInAckMsg["question"]))

		// send map[string]string messages to SN
		sendoutMsg := new(Message)
		err := sendoutMsg.NewMsgwithData(SuperNodeIP, SN_ASKINFO, "")
		if err != nil {
			fmt.Println(err)
		}
		// send message to SN
		MsgPasser.Send(sendoutMsg)

	}

	// update signIn channel to stop the channel waiting
	SignInChan <- signInAckMsg["status"]

	return signInAckMsg, nil
}

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
		// update app question set
		SendtoApp(App_url+"hacks/updateq/", string(signUpAckMsg["question"]))

		// send map[string]string messages to SN
		sendoutMsg := new(Message)
		err := sendoutMsg.NewMsgwithData(SuperNodeIP, SN_ASKINFO, "")
		if err != nil {
			fmt.Println(err)
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
	if msg.Kind != STARTEND {
		return nil, errors.New("message Kind indicates not a STARTEND")
	}

	fmt.Println("Handlers RcvStartEnd: ", msg.String())

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
		return nil, errors.New("STARTEND message's inner type wrong")
	}
	return startEndMsg, nil
}
