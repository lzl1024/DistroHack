package msg

import (
	"fmt"
	"net"
	"time"
)

func TestMessagePasser() {
	channel := make(chan error)

	go clientTestThread(MsgPasser, channel)
	value := <- channel
	fmt.Println(value)
}

func clientTestThread(mp *Messagepasser, c chan error) {
	var ip string
	for {
		fmt.Println("Enter who you would like to connect to")
		fmt.Scanf("%s", &ip)
		if net.ParseIP(ip) == nil {
			fmt.Println("Invalid ip: try again")
			continue
		}
		
		msg1 := new(Message)
		msg1.Dest = ip
		msg1.Kind = STRING
		err:= Handlers[msg1.Kind].Encode(msg1, "ashish kaila")
		if err != nil {
			fmt.Println(err);
			continue
		}
		mp.Send(msg1, false)
		
		time.Sleep(20)
		
		msg2 := new(Message)
		msg2.Dest = ip
		msg2.Kind = PBLSUCCESS
		mapData := map[string]string{
			"1": "1",
			"2": "2",
		}
		err = Handlers[msg2.Kind].Encode(msg2, mapData)
		if err != nil {
			fmt.Println(err);
			continue
		}
		mp.Send(msg2, false)
	}
}
