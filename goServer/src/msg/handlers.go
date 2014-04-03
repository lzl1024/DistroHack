package msg

import (
	"errors"
	"util"
	"fmt"
	"time"
	"encoding/json"
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

// Received in SN
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


func RcvSignIn(msg *Message)(interface{}, error) {
	if msg.Kind != SIGNIN {
		return nil, errors.New("message Kind indicates not a SIGNIN")
	}
	
	var signInMsg map[string]string
	if err := ParseRcvInterfaces(msg, &signInMsg); err!= nil {
		return nil, err
	}

	// check database and send back SignInAck
	// TODO: update number of node register, send by heartbeat
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

// received in SN
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

	// check database and send back SignUpAck
	// TODO: send register to other SN if success
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

// received in SN
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

//TODO: SN receive start or end, forward message to all ON and SN
// use NewMsgwithBytes will be efficient without en/decode msg
func RcvStartEnd_SN(msg *Message)(interface{}, error) {
	return nil, nil	
}

// When ON receive start or end, call app url, to inform app
func RcvStartEnd_ON(msg *Message)(interface{}, error) {
	if msg.Kind != STARTEND_ON {
		return nil, errors.New("message Kind indicates not a STARTEND_ON")
	}
	
	var startEndMsg map[string]string
	if err := ParseRcvInterfaces(msg, &startEndMsg); err!= nil {
		return nil, err
	}
	
	// end message
	if startEndMsg["type"] == "end_hack"{
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

func RcvSnRank(msg *Message) (interface{}, error) {
	var rankList [RankNum]Rank
	if msg.Kind == SN_RANK {
		buffer := bytes.NewBuffer(msg.Data)
		tmpdecoder := gob.NewDecoder(buffer)
		err := tmpdecoder.Decode(&rankList)
		if err != nil {
			return nil, errors.New("Unable to do conversion of data")
		} else {
			return rankList, nil
		}
	}

	return nil, errors.New("message Kind indicates not a SN_RANK")
}

func SendSnOnSubmit(msg *Message, data interface{}) error {
	var buffer bytes.Buffer
	if msg.Kind == SN_RANK {
		str, ok := data.(string)
		if !ok {
			return errors.New("data passed is not SN_ON_SUBMIT")
		}
		encoder := gob.NewEncoder(&buffer)
		err := encoder.Encode(str)
		if err != nil {
			return errors.New("unable to encode SN_ON_SUBMIT data")
		}
		msg.Data = buffer.Bytes()
		buffer.Reset()

		return nil
	}

	return errors.New("message Kind indicates not a SN_ON_SUBMIT")
}

func RcvSnOnSubmit(msg *Message) (interface{}, error) {
	var str string
	if msg.Kind == SN_RANK {
		buffer := bytes.NewBuffer(msg.Data)
		tmpdecoder := gob.NewDecoder(buffer)
		err := tmpdecoder.Decode(&str)
		if err != nil {
			return nil, errors.New("Unable to do conversion of data")
		} else {
			return str, nil
		}
	}

	return nil, errors.New("message Kind indicates not a SN_ON_SUBMIT")
}

func SendSnOnSignIn(msg *Message, data interface{}) error {
	var buffer bytes.Buffer
	if msg.Kind == SN_ON_SIGNIN {
		str, ok := data.(string)
		if !ok {
			return errors.New("data passed is not SN_ON_SIGNIN")
		}
		encoder := gob.NewEncoder(&buffer)
		err := encoder.Encode(str)
		if err != nil {
			return errors.New("unable to encode SN_ON_SIGNIN data")
		}
		msg.Data = buffer.Bytes()
		buffer.Reset()

		return nil
	}

	return errors.New("message Kind indicates not a SN_ON_SIGNIN")
}

func RcvSnOnSignIn(msg *Message) (interface{}, error) {
	var str string
	if msg.Kind == SN_ON_SIGNIN {
		buffer := bytes.NewBuffer(msg.Data)
		tmpdecoder := gob.NewDecoder(buffer)
		err := tmpdecoder.Decode(&str)
		if err != nil {
			return nil, errors.New("Unable to do conversion of data")
		} else {
			return str, nil
		}
	}

	return nil, errors.New("message Kind indicates not a SN_ON_SIGNIN")
}
