package msg

import (
	"fmt"
	"net"
	"yaml"
	"errors"
	"io/ioutil"
	"strings"
	"sync/atomic"
)

var ListenPortLocal = 4213
var ListenPortPeer  = 4214
var ListenPortSuperNode = 4214
var configSNList = make([]net.TCPAddr, 0)
var uri string
var SNdone = make(chan bool)

func ReadConfig() error {
	filename := "config.txt"
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println(err)
		return err
	} 
    m := make(map[interface{}]interface{})
    err = yaml.Unmarshal([]byte(data), &m)
 	if err != nil {
 		fmt.Println(err)
 		return err
 	}
 	for key,_ := range m {
 		tcpAddr,err := net.ResolveTCPAddr("tcp", fmt.Sprint(m[key].(string), ":", ListenPortSuperNode))
 		if err == nil {
 			configSNList = append(configSNList, *tcpAddr)
 			fmt.Println(tcpAddr.String())
 		}
 	}

 	filename = "questions.txt"
 	data, err = ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println(err)
		return err
	}
	m = make(map[interface{}]interface{})
    err = yaml.Unmarshal([]byte(data), &m)
 	if err != nil {
 		fmt.Println(err)
 		return err
 	}
 	uri = m["url"].(string)
 	return nil
}

func BootStrapSN() error{
	bootStrapMsg := new(Message)
	err := bootStrapMsg.NewMsgwithData("", SN_JOIN, MsgPasser.ServerIP)
	if err != nil {
		fmt.Println(err)
		return err
	}
	for i := range configSNList {
		bootStrapMsg.Dest,_,_ = net.SplitHostPort(configSNList[i].String())
		err := MsgPasser.Send(bootStrapMsg)
		if err != nil {
			continue
		}
		break
	}
	
	return nil
}

func RcvSnJoin(msg *Message) (interface{}, error) {
	if msg.Kind != SN_JOIN {
		return nil, errors.New("message Kind indicates not a SN_ONSIGNIN")
	}
	
	var ip string
	err := ParseRcvInterfaces(msg, &ip)
	if err != nil {
		return nil, err
	}
	
	if strings.EqualFold(ip, msg.Src) == false {
		return nil, errors.New("message Src Doesn't match IP address sent")
	}
	
	/* a new super node has tried to join , add him to our list and multicast that 
	 * a new node has joined, and everyone should update their lists
	 */
	hostlist := make([]string, 0)
	hostlist = append(hostlist, ip)
	for k,_ := range MsgPasser.SNHostlist {
		hostlist = append(hostlist, MsgPasser.SNHostlist[k])
	}
	
	newMCastMsg := new(MultiCastMessage)
	newMCastMsg.NewMCastMsgwithData("", SN_SNLISTUPDATE, hostlist)
	newMCastMsg.HostList = hostlist
	newMCastMsg.Origin = MsgPasser.ServerIP
	newMCastMsg.Seqnum = atomic.AddInt32(&MsgPasser.SeqNum, 1)
	MsgPasser.SendMCast(newMCastMsg)
	
	return ip, nil
}