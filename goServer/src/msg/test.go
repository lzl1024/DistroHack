package msg

import (
	"fmt"
	"net"
	"time"
)

func TestMessagePasser() {
	channel := make(chan error)

	go clientTestThread(MsgPasser, channel)
	value := <-channel
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
		err := msg1.NewMsgwithData(ip, STRING, "ashish kaila")
		if err != nil {
			fmt.Println(err)
			continue
		}
		mp.Send(msg1, false)

		time.Sleep(20)

		msg2 := new(Message)
		userData := UserRecord{
			"akaila",
			100,
			time.Now(),
		}
		err = msg2.NewMsgwithData(ip, SN_PBLSUCCESS, userData)
		if err != nil {
			fmt.Println("Reached here:", err)
			continue
		}
		mp.Send(msg2, false)
		
		msg3 := new(MultiCastMessage)
		msg3.NewMCastMsgwithData(ip, STRING, "Sending MCAST")
		hostlist := make([]string, 0)
		hostlist = append(hostlist, "128.237.227.84")
		hostlist = append(hostlist, "128.2.13.133")
		hostlist = append(hostlist, "128.2.13.134")
		msg3.Origin = mp.ServerIP
		msg3.HostList = hostlist
		mp.SendMCast(msg3) 
	}
}
