package msg

import (
	"errors"
	"util"
	"fmt"
	"time"
)

var Handlers [NUMTYPES]Rcvhdlr

type Rcvhdlr func(*Message)(interface{},error)

/*type Handler struct {
	Action Rcvhdlr
}*/

var SignInChan chan string
var SignUpChan chan string

type User_record struct {
	UserName string
	Score    int
	Ctime    time.Time
}

var App_url = "http://localhost:8000/"

var Global_ranking_size = 20

var Global_ranking  = []User_record{}
var Local_info  = map[string]User_record{}


func RcvString(msg *Message)(interface{}, error) {
	if msg.Kind != STRING {
		return nil, errors.New("message Kind indicates not a STRING")
	}
	
	var rcvString string
	if err := ParseRcvInterfaces(msg, &rcvString); err!= nil {
		return nil, err
	}
	
	return rcvString, nil
}

func RcvPblSuccess(msg *Message)(interface{}, error) {
	if msg.Kind != PBLSUCCESS {
		return nil, errors.New("message Kind indicates not a PBLSUCCESS")
	}
	
	// TODO: for SN, it should merge to global ranking, and send back if needed
	var successMsg map[string]string
	if err := ParseRcvInterfaces(msg, &successMsg); err!= nil {
		return nil, err
	}
	return successMsg, nil
}

// should be in SN
func RcvSignIn(msg *Message)(interface{}, error) {
	if msg.Kind != SIGNIN {
		return nil, errors.New("message Kind indicates not a SIGNIN")
	}
	
	var signInMsg map[string]string
	if err := ParseRcvInterfaces(msg, &signInMsg); err!= nil {
		return nil, err
	}

	// check database and send back SignInAck
	backMsg := util.DatabaseSignIn(signInMsg["username"], signInMsg["password"])
	
	backData := map[string]string {
		"user" : signInMsg["username"],
		"status" : backMsg,
	}

	sendoutMsg := new(Message)
	err := sendoutMsg.NewMsgwithData(msg.Src, SIGNINACK, backData)
	if err != nil {
		fmt.Println(err);
	}

	// send message to SN
	MsgPasser.Send(sendoutMsg, false)
	return signInMsg, err
}

func RcvSignInAck(msg *Message)(interface{}, error) {
	if msg.Kind != SIGNINACK {
		return nil, errors.New("message Kind indicates not a SIGNINACK")
	}
	
	var signInAckMsg map[string]string
	if err := ParseRcvInterfaces(msg, &signInAckMsg); err!= nil {
		return nil, err
	}
	
	// update signIn channel to stop the channel waiting
	SignInChan <- signInAckMsg["status"]
	
	return signInAckMsg, nil
}


func RcvSignUp(msg *Message)(interface{}, error) {
	if msg.Kind != SIGNUP {
		return nil, errors.New("message Kind indicates not a SIGNUP")
	}
	
	var signUpMsg map[string]string
	if err := ParseRcvInterfaces(msg, &signUpMsg); err!= nil {
		return nil, err
	}

	// check database and send back SignInAck
	backMsg := util.DatabaseSignUp(signUpMsg["username"], 
		signUpMsg["password"], signUpMsg["email"])

	backData := map[string]string {
		"email" : signUpMsg["email"],
		"status" : backMsg,
	}

	sendoutMsg := new(Message)
	err := sendoutMsg.NewMsgwithData(msg.Src, SIGNUPACK, backData)
	if err != nil {
		fmt.Println(err);
	}
	// send message to SN
	MsgPasser.Send(sendoutMsg, false)
	return signUpMsg, err
}

func RcvSignUpAck(msg *Message)(interface{}, error) {
	if msg.Kind != SIGNUPACK {
		return nil, errors.New("message Kind indicates not a SIGNUPACK")
	}
	
	var signUpAckMsg map[string]string
	if err := ParseRcvInterfaces(msg, &signUpAckMsg); err!= nil {
		return nil, err
	}

	// update signIn channel to stop the channel waiting
	SignUpChan <- signUpAckMsg["status"]
	
	return signUpAckMsg, nil
}
