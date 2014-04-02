package main

import (
	"encoding/json"
	"fmt"
	"net"
	"msg"
	"time"
)

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

func handleSignIn(message map[string]string) string {
	// msg = {"type": "sign_in", "username": name, "password": password}
	if name, exist := message["username"]; exist {
		if _, exist := message["password"]; exist {
			// send map[string]string messages to SN
			sendoutMsg := new(msg.Message)
			sendoutMsg.Kind = msg.SIGNIN

			err := msg.Handlers[sendoutMsg.Kind].Encode(sendoutMsg, message)
			if err != nil {
				fmt.Println(err);
			}

			// send message to SN
			msg.MsgPasser.Send(sendoutMsg, true)
			
			// busy waiting for rcv
			var val string
			for {
				if val, exist = msg.SignInMap[name]; exist {
					delete(msg.SignInMap, name)
					break;
				}
				time.Sleep(20)
			}		
			return val
		}
	}
	return "Message Error"
}

func handleSignUp(message map[string]string) string {
	//  msg = {"type": "sign_up", "username": name, "password": password, "email": email}
	if _, exist := message["username"]; exist {
		if _, exist := message["password"]; exist {
			if email, exist := message["email"]; exist {
				// send map[string]string messages to SN
				sendoutMsg := new(msg.Message)
				sendoutMsg.Kind = msg.SIGNUP

				err := msg.Handlers[sendoutMsg.Kind].Encode(sendoutMsg, message)
				if err != nil {
					fmt.Println(err);
				}

				// send message to SN
				msg.MsgPasser.Send(sendoutMsg, true)
			
				// busy waiting for rcv
				var val string
				for {
					if val, exist = msg.SignUpMap[email]; exist {
						delete(msg.SignUpMap, email)
						break;
					}
					time.Sleep(20)
				}		
				return val
			}
		}
	}
	return "Message Error"
}

func handleSuccess(msg map[string]string) string {
	// TODO: merge local_info
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
