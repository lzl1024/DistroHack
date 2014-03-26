package main

import (
	"fmt"
    "net"
    "encoding/json"
)

var userMap = map[string] string{
    "1": "111111",
    "2": "222222",
    "admin": "admin",
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
    
    fmt.Println(msg)

    // handle different type of message
    switch msgType {
    case "sign_in":
    	conn.Write([]byte(handleSignIn(msg)))
    case "sign_up":
    	conn.Write([]byte(handleSignUp(msg)))
    case "submit_success":
    	handleSuccess(msg)
    default:
    	fmt.Println("Messge Type Undefined")
    }
}

func handleSignIn(msg map[string]string) string{
	// name, password match
	// msg = {"type": "sign_in", "username": name, "password": password}
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

func handleSignUp(msg map[string]string) string{
	//  msg = {"type": "sign_up", "username": name, "password": password, "email": email}
	if name, exist := msg["username"]; exist {
		// check user in userMap
		if _, exist := userMap[name]; exist {
			return "This email/username has already been registered!"
		}	
		if password, exist := msg["password"]; exist {
			userMap[name] = password
			return "success"
		}	
	}
	
	return "Message Error"
}

func handleSuccess(msg map[string]string) {
	// msg = {"type": "submit_success", "username": user, "pid": problem_id}
	// TODO: send success message to other severs
	
}
