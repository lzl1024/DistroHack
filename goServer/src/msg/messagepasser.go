package msg

import (
	"net"
	"fmt"
	"os"
	"encoding/gob"
)

type Messagepasser struct {
	Hostlist []string
	Connmap map[string]net.Conn
	ServerIP string
	ServerPort int
}

func NewMsgPasser(name string, Server_PORT int) *Messagepasser {
	mp := new(Messagepasser)
	if net.ParseIP(name) == nil {
		fmt.Println("Invalid IP address")
		os.Exit(-1)
	}
	mp.ServerIP = name
	mp.Connmap = make(map[string]net.Conn)
	mp.ServerPort = Server_PORT
	
	return mp;
}

func (mp *Messagepasser) Send(msg *Message) error{
	msg.Src = mp.ServerIP
	
	/* check if already existent connection is there */
	conn, ok := mp.Connmap[msg.Dest]
	var err error
	if !ok {
		service := fmt.Sprintf("%s:%d", msg.Dest, mp.ServerPort)
		conn,err = net.Dial("tcp", service)
		if err != nil {
			fmt.Println("error connecting to: ", msg.Dest, "reason: ", err)
			delete(mp.Connmap, msg.Dest)
			return err
		}
		fmt.Println("adding a new connection to: ", service)
		mp.Connmap[msg.Dest] = conn
	} else {
		fmt.Println("Re-using connection to: ", msg.Dest)
	}
	
	encoder := gob.NewEncoder(conn)
	err = encoder.Encode(msg)
	if err != nil {
		fmt.Println("error encoding data: ", err)
		delete(mp.Connmap, msg.Dest)
		return err
	}
	
	return nil
}

