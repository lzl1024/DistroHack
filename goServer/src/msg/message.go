package msg

import (
	"bytes"
)

type Message struct {
	Src, Dest string
	Seqnum int
	Kind int
	Data []byte
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
	s := "dest: " + msg.Dest + " " + " src: " + msg.Src + " data: " + bytes.NewBuffer(msg.Data[0:]).String()
	return s
}