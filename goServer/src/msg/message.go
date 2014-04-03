package msg

import (
	"bytes"
	"time"
	"util"
)

/* Represents message types */
const (
	STRING = iota
	// add types below

	// Receive rank update from other super node
	SN_RANK
	// Receive submit success from ordinary node
	SN_ON_SUBMIT
	// Receive sign in from ordinary node
	SN_ON_SIGNIN

	NUMTYPES
)

type Message struct {
	Src, Dest string
	Seqnum    int
	Kind      int
	Data      []byte
	TimeStamp time.Time
}

func (msg *Message) NewMsg(dest string, kind int, data *bytes.Buffer) {
	msg.Dest = dest
	msg.Kind = kind
	msg.Data = data.Bytes()
}

func (msg *Message) CopyMsg(m *Message) {
	msg.Dest = m.Dest
	msg.Kind = m.Kind
	msg.Data = m.Data
	msg.Seqnum = m.Seqnum
	msg.Src = m.Src
}

func (msg Message) String() string {
	ts, _ := util.Time()
	time_stamp := *ts

	s := "dest: " + msg.Dest + " " + " src: " + msg.Src + " kind: " + " timeStamp: " +
		msg.TimeStamp.String() + " ref time: " + time_stamp.String()
	return s
}
