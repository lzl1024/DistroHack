package main

import (
	"fmt"
	"net"
	"msg"
	"encoding/gob"
)

func initConnectionFromPeers() {
	channel := make(chan error)
	go serverthread(msg.MsgPasser, channel)
	value := <- channel
	fmt.Println(value)
	
}

func serverthread(mp *msg.Messagepasser, c chan error) {
	fmt.Println("Started server thread")
	service := fmt.Sprint(":", mp.ServerPort)
	tcpAddr, err := net.ResolveTCPAddr("ip", service)
	if err != nil {
		fmt.Println("Unrecoverable error trying to start go server")
		c <- err
		return
	}
	
	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		fmt.Println("Unrecoverable error trying to start listening on server")
		c <- err
		return
	}
	
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("error accepting connection...continuing")
			continue
		}
		go rcvthread(mp, conn)
	}
}

func rcvthread(mp *msg.Messagepasser, conn net.Conn) {
	fmt.Println("Started recevier thread\n")
	var tcpconn *net.TCPConn
	var ok bool
	var err error
	var msg msg.Message
	
	tcpconn, ok = conn.(*net.TCPConn)
	if ok {
			err = tcpconn.SetLinger(0)
			if err != nil {
				fmt.Println("cannot set linger options")
			}
	}
	
	decoder := gob.NewDecoder(conn)
	for {
		err := decoder.Decode(&msg)
		if err != nil {
			fmt.Println("error while decoding: ", err)
			conn.Close()
			break
		}
		go mp.DoAction(&msg)
	}
}

