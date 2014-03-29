package msg

import (
	"fmt"
	"net"
)

func serverthread(mp *Messagepasser, c chan error) {
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

