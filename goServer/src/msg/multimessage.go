package msg

import (
	"bytes"
)

type MultiCastMessage struct {
	Message
	Origin string
}

func (msg *MultiCastMessage) NewMcastMsgwithBytes(dest string, kind int, data *bytes.Buffer) {
	tmpMsg := &msg.Message
	tmpMsg.NewMsgwithBytes(dest, kind, data)
}

func (msg *MultiCastMessage) NewMCastMsgwithData(dest string, kind int, data interface{}) error {
	tmpMsg := &msg.Message
	return tmpMsg.NewMsgwithData(dest, kind, data)
}

func (msg *MultiCastMessage) CopyMCastMsg(m *MultiCastMessage) {
	tmpMsg := &msg.Message
	tmpMsg.CopyMsg(&m.Message)
	msg.Origin = m.Origin
}

func (msg MultiCastMessage) String() string {
	s := msg.Message.String()
	return s + msg.Origin
}
