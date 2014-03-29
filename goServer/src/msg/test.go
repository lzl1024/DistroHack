package msg

import (
	"fmt"
	"os"
	"net"
)

const SERVER_PORT = 5989

func TestMessagePasser(arg string) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Println("error getting interfaces: ", err)
		os.Exit(-1)
	}
	
	var addr net.IP
	for i := range addrs {
		ip,_,_ := net.ParseCIDR(addrs[i].String())
		if ip.IsLoopback() || ip.IsInterfaceLocalMulticast() || 
		ip.IsLinkLocalMulticast() || ip.IsLinkLocalUnicast() || 
		ip.IsMulticast() {
			continue
		} else {
			addr = ip;
			break;
		}
	}
	
	var mp *Messagepasser
	mp = NewMsgPasser(addr.String(), SERVER_PORT)
	channel := make(chan error)
	
	isclient := arg
	if isclient == "true" {
		go clienthread(mp, channel)
		value := <- channel
		fmt.Println(value)
	} else {
		go serverthread(mp, channel)
		value := <- channel
		fmt.Println(value)
	}
}

