package main

import (
	"fmt"
	"net"
)

const ListenPort = ":4213"
const rcvBufLen = 1024

func main() {
	// open the listen port
	listener,err := net.Listen("tcp",ListenPort)
	
	if err != nil {
    	fmt.Println("Listener port has been used:",err.Error())
       	return
    }
	
	for {
		conn,err := listener.Accept()
	
    	if err != nil {
    		fmt.Println("Accept Error:",err.Error())
        	return
    	}
    
    	// new thread to handle request
    	go handleConnection(conn)
    	
    	// active connect to application
    	go activeThread()
    }
}
