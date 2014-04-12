package msg

import (
	"bytes"
)

type MultiCastMessage struct {
	Message
	HostList []string
}

func (msg *MultiCastMessage) NewMcastMsgwithBytes(dest string, kind int, data *bytes.Buffer) {
	tmpMsg := &msg.Message
	tmpMsg.NewMsgwithBytes(dest, kind, data)
	msg.HostList = make([]string,0)
}

func (msg *MultiCastMessage) NewMCastMsgwithData(dest string, kind int, data interface{}) error {
	msg.HostList = make([]string,0)
	tmpMsg := &msg.Message
	return tmpMsg.NewMsgwithData(dest, kind, data)
}

func (msg *MultiCastMessage) CopyMCastMsg(m *MultiCastMessage) {
	tmpMsg := &msg.Message
	tmpMsg.CopyMsg(&m.Message)
	msg.HostList = make([]string, 0)
	for i := range m.HostList {
		msg.HostList = append(msg.HostList, m.HostList[i])
	}
}

func (msg MultiCastMessage) String() string {
	s := msg.Message.String()
	return s
}
