package msg

import (
	"encoding/gob"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
	"util"
)

type Connection struct {
	Conn    net.Conn
	encoder *gob.Encoder
}

// TODO: finally should be set SuperNodeIP =""
var SuperNodeIP = "10.0.1.17"
var rcvdlistMutex = &sync.Mutex{}

var BusyWaitingSleepInterval = time.Millisecond * time.Duration(100)
var BusyWaitingTimeOutRound = 100

type Messagepasser struct {
	SNHostlist       map[string]string
	SNLoadlist       map[string]int
	ONHostlist       map[string]string
	Connmap          map[string]Connection
	ConnMutex		 sync.Mutex
	ServerIP         string
	ONPort           int
	SNPort           int
	Drift            time.Duration
	IncomingMsg      chan Message
	IncomingMCastMsg chan MultiCastMessage
	RcvdMCastMsgs    []*MultiCastMessage
	SeqNum           int32
}

var MsgPasser *Messagepasser
var DLock *DsLock

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
	mp.IncomingMsg = make(chan Message)
	mp.IncomingMCastMsg = make(chan MultiCastMessage)
	mp.ONPort = ONPort
	mp.SNPort = SNPort
	mp.ONHostlist = make(map[string]string)
	mp.SNHostlist = make(map[string]string)
	mp.SNLoadlist = make(map[string]int)
	
	// sign up commit register
	signUp_requestMap = make(map[string]*SignUpCommitStatus)
	SignUp_commit_readySet = make(map[string]string) 

	mp.RcvdMCastMsgs = make([]*MultiCastMessage, 0)
	mp.SeqNum = 0

	for retry != 3 {
		ts, err = util.Time()
		if err != nil || ts == nil {
			retry = retry + 1
		}
		break
	}

	if ts == nil {
		fmt.Println("NTP server busy, use local time!")
		mp.Drift = 0;
	} else {
		refTime := *ts
		curTime := time.Now()
		fmt.Println("current: " + curTime.String())
		fmt.Println("reftime: " + refTime.String())
		if refTime.Before(curTime) {
			mp.Drift = -1 * curTime.Sub(refTime)
		} else {
			mp.Drift = refTime.Sub(curTime)
		}
		fmt.Println("Duration : " + mp.Drift.String())
	}

	go mp.RcvMessage()
	go mp.RcvMCastMessage()

	return mp, nil
}

func (mp *Messagepasser) getConnection(msgDest string, port string) (*Connection, error) {
	/* check if already existent connection is there */
	dest := net.JoinHostPort(msgDest, port)
	fmt.Println(dest)
	connection, ok := mp.Connmap[msgDest]
	if !ok {
		conn, err := net.DialTimeout("tcp", dest, (time.Duration(3) * time.Second))
		if err != nil {
			fmt.Println("error connecting to: ", dest, "reason: ", err)
			connection, ok := mp.Connmap[msgDest]
			if ok {
				connection.Conn.Close()
				delete(mp.Connmap, msgDest)
			}
			return nil, err
		}
		fmt.Println("MessagePasser: adding a new connection to ", dest)
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
		connection.Conn = conn
		connection.encoder = encoder
		mp.Connmap[msgDest] = connection
		return &connection, nil
	} else {
		//fmt.Println("Re-using connection to: ", dest)
	}

	return &connection, nil
}

func (mp *Messagepasser) actuallySend(connection *Connection, dest string, msg interface{}) error {
	//fmt.Println("MessagePasser: actuallySend")

	encoder := connection.encoder
	err := encoder.Encode(&msg)
	if err != nil {
		fmt.Println("MessagePasser actuallySend: error encoding data: ", err)
		mp.ConnMutex.Lock()
		connection, ok := mp.Connmap[dest]
		if ok {
			connection.Conn.Close()
			delete(mp.Connmap, dest)
		}
		mp.ConnMutex.Unlock()
		return err
	}

	return nil
}

