package msg

import (
	"net"
	"bytes"
	"fmt"
)

func clienthread(mp *Messagepasser, c chan error) {
	var ip string
	var msg Message
	for {
		fmt.Println("Enter who you would like to connect to")
		fmt.Scanf("%s", &ip)
		if net.ParseIP(ip) == nil {
			fmt.Println("Invalid ip: try again")
			continue
		}
		
		msg.Dest = ip
		msg.Data = bytes.NewBufferString("ashish kaila").Bytes()
		mp.Send(&msg)
	}
}

