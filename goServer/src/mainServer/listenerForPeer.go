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
	

	// go through the SignUp_commit_readySet to clean up the commit coordinator status
	msg.SignUp_commitLock.Lock()
	for k, v := range msg.SignUp_commit_readySet {
		if v == dest {
			delete(msg.SignUp_commit_readySet, k)
		}
	}
	msg.SignUp_commitLock.Unlock()
	
	msg.SNHostlistMutex.Lock()
	// SN peer fails, only need to delete it from map
	_,ok = mp.SNHostlist[dest]
	if ok {
		fmt.Println("RcvThread: Removing entry to ", dest, " from SNHostList map")
		delete(mp.SNHostlist, dest)
		delete(mp.SNLoadlist, dest)
		msg.DLock.ResetLock(dest)
	}
	msg.SNHostlistMutex.Unlock()


	// ON fails, SN should notify other to change status
	_,ok = mp.ONHostlist[dest]
	if ok {
		msg.ONHostlistMutex.Lock()
		fmt.Println("RcvThread: Removing entry to ", dest, " from ONHostlist map")
		delete(mp.ONHostlist, dest)
		msg.ONHostlistMutex.Unlock()
		
		// for SN, Notify others of my load change
		if isSN {
			msg.ONHostlistMutex.Lock()
			loadNotify := new(msg.Message)
			err := loadNotify.NewMsgwithData("", msg.SN_SN_LOADMERGE, len(mp.ONHostlist))
			msg.ONHostlistMutex.Unlock()
			
			if err != nil {
				fmt.Println("When ON failure: ", err)
				return
			}
	
			// send message to SNs
			msg.MulticastMsgInGroup(loadNotify, true)
			
			msg.ONHostlistMutex.Lock()
			/* Send MCast to other ONs in the group */
			changeONList := new(msg.Message)
			err = changeONList.NewMsgwithData("", msg.SN_ON_CHANGEONLIST, mp.ONHostlist)
			msg.ONHostlistMutex.Unlock()
			if err != nil {
				fmt.Println("When ON failure: ", err)
				return
			}
	
			// send message to ONs
			msg.MulticastMsgInGroup(changeONList, false)
		}
		
	}
	
	// clear receive archive
	mp.RefreshAlreadyRcvdlist(dest)
	
	// ON's SN fails, improved bully algorithm
	if msg.SuperNodeIP == dest {
		msg.SNFailure()
	}
}

