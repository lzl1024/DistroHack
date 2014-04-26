package msg

import (
	"bytes"
	"strconv"
	"time"
)

/* Represents message types */
const (
	STRING = iota
	// SN to SN
	SN_SN_SIGNUP        // ID: 1 Tell other sn the sign up of an ordinary node
	SN_SN_STARTEND      // ID: 2 Receive START or END from another SN
	SN_SN_RANK          // ID: 3 Receive rank update from other super node
	SN_SN_COMMIT_RD     // ID: 4 SN send Commit_ready to other SNs
	SN_SN_COMMIT_RD_ACK // ID: 5 SN workers send back commit_ready to master
	SN_SN_JOIN          // ID: 6 Super Node Join from another super Node
	SN_SN_JOIN_ACK      // ID: 7 Super Node join ack msg tell the new super node data is ready on server
	SN_SN_LISTUPDATE    // ID: 8 Send change in super node list to other super nodes
	SN_SN_LISTMERGE     // ID: 9 Merge list information
	SN_SN_LOADUPDATE    // ID: 10 Get the load information from supernodes
	SN_SN_LOADMERGE     // ID: 11 Merge all load inforamtion

	// SN to ON
	SN_ON_SIGNIN_ACK   // ID: 12 ON Receivce the sign in status msg from SN
	SN_ON_SIGNUP_ACK   // ID: 13 ON Receivce the sign up status msg from SN
	SN_ON_ASKINFO_ACK  // ID: 14 SN ACK the global_ranking and local_info to ON
	SN_ON_STARTEND     // ID: 15 ON get start/end msg from SN
	SN_ON_RANK         // ID: 16 SN send new rank to ON
	SN_ON_JOIN_ACK     // ID: 17 Ack Message when SN get ON's first join request: who you to connect
	SN_ON_CHANGEONLIST // ID: 18 SN send new ONList to ONs

	// ON to SN
	ON_SN_SIGNUP     // ID: 19 Receive sign up from ordinary node
	ON_SN_SIGNIN     // ID: 20 Receive sign in from ordinary node
	ON_SN_PBLSUCCESS // ID: 21 Receive public success from ordinary node
	ON_SN_ASKINFO    // ID: 22 Receive information request from ordinary node
	ON_SN_STARTEND   // ID: 23 Receive START or END from ordinary
	ON_SN_JOIN       // ID: 24 Receive connect msg from ordinary node
	ON_SN_REGISTER   // ID: 25 Register message from ON to SN

	// ON to ON : SN election
	ON_ON_ELECTION     // ID: 26 One ON want to be the leader
	ON_ON_LEADER       // ID: 27 One ON notify other to be the leader
	ON_ON_ELECTION_ACK // ID: 28 One ON ack the leader

	//Distributed Lock
	SN_SNLOCKREQ
	SN_SNLOCKREL
	SN_SNLOCKACK
	SN_SNACKRENEG

	NUMTYPES
)

type Message struct {
	Origin    string
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
	msg.Origin = m.Origin
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
