package main

import (
	"fmt"
	"net"
	"msg"
	"encoding/gob"
)

func InitListenerForPeers() {
	channel := make(chan error)
	go serverthread(msg.MsgPasser, channel)
	value := <- channel
	fmt.Println(value)
}

func serverthread(mp *msg.Messagepasser, c chan error) {
	fmt.Println("Started server thread")
	service := fmt.Sprint(":", mp.ONPort)

	tcpAddr, err := net.ResolveTCPAddr("tcp", service)
	if err != nil {
		fmt.Println("ServerThread: Unrecoverable error trying to start go server")
		c <- err
		return
	}
	
	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		fmt.Println("ServerThread: Unrecoverable error trying to start listening on server ", err)
		c <- err
		return
	}
	
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("ServerThread: error accepting connection...continuing")
			continue
		}
		go rcvthread(mp, conn)
	}
}

func rcvthread(mp *msg.Messagepasser, conn net.Conn) {
	fmt.Println("Started recevier thread\n", conn.RemoteAddr().String())
	var tcpconn *net.TCPConn
	var ok bool
	var err error
	var data interface{}
	
	tcpconn, ok = conn.(*net.TCPConn)
	if ok {
		err = tcpconn.SetLinger(0)
		if err != nil {
			fmt.Println("RcvThread: cannot set linger options")
		}
	}
	
	decoder := gob.NewDecoder(conn)
	for {
		err := decoder.Decode(&data)
		if err != nil {
			fmt.Println("RcvThread: error while decoding: ", err)
			conn.Close()
			break
		}
		
		switch t := data.(type) {
			case msg.Message :
				mp.IncomingMsg <- t
			case msg.MultiCastMessage :
				mp.IncomingMCastMsg <- t
			default :
				fmt.Println("RcvThread: Issues are there, msg is not message or multicasemsg")
		}
	}
	
	/* remove the connection from all maps (SN/ON or Conn)*/
	mp.ConnMutex.Lock()
	dest,_,_ := net.SplitHostPort(conn.RemoteAddr().String()) 
	connection, ok := mp.Connmap[dest]
	if ok {
		fmt.Println("RcvThread: Removing connection to ", dest, " from connection map")
		connection.Conn.Close()
		delete(mp.Connmap, dest)
	}
	mp.ConnMutex.Unlock()
	
	_,ok = mp.SNHostlist[dest]
	if ok {
		fmt.Println("RcvThread: Removing entry to ", dest, " from SNHostList map")
		delete(mp.SNHostlist, dest)
		delete(mp.SNLoadlist, dest)
		msg.DLock.ResetLock(dest)
	}

	_,ok = mp.ONHostlist[dest]
	if ok {
		fmt.Println("RcvThread: Removing entry to ", dest, " from ONHostlist map")
		delete(mp.ONHostlist, dest)
		
		// for SN, Notify others of my load change
		if isSN {
			loadNotify := new(msg.Message)
			err := loadNotify.NewMsgwithData("", msg.SN_SN_LOADMERGE, len(mp.ONHostlist))
			if err != nil {
				fmt.Println("When ON failure: ", err)
				return
			}
	
			// send message to SNs
			msg.MulticastMsgInGroup(loadNotify, true)
		}
		
	}
	
	mp.RefreshAlreadyRcvdlist(dest)	
}

