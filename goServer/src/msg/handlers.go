package msg

import (
	"errors"
	"util"
	"fmt"
	"time"
)

var Handlers [NUMTYPES]Handler

type Sendhdlr func(*Message, interface{})error
type Rcvhdlr func(*Message)(interface{},error)

type Handler struct {
	Decode Rcvhdlr
}

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
	return ParseRcvString(msg)
}

func RcvPblSuccess(msg *Message)(interface{}, error) {
	if msg.Kind != PBLSUCCESS {
		return nil, errors.New("message Kind indicates not a PBLSUCCESS")
	}
	
	// TODO: for SN, it should merge to global ranking, and send back if needed
	return ParseRcvMapStrings(msg)
}

// should be in SN
func RcvSignIn(msg *Message)(interface{}, error) {
	if msg.Kind != SIGNIN {
		return nil, errors.New("message Kind indicates not a SIGNIN")
	}

	signInMsg, err :=  ParseRcvMapStrings(msg)
	if err != nil {
		return signInMsg, err
	}

	// check database and send back SignInAck
	backMsg := util.DatabaseSignIn(signInMsg["username"], signInMsg["password"])
	
	backData := map[string]string {
		"user" : signInMsg["username"],
		"status" : backMsg,
	}

	sendoutMsg := new(Message)
	err = sendoutMsg.NewMsgwithData(msg.Src, SIGNINACK, backData)
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
	
	signInAckMsg, err :=  ParseRcvMapStrings(msg)
	if err != nil {
		return signInAckMsg, err
	}
	
	// update signIn channel to stop the channel waiting
	SignInChan <- signInAckMsg["status"]
	
	return signInAckMsg, err
}


func RcvSignUp(msg *Message)(interface{}, error) {
	if msg.Kind != SIGNUP {
		return nil, errors.New("message Kind indicates not a SIGNUP")
	}

	signUpMsg, err :=  ParseRcvMapStrings(msg)
	if err != nil {
		return signUpMsg, err
	}

	// check database and send back SignInAck
	backMsg := util.DatabaseSignUp(signUpMsg["username"], 
		signUpMsg["password"], signUpMsg["email"])

	backData := map[string]string {
		"email" : signUpMsg["email"],
		"status" : backMsg,
	}

	sendoutMsg := new(Message)
	err = sendoutMsg.NewMsgwithData(msg.Src, SIGNUPACK, backData)
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
	
	signUpAckMsg, err :=  ParseRcvMapStrings(msg)
	if err != nil {
		return signUpAckMsg, err
	}

	// update signIn channel to stop the channel waiting
	SignUpChan <- signUpAckMsg["status"]
	
	return signUpAckMsg, err
}
