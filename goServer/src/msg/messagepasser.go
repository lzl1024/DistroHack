package msg

import (
	"net"
	"fmt"
	"os"
	"encoding/gob"
	"time"
	"util"
)

type Connection struct {
	conn net.Conn
	encoder *gob.Encoder
}

type Messagepasser struct {
	Hostlist []string
	Connmap map[string]Connection
	ServerIP string
	ServerPort int
	drift time.Duration
}


var MsgPasser *Messagepasser

/* name is the IP address in string format */
func NewMsgPasser(serverIP string, serverPort int) (*Messagepasser, error) {
	var ts *time.Time
	var err error
	var retry int = 0
	
	mp := new(Messagepasser)
	if net.ParseIP(serverIP) == nil {
		fmt.Println("Invalid IP address")
		os.Exit(-1)
	}
	mp.ServerIP = serverIP
	mp.Connmap = make(map[string]Connection)
	mp.ServerPort = serverPort
	
	for ;retry != 3; {
		ts,err = util.Time()
		if err != nil {
			retry = retry + 1
		}
		break
	}
	
	if retry > 2 {
		return nil,err
	}
	
	refTime := *ts
	curTime := time.Now()
	fmt.Println("current: " + curTime.String())
	fmt.Println("reftime: " + refTime.String())
	if refTime.Before(curTime) {
		mp.drift = -1 * curTime.Sub(refTime)
	} else {
		mp.drift = refTime.Sub(curTime)
	}
	fmt.Println("Duration : " + mp.drift.String())
	
	return mp, nil;
}

func (mp *Messagepasser) Send(msg *Message) error{
	var encoder *gob.Encoder
	var err error
	var conn net.Conn
	
	msg.Src = mp.ServerIP
	// TODO:!!
	msg.TimeStamp = time.Now().Add(mp.drift)
	
	/* check if already existent connection is there */
	port := fmt.Sprint(mp.ServerPort)
	dest := net.JoinHostPort(msg.Dest, port)
	connection, ok := mp.Connmap[msg.Dest]
	if !ok {
		conn,err = net.Dial("tcp", dest)
		if err != nil {
			fmt.Println("error connecting to: ", dest, "reason: ", err)
			connection,ok := mp.Connmap[msg.Dest]
			if ok {
				connection.conn.Close();
				delete(mp.Connmap, msg.Dest)
			}
			return err
		}
		fmt.Println("adding a new connection to: ", dest)
		var tcpconn *net.TCPConn
		tcpconn, ok = conn.(*net.TCPConn)
		if ok {
			err = tcpconn.SetKeepAlive(true)
			if err != nil {
				fmt.Println("cannot set keepalive on connection")
			}
			
			err = tcpconn.SetLinger(0)
			if err != nil {
				fmt.Println("cannot set linger options")
			}
		}
		encoder = gob.NewEncoder(conn)
		connection.conn = conn
		connection.encoder = encoder
		mp.Connmap[msg.Dest] = connection
	} else {
		fmt.Println("Re-using connection to: ", dest)
		encoder = connection.encoder
	}
	
	err = encoder.Encode(msg)
	if err != nil {
		fmt.Println("error encoding data: ", err)
		connection,ok := mp.Connmap[msg.Dest]
		if ok {
			connection.conn.Close();
			delete(mp.Connmap, msg.Dest)
		}
		return err
	}
	
	return nil
}

/* based on message types take action */
func (mp *Messagepasser) DoAction(msg *Message) {
	str, err := Handlers[msg.Kind].Decode(msg)
	if err != nil {
		return
	}
	fmt.Println((*msg).String())
	fmt.Println(str)
}