func (mp *Messagepasser) Send(msg *Message) error {
	var port string
	var dest string

	msg.Origin = mp.ServerIP
	msg.Src = mp.ServerIP
	msg.TimeStamp = time.Now().Add(mp.Drift)
	
	fmt.Println("Send Message: ", msg.String())

	port = fmt.Sprint(mp.ONPort)
	mp.ConnMutex.Lock()
	connection, err := mp.getConnection(msg.Dest, port)
	mp.ConnMutex.Unlock()
	if err != nil {
		fmt.Println("MessagePasser Send: Error getting connection")
		return err
	}

	err = mp.actuallySend(connection, dest, msg)
	return err
}

func (mp *Messagepasser) SendMCast(msg *MultiCastMessage) {
	msg.Src = mp.ServerIP
	msg.TimeStamp = time.Now().Add(mp.Drift)
	
	fmt.Println("Send Multicast Message: ", msg.String())

	for e := range msg.HostList {
		host := msg.HostList[e]
		msg.Dest = host
		mp.ConnMutex.Lock()
		connection, err := mp.getConnection(host, fmt.Sprint(mp.ONPort))
		mp.ConnMutex.Unlock()
		//fmt.Println("MessagePasser : Sending message to ", host)
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

/* based on message types take action */
func (mp *Messagepasser) DoAction(msg *Message) {
	fmt.Println("MessagePasser DoAction :", (*msg).String())
	str, err := Handlers[msg.Kind](msg)
	if err != nil {
		fmt.Println("DO ACTION ERROR: ", err)
		return
	}

	fmt.Println("Receive data: ", str)
}

func (mp *Messagepasser) HandleMCast(msg *MultiCastMessage) {
	if strings.EqualFold(msg.Origin, mp.ServerIP) {
		return
	}
	newMCastMsg := new(MultiCastMessage)
	newMCastMsg.CopyMCastMsg(msg)
	mp.SendMCast(newMCastMsg)
}

// truely send out data to app
func SendtoApp(urlAddress string, data string) {
	_, err := http.PostForm(urlAddress,
		url.Values{"data": {data}})

	if err != nil {
		fmt.Println("Post failure: " + urlAddress + "," + data)
	}
}

func (mp *Messagepasser) RcvMessage() {
	for {
		msg := <-mp.IncomingMsg
		go mp.DoAction(&msg)
	}
}

func (mp *Messagepasser) RcvMCastMessage() {
	var v bool
	for {
		msg := <-mp.IncomingMCastMsg
		//fmt.Println("MessagePasser: A Multicast Message received", msg.Origin, msg.Seqnum)
		rcvdlistMutex.Lock()
		v = mp.isAlreadyRcvd(&msg)
		rcvdlistMutex.Unlock()
		if v == false {
			//fmt.Println("Never rcvd")
			mp.HandleMCast(&msg)
			go mp.DoAction(&msg.Message)
		} else {
			//fmt.Println("MessagePasser: The message has been seen before so moving on")
		}
	}
}

func (mp *Messagepasser) isAlreadyRcvd(msg *MultiCastMessage) bool {
	for e := range mp.RcvdMCastMsgs {
		if mp.RcvdMCastMsgs[e] != nil && msg.Seqnum == mp.RcvdMCastMsgs[e].Seqnum && 
		strings.EqualFold(msg.Origin, mp.RcvdMCastMsgs[e].Origin) == true {
			return true
		}
	}
	mp.RcvdMCastMsgs = append(mp.RcvdMCastMsgs, msg)
	return false
}

func (mp *Messagepasser) RefreshAlreadyRcvdlist(src string) {
	rcvdlistMutex.Lock()
	for e := range mp.RcvdMCastMsgs {
		if mp.RcvdMCastMsgs[e] != nil && strings.EqualFold(src, mp.RcvdMCastMsgs[e].Origin) {
			mp.RcvdMCastMsgs[e] = nil
		}
	}
	rcvdlistMutex.Unlock()
}
