package main

import (
	"encoding/json"
	"fmt"
	"net"
	"msg"
	"time"
	"strconv"
	"util"
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

	var message map[string]string

	err = json.Unmarshal(rcvbuf[:size], &message)

	if err != nil {
		fmt.Println("Unmarshal Error:", err.Error())
		return
	}

	msgType, exist := message["type"]

	if exist == false {
		fmt.Println("No Message Type!")
		return
	}

	fmt.Println(message)

	// handle different type of message
	switch msgType {
	case "sign_in":
		conn.Write([]byte(handleSignIn(message)))
	case "sign_up":
		conn.Write([]byte(handleSignUp(message)))
	case "submit_success":
		conn.Write([]byte(handleSuccess(message)))
	case "end_hack":
		conn.Write([]byte(handleEnd()))
	case "start_hack":
		conn.Write([]byte(handleStart()))
	case "problem_id":
		conn.Write([]byte(strconv.Itoa(
			msg.Local_info[message["username"]].Score + 1)))
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
			
			// add user into local info
			if _, exist := msg.Local_info[name]; !exist {
				msg.Local_info[name] = msg.User_record{
					name, 0, time.Now(),
				}
			}	
			return val
		}
	}
	return "Message Error"
}

func handleSignUp(message map[string]string) string {
	//  msg = {"type": "sign_up", "username": name, "password": password, "email": email}
	if name, exist := message["username"]; exist {
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
				
				// add user into local info
				ntpTime, err := util.Time()
				if err != nil {
					fmt.Println(err)
				}
				
				msg.Local_info[name] = msg.User_record{
					name, 0, *ntpTime,
				}	
				return val
			}
		}
	}
	return "Message Error"
}

func handleSuccess(message map[string]string) string {
	// msg = {"type": "submit_success", "username": user, "pid": problem_id}
	// merge local_info
	pid, _ := strconv.Atoi(message["pid"])
	name := message["username"]
	
	if msg.Local_info[name].Score < pid {
		msg.Local_info[name] = msg.User_record{
			UserName: name, Score: pid, Ctime: time.Now()}
				
		// TODO: send success message to other servers
		// send map[string]string messages to SN
		sendoutMsg := new(msg.Message)
		sendoutMsg.Kind = msg.PBLSUCCESS

		err := msg.Handlers[sendoutMsg.Kind].Encode(sendoutMsg, message)
		if err != nil {
			fmt.Println(err);
		}

		// send message to SN
		msg.MsgPasser.Send(sendoutMsg, true)
	}
	return "success"
}

func handleEnd() string {
	// TODO: multicast or send to SN end hackthon message to other servers
	return "success"
}

func handleStart() string {
	// TODO: multicast or send to SN start hackthon message to other servers
	return "success"
}
