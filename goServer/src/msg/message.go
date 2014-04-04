package msg

import (
	"bytes"
	"time"
)

/* Represents message types */
const (
	STRING = iota
	// add types below

	// Receive rank update from other super node
	SN_RANK
	// Receive sign in from ordinary node
	SN_SIGNIN
	// Receive public success from ordinary node
	SN_PBLSUCCESS
	// Receive connect msg from ordinary node
	SN_TOCONNECT

	PBLSUCCESS
	SIGNINACK
	SIGNUP
	SIGNUPACK
	STARTEND_ON
	STARTEND_SN
	NUMTYPES
)

type Message struct {
	Src, Dest string
	Seqnum    int
	Kind      int
	Data      []byte
	TimeStamp time.Time
}

func (msg *Message) NewMsgwithBytes(dest string, kind int, data *bytes.Buffer) {
	msg.Dest = dest
	msg.Kind = kind
	msg.Data = data.Bytes()
}

func (msg *Message) NewMsgwithData(dest string, kind int, data interface{}) error {
	msg.Dest = dest
	msg.Kind = kind
	return ParseSendInterfaces(msg, data)
}

func (msg *Message) CopyMsg(m *Message) {
	msg.Dest = m.Dest
	msg.Kind = m.Kind
	msg.Data = m.Data
	msg.Seqnum = m.Seqnum
	msg.Src = m.Src
}

func (msg Message) String() string {
	s := "dest: " + msg.Dest + " " + " src: " + msg.Src + " kind: " + " timeStamp: " +
		msg.TimeStamp.String()
	return s
}
