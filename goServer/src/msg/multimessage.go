package msg

import (
	"bytes"
	"fmt"
)

type MultiCastMessage struct {
	Message
	HostList map[string]string
}

func (msg *MultiCastMessage) NewMcastMsgwithBytes(dest string, kind int, data *bytes.Buffer) {
	tmpMsg := &msg.Message
	tmpMsg.NewMsgwithBytes(dest, kind, data)
	msg.HostList = make(map[string]string)
}

func (msg *MultiCastMessage) NewMCastMsgwithData(dest string, kind int, data interface{}) error {
	msg.HostList = make(map[string]string)
	tmpMsg := &msg.Message
	return tmpMsg.NewMsgwithData(dest, kind, data)
}

func (msg *MultiCastMessage) CopyMCastMsg(m *MultiCastMessage) {
	tmpMsg := &msg.Message
	tmpMsg.CopyMsg(&m.Message)
	msg.HostList = make(map[string]string)
	for k,_ := range m.HostList {
		msg.HostList[k] = m.HostList[k]
	}
}

func (msg MultiCastMessage) String() string {
	s := msg.Message.String() + fmt.Sprintf("%s", msg.HostList)
	return s
}
