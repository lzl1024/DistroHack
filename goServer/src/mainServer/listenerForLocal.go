package main

import (
	"encoding/json"
	"fmt"
	"msg"
	"net"
	"strconv"
	"time"
	"util"
)

func HandleConnectionFromLocal(listener net.Listener) {
	for {
		conn, err := listener.Accept()

		// if SN is under repairing, busy waiting until it got repaired
		for msg.SuperNodeIP == "" {
			time.Sleep(msg.BusyWaitingSleepInterval)
		}

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

	// handle different type of message
	switch msgType {
	case "sign_in":
		conn.Write([]byte(handleSignIn(message)))
	case "sign_up":
		conn.Write([]byte(handleSignUp(message)))
	case "submit_success":
		conn.Write([]byte(handleSuccess(message)))
	case "end_hack":
		conn.Write([]byte(handleEndandStart(message)))
	case "start_hack":
		conn.Write([]byte(handleEndandStart(message)))
	case "problem_id":
		name := message["username"]

		//TO DO read from DB
		score, e := util.DatabaseReadScore(name)
		tuple, exist := msg.Local_map[name]

		if score == -1 {
			if exist {
				score = tuple.Score
			}
			else {
				score = 0
			}
			util.DatabaseUpdateStage(name, score)			
		}

		// add user if he use session cache to login
		/*if !exist {
			score = 0
		} else {
			score = tuple.Score
		}*/

		conn.Write([]byte(strconv.Itoa(score + 1)))
	default:
		fmt.Println("Messge Type Undefined")
	}
}

func handleSignIn(message map[string]string) string {
	// msg = {"type": "sign_in", "username": name, "password": password}
	if _, exist := message["username"]; exist {
		if _, exist := message["password"]; exist {
			// send map[string]string messages to SN
			sendoutMsg := new(msg.Message)
			err := sendoutMsg.NewMsgwithData(msg.SuperNodeIP, msg.ON_SN_SIGNIN, message)
			if err != nil {
				fmt.Println(err)
			}

			// send message to SN
			msg.MsgPasser.Send(sendoutMsg)

			// channel waiting for rcv
			msg.SignInChan = make(chan string)
			val := <-msg.SignInChan
			return val
		}
	}
	return "Message Error"
}

func handleSignUp(message map[string]string) string {
	//  msg = {"type": "sign_up", "username": name, "password": password, "email": email}
	if _, exist := message["username"]; exist {
		if _, exist := message["password"]; exist {
			if _, exist := message["email"]; exist {
				// send map[string]string messages to SN
				sendoutMsg := new(msg.Message)
				err := sendoutMsg.NewMsgwithData(msg.SuperNodeIP, msg.ON_SN_SIGNUP, message)
				if err != nil {
					fmt.Println(err)
				}

				// send message to SN
				msg.MsgPasser.Send(sendoutMsg)

				// channel waiting for rcv
				msg.SignUpChan = make(chan string)
				val := <-msg.SignUpChan

				return val
			}
		}
	}
	return "Message Error"
}

func handleSuccess(message map[string]string) string {
	// msg = {"type": "submit_success", "username": user, "pid": problem_id, "time": time}
	// merge local_info
	pid, _ := strconv.Atoi(message["pid"])
	name := message["username"]
	submitTime, _ := time.Parse("2006-01-02T15:04:05.999999", message["time"])

	currentScore := msg.Local_map[name].Score
	score, e := util.DatabaseReadScore(name)
	if score < currentScore {
		score = currentScore
	}


	if score < pid {

		util.DatabaseUpdateScore(name, pid)

		sendOutRecord := msg.UserRecord{
			UserName: name, Score: pid, Ctime: submitTime.Add(msg.MsgPasser.Drift)}

		// TODO: send success message to other servers
		// send map[string]string messages to SN
		sendoutMsg := new(msg.Message)

		err := sendoutMsg.NewMsgwithData(msg.SuperNodeIP, msg.ON_SN_PBLSUCCESS, sendOutRecord)
		if err != nil {
			fmt.Println(err)
		}

		// send message to SN
		msg.MsgPasser.Send(sendoutMsg)
	}

	return "success"
}

func handleEndandStart(meesage map[string]string) string {
	// TODO: multicast or send to SN end hackthon message to other servers
	// send to SN
	sendoutMsg := new(msg.Message)

	err := sendoutMsg.NewMsgwithData(msg.SuperNodeIP, msg.ON_SN_STARTEND, meesage)
	if err != nil {
		fmt.Println(err)
	}

	// send message to SN
	msg.MsgPasser.Send(sendoutMsg)
	return "success"
}
