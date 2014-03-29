package main

import (
	"encoding/json"
	"fmt"
	"net"
	"util"
)

var userMap = map[string]string{
	"1":     "111111",
	"2":     "222222",
	"admin": "admin",
}

func HandleConnectionFromLocal(listener net.Listener) {

	for {
		conn, err := listener.Accept()

		if err != nil {
			fmt.Println("Accept Error:", err.Error())
			return
		}

		// new thread to handle request
		go handleConnectionFromLocalThread(conn)
	}
}

func handleConnectionFromLocalThread(conn net.Conn) {
	rcvbuf := make([]byte, rcvBufLen)

	// receive messages
	size, err := conn.Read(rcvbuf)

	if err != nil {
		fmt.Println("Read Error:", err.Error())
		return
	}

	var msg map[string]string

	err = json.Unmarshal(rcvbuf[:size], &msg)

	if err != nil {
		fmt.Println("Unmarshal Error:", err.Error())
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
		conn.Write([]byte(handleSuccess(msg)))
	case "end_hack":
		conn.Write([]byte(handleEnd()))
	case "start_hack":
		conn.Write([]byte(handleStart()))
	default:
		fmt.Println("Messge Type Undefined")
	}
}

func handleSignIn(msg map[string]string) string {
	// msg = {"type": "sign_in", "username": name, "password": password}
	if name, exist := msg["username"]; exist {
		if password, exist := msg["password"]; exist {
			/*if realPassword, exist := userMap[name]; exist && password == realPassword {
			//	return "success"
			}*/
			return util.DatabaseSignIn(name, password)
		}
	}
	return "Message Error"
}

func handleSignUp(msg map[string]string) string {
	//  msg = {"type": "sign_up", "username": name, "password": password, "email": email}
	if name, exist := msg["username"]; exist {
		if password, exist := msg["password"]; exist {
			if email, exist := msg["email"]; exist {
				return util.DatabaseSignUp(name, password, email)
			}
		}
		/*if _, exist := userMap[name]; exist {
			return "This email/username has already been registered!"
		}
		if password, exist := msg["password"]; exist {
			userMap[name] = password
			return "success"
		}*/
	}
	return "Message Error"
}

func handleSuccess(msg map[string]string) string {
	// msg = {"type": "submit_success", "username": user, "pid": problem_id}
	// TODO: send success message to other servers

	return "success"
}

func handleEnd() string {
	// TODO: send end hackthon message to other servers
	return "success"
}

func handleStart() string {
	// TODO: send start hackthon message to other servers
	return "success"
}
