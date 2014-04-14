package msg

import (
	"bytes"
	"time"
	"strconv"
)

/* Represents message types */
const (
	STRING = iota
	// SN to SN
	SN_SN_SIGNUP		// Tell other sn the sign up of an ordinary node
	SN_SN_STARTEND		// Receive START or END from another SN 
	SN_SN_RANK			// Receive rank update from other super node
	
	// SN to ON	
	SN_ON_SIGNIN_ACK 	// ON Receivce the sign in status msg from SN
	SN_ON_SIGNUP_ACK	// ON Receivce the sign up status msg from SN
	SN_ON_ASKINFO_ACK	// SN ACK the global_ranking and local_info to ON
	SN_ON_STARTEND		// ON get start/end msg from SN
	SN_ON_RANK			// SN send new rank to ON
		
	// ON to SN
	ON_SN_SIGNUP		// Receive sign up from ordinary node
	ON_SN_SIGNIN		// Receive sign in from ordinary node
	ON_SN_PBLSUCCESS	// Receive public success from ordinary node
	ON_SN_ASKINFO		// Receive information request from ordinary node
	ON_SN_STARTEND		// Receive START or END from ordinary
	
	// TO BE IMPLE
	SN_NODEJOIN			// Receive connect msg from ordinary node
	SN_SNLISTUPDATE		// Send change in super node list to other super nodes
	GROUPINFO
	SN_JOIN				// Super Node Join from another super Node

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
	s := "dest: " + msg.Dest + " " + " src: " + msg.Src + " kind: " + strconv.Itoa(msg.Kind) + " timeStamp: " +
		msg.TimeStamp.String()
	return s
}
