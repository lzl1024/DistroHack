package msg

import (
	"errors"
	"util"
	"fmt"
)

var Handlers [NUMTYPES]Handler

type Sendhdlr func(*Message, interface{})error
type Rcvhdlr func(*Message)(interface{},error)

type Handler struct {
	Encode Sendhdlr
	Decode Rcvhdlr
}

var SignInMap = map[string]string{}
var SignUpMap = map[string]string{}

/*
 * Plain String send and receive
 */
func SendString(msg *Message, data interface{}) error {
	if msg.Kind != STRING {
		return errors.New("message Kind indicates not a STRING")
	}
	
	return ParseSendInterfaces(msg, data)
}

func RcvString(msg *Message)(interface{}, error) {
	if msg.Kind != STRING {
		return nil, errors.New("message Kind indicates not a STRING")
	}
	return ParseRcvString(msg)
}

/*
 * Problem success send and receive
 */
func SendPblSuccess(msg *Message, data interface{}) error {
	if msg.Kind != PBLSUCCESS {
		return errors.New("message Kind indicates not a PBLSUCCESS")
	}
	
	return ParseSendInterfaces(msg, data)
}

func RcvPblSuccess(msg *Message)(interface{}, error) {
	if msg.Kind != PBLSUCCESS {
		return nil, errors.New("message Kind indicates not a PBLSUCCESS")
	}
	return ParseRcvMapStrings(msg)
}

/*
 * Sign In send and receive
 */
func SendSignIn(msg *Message, data interface{}) error {
	if msg.Kind != SIGNIN {
		return errors.New("message Kind indicates not a SIGNIN")
	}
	
	return ParseSendInterfaces(msg, data)
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
	sendoutMsg.Kind = SIGNINACK
	sendoutMsg.Dest = msg.Src

	err = Handlers[sendoutMsg.Kind].Encode(sendoutMsg, backData)
	if err != nil {
		fmt.Println(err);
	}

	// send message to SN
	MsgPasser.Send(sendoutMsg, false)
	return signInMsg, err
}

/*
 * SignInAck send and receive
 */
func SendSignInAck(msg *Message, data interface{}) error {
	if msg.Kind != SIGNINACK {
		return errors.New("message Kind indicates not a SIGNINACK")
	}
	
	return ParseSendInterfaces(msg, data)
}

func RcvSignInAck(msg *Message)(interface{}, error) {
	if msg.Kind != SIGNINACK {
		return nil, errors.New("message Kind indicates not a SIGNINACK")
	}
	
	signInAckMsg, err :=  ParseRcvMapStrings(msg)
	if err != nil {
		return signInAckMsg, err
	}
	
	// update signInMap to stop the busy waiting
	SignInMap[signInAckMsg["user"]], _ = signInAckMsg["status"]
	
	return signInAckMsg, err
}


/*
 * Sign Up send and receive
 */
func SendSignUp(msg *Message, data interface{}) error {
	if msg.Kind != SIGNUP {
		return errors.New("message Kind indicates not a SIGNUP")
	}
	
	return ParseSendInterfaces(msg, data)
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
	sendoutMsg.Kind = SIGNUPACK
	sendoutMsg.Dest = msg.Src

	err = Handlers[sendoutMsg.Kind].Encode(sendoutMsg, backData)
	if err != nil {
		fmt.Println(err);
	}

	// send message to SN
	MsgPasser.Send(sendoutMsg, false)
	return signUpMsg, err
}

/*
 * SignUpAck send and receive
 */
func SendSignUpAck(msg *Message, data interface{}) error {
	if msg.Kind != SIGNUPACK {
		return errors.New("message Kind indicates not a SIGNUPACK")
	}
	
	return ParseSendInterfaces(msg, data)
}

func RcvSignUpAck(msg *Message)(interface{}, error) {
	if msg.Kind != SIGNUPACK {
		return nil, errors.New("message Kind indicates not a SIGNUPACK")
	}
	
	signUpAckMsg, err :=  ParseRcvMapStrings(msg)
	if err != nil {
		return signUpAckMsg, err
	}

	// update signInMap to stop the busy waiting
	SignUpMap[signUpAckMsg["email"]] = signUpAckMsg["status"]
	
	return signUpAckMsg, err
}
