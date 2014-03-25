package main

import (
	"fmt"
    "net"
    "encoding/json"
)

var userMap = map[string] string{
    "1": "111111",
    "2": "222222",
}

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
    }
}

func handleConnection(conn net.Conn) {
	rcvbuf := make([]byte,rcvBufLen)
	
    // receive messages
    size,err := conn.Read(rcvbuf)
    
    if err != nil {
    	fmt.Println("Read Error:",err.Error())
        return
    }

	var msg map[string]string
	
    err = json.Unmarshal(rcvbuf[:size],&msg)
  
    if err != nil {
    	fmt.Println("Unmarshal Error:",err.Error())
        return
    }
    
    msgType, exist := msg["type"]
    
    if exist == false {
    	fmt.Println("No Message Type!")
    	return
    }
    
    // handle different type of message
    switch msgType {
    case "sign_in":
    	conn.Write([]byte(handleSignIn(msg)))
    default:
    	fmt.Println("Messge Type Undefined")
    }
}

func handleSignIn(msg map[string]string) string{
	if name, exist := msg["username"]; exist {
		if password, exist := msg["password"]; exist {
			if realPassword, exist := userMap[name]; 
				exist && password == realPassword {
				return "success" 
			}
		}
	}
	return "failed"
}


