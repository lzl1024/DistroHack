package msg

import (
	"bytes"
	"time"
)

/* Represents message types */
const (
	STRING = iota
	// add types below

	// Receive sign up from ordinary node
	SN_ONSIGNUP
	SIGNUPACK
	// Tell other sn the sign up of an ordinary node
	SN_MSIGNUP
	// Receive sign in from ordinary node
	SN_ONSIGNIN
	SIGNINACK
	// Receive information request from ordinary node
	SN_ASKINFO
	ASKINFOACK
	// Receive rank update from other super node
	SN_RANK
	// Receive public success from ordinary node
	SN_PBLSUCCESS
	// Receive START or END from ordinary
	SN_STARTENDON
	// Receive START or END from sn
	SN_STARTEND
	// Receive connect msg from ordinary node
	SN_NODEJOIN
	// Super Node Join from another super Node
	SN_JOIN
	// Send change in super node list to other super nodes
	SN_SNLISTUPDATE

	GROUPINFO
	STARTEND
	NUMTYPES
)

type Message struct {
	Src, Dest string
	Seqnum    int32
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
