package msg

import (
	"encoding/gob"
	"fmt"
	"net"
	"os"
	"time"
	"util"
	"net/http"
	"net/url"
	"container/list"
	"strings"
)

type Connection struct {
	conn    net.Conn
	encoder *gob.Encoder
}

// TODO: change it!
var SuperNodeIP = "127.0.0.1"

type Messagepasser struct {
	Hostlist []string
	Connmap map[string]Connection
	ServerIP string
	ONPort int
	SNPort int
	drift time.Duration
}

var MsgPasser *Messagepasser

/* name is the IP address in string format */
func NewMsgPasser(serverIP string, ONPort int, SNPort int) (*Messagepasser, error) {
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
	mp.ONPort = ONPort
	mp.SNPort = SNPort
	
	for ;retry != 3; {
		ts,err = util.Time()
		if err != nil || ts == nil{
			retry = retry + 1
		}
		break
	}

	if ts == nil {
		return nil, err
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

	return mp, nil
}

func (mp *Messagepasser) getConnection(msgDest string, port string) (*Connection, error) {
	/* check if already existent connection is there */
	dest := net.JoinHostPort(msgDest, port)
	connection, ok := mp.Connmap[msgDest]
	if !ok {
		conn, err := net.Dial("tcp", dest)
		if err != nil {
			fmt.Println("error connecting to: ", dest, "reason: ", err)
			connection, ok := mp.Connmap[msgDest]
			if ok {
				connection.conn.Close()
				delete(mp.Connmap, msgDest)
			}
			return nil, err
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
		encoder := gob.NewEncoder(conn)
		connection.conn = conn
		connection.encoder = encoder
		mp.Connmap[msgDest] = connection
		return &connection, nil
	} else {
		fmt.Println("Re-using connection to: ", dest)
	}
	
	return &connection, nil
}

func (mp *Messagepasser) actuallySend(connection *Connection, dest string, msg interface{}) error {
	encoder := connection.encoder
	err := encoder.Encode(&msg)
	if err != nil {
		fmt.Println("error encoding data: ", err)
		connection, ok := mp.Connmap[dest]
		if ok {
			connection.conn.Close()
			delete(mp.Connmap, dest)
		}
		return err
	}
	
	return nil
}

func (mp *Messagepasser) Send(msg *Message, isSN bool) error{
	var port string
	var dest string
	
	msg.Src = mp.ServerIP
	msg.TimeStamp = time.Now().Add(mp.drift)
	
	// check destination
	if (isSN) {
		msg.Dest = SuperNodeIP
		port = fmt.Sprint(mp.SNPort)
	} else {
		port = fmt.Sprint(mp.ONPort)
	}
	
	connection, err := mp.getConnection(msg.Dest, port)
	if err != nil {
		fmt.Println("Error getting connection")
		return err
	}
	
	err = mp.actuallySend(connection, dest, msg)
	
	return nil
}

func (mp *Messagepasser) SendMCast(msg *MultiCastMessage, hostlist *list.List) {
	for e := hostlist.Front(); e != nil; e = e.Next() {
		host := e.Value.(string)
		if strings.EqualFold(host, mp.ServerIP) == false {
			msg.Dest = host
			msg.TimeStamp = time.Now().Add(mp.drift)
			connection, err := mp.getConnection(host, fmt.Sprint(mp.ONPort))
			if err != nil {
				fmt.Println("Error getting connection to host:", host)
				/* try to send to atleast 1 person in the list */
				continue
			}
			/* try to send to atleast 1 person in the list */
			err = mp.actuallySend(connection, host, msg)
			if err != nil {
				fmt.Println("Unable to send message to host:", host)
			}
		}
	}
}

/* based on message types take action */
func (mp *Messagepasser) DoAction(msg *Message) {
	str, err := Handlers[msg.Kind](msg)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println((*msg).String())
	fmt.Println(str)
}

func (mp *Messagepasser) HandleMCast(msg *MultiCastMessage) {
	if strings.EqualFold(msg.Origin, mp.ServerIP) {
		return
	}
	newMCastMsg := new(MultiCastMessage)
	newMCastMsg.CopyMCastMsg(msg)
	hostlist := list.New()
	hostlist.PushBack("128.237.227.84")
	hostlist.PushBack("128.2.210.206")
	mp.SendMCast(newMCastMsg, hostlist)
}

// truely send out data to app
func SendtoApp(urlAddress string, data string) {
	_, err := http.PostForm(urlAddress,
		url.Values{"data": {data}})

	if err != nil {
		fmt.Println("Post failure: " + urlAddress + "," + data)
	}
}
