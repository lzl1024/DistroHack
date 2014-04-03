package msg

import (
	"fmt"
	"net"
)

const SERVER_PORT = 5989

func TestMessagePasser() {
	channel := make(chan error)

	go clientTestThread(MsgPasser, channel)
	value := <-channel
	fmt.Println(value)
}

func clientTestThread(mp *Messagepasser, c chan error) {
	var ip string
	msg := new(Message)
	for {
		fmt.Println("Enter who you would like to connect to")
		fmt.Scanf("%s", &ip)
		if net.ParseIP(ip) == nil {
			fmt.Println("Invalid ip: try again")
			continue
		}

		msg.Dest = ip
		msg.Kind = STRING
		err := Handlers[msg.Kind].Encode(msg, "ashish kaila")
		if err != nil {
			fmt.Println(err)
			continue
		}
		mp.Send(msg)
	}
}
