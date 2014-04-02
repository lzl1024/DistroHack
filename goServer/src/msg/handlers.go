package msg

import (
	"errors"
)

var Handlers [NUMTYPES]Handler

type Sendhdlr func(*Message, interface{})error
type Rcvhdlr func(*Message)(interface{},error)

type Handler struct {
	Encode Sendhdlr
	Decode Rcvhdlr
}

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
