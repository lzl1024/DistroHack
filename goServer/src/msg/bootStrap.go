package msg

import (
	"fmt"
	"net"
	"yaml"
	"errors"
	"io/ioutil"
	"strings"
	"sync/atomic"
	"net/http"
	"math/rand"
	"time"
)

var ListenPortLocal = 4213
var ListenPortPeer  = 4214
var ListenPortSuperNode = 4214
var configSNList = make([]net.TCPAddr, 0)
var Question_URI string
var SNdone = make(chan bool)

const Config_ALL = "https://s3.amazonaws.com/dsconfig/config_local.txt"
const Config_SN = "https://s3.amazonaws.com/dsconfig/config_cmu.txt"
const Question_file = "https://s3.amazonaws.com/dsconfig/questions.txt"


func ReadConfig() error {
	// local read file
	filename := "config.txt"
	data, err := ioutil.ReadFile(filename)
	//data, err :=readWebFile(Config_ALL)
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
 			fmt.Println("Connect to: " , tcpAddr.String())
 		}
 	}
 	
 	return nil
}

func ReadQuestions() error {
	// local read file
	/*filename := "questions.txt"
 	data, err := ioutil.ReadFile(filename)*/
 	data, err :=readWebFile(Question_file)
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
 	Question_URI = m["url"].(string)
 	return nil
}


// small helper function, read file from web
func readWebFile(url string) ([]byte, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(res.Body)
    if err != nil {
        return nil, err
    }
    defer res.Body.Close()
    
    return data, nil
}

func BootStrapSN() error {
	// read the url question from configuration file
	ReadQuestions()
	
	// send msg to one SN in the list
	bootStrapMsg := new(Message)
	err := bootStrapMsg.NewMsgwithData("", SN_SN_JOIN, MsgPasser.ServerIP)
	if err != nil {
		fmt.Println("In BootStrapSN: ", err)
		return err
	}
	
	// select the first entry randomly
	rand.Seed(time.Now().UnixNano())
	listLength := len(configSNList)
	start := rand.Intn(listLength)
	for i := range configSNList {
		chose := (start + i) % listLength
		bootStrapMsg.Dest,_,_ = net.SplitHostPort(configSNList[chose].String())
		err = MsgPasser.Send(bootStrapMsg)
		if err != nil {
			continue
		}
		break
	}

	return err
}

func BootStrapON() error {
	// send ON join msg to all bootstrap SNs
	bootStrapMsg := new(Message)
	err := bootStrapMsg.NewMsgwithData("", ON_SN_JOIN, MsgPasser.ServerIP)
	if err != nil {
		fmt.Println("In BootStrapON: ", err)
		return err
	}
	
	// select the first entry randomly
	rand.Seed(time.Now().UnixNano())
	listLength := len(configSNList)
	start := rand.Intn(listLength)
	for i := range configSNList {
		chose := (start + i) % listLength
		bootStrapMsg.Dest,_,_ = net.SplitHostPort(configSNList[chose].String())
		err = MsgPasser.Send(bootStrapMsg)
		if err != nil {
			continue
		}
		break
	}	
		
	return err
}

func RcvOnJoin(msg *Message) (interface{}, error) {
	if msg.Kind != ON_SN_JOIN {
		return nil, errors.New("message Kind indicates not a ON_SN_JOIN")
	}
	
	// get the SN with the lightest load
	min := (1<<31)
	var snIP string
	for k,_ := range MsgPasser.SNLoadlist {
		if MsgPasser.SNLoadlist[k] < min {
			min = MsgPasser.SNLoadlist[k]
			snIP = k
		}
	}
	
	/* Send ON the SN IP it should connect to */
	m := new(Message)
	m.NewMsgwithData(msg.Origin, SN_ON_JOIN_ACK, snIP)
	err := MsgPasser.Send(m)
	
	return msg, err
}

func RcvOnJoinAck(msg *Message) (interface{}, error) {
	if msg.Kind != SN_ON_JOIN_ACK {
		return nil, errors.New("message Kind indicates not a SN_ON_JOIN_ACK")
	}
	
	var ip string
	err := ParseRcvInterfaces(msg, &ip)
	if err != nil {
		fmt.Println("In RcvOnJoinAck: ")
		return nil, err
	}
	
	SuperNodeIP = ip
	bootStrapMsg := new(Message)
	err = bootStrapMsg.NewMsgwithData(ip, ON_SN_REGISTER, MsgPasser.ServerIP)
	err = MsgPasser.Send(bootStrapMsg)

	return ip, err
}

func RcvSnOnRegister(msg *Message) (interface{}, error) {
	if msg.Kind != ON_SN_REGISTER {
		return nil, errors.New("message Kind indicates not a ON_SN_REGISTER")
	}

	var ip string
	err := ParseRcvInterfaces(msg, &ip)
	if err != nil {
		return nil, err
	}
	/* Update ONlist */
	MsgPasser.ONHostlist[ip] = ip
	
	/* Send MCast to other ONs in the group */
	changeONList := new(Message)
	err = changeONList.NewMsgwithData("", SN_ON_CHANGEONLIST, MsgPasser.ONHostlist)
	if err != nil {
		fmt.Println("In RcvSnOnRegister: ")
		return nil, err
	}
	
	// send message to ONs
	MulticastMsgInGroup(changeONList, false)
	
	
	/* Send Load Message to Others */
	newMCastMsg := new(MultiCastMessage)
	newMCastMsg.NewMCastMsgwithData("", SN_SN_LOADUPDATE, len(MsgPasser.ONHostlist))
	newMCastMsg.HostList = MsgPasser.SNHostlist
	newMCastMsg.Origin = MsgPasser.ServerIP
	newMCastMsg.Seqnum = atomic.AddInt32(&MsgPasser.SeqNum, 1)
	MsgPasser.SendMCast(newMCastMsg)
	
	return msg, nil
}


// all ON get this msg should change their point to new ONlist directly
func RcvSNChangeONList(msg *Message) (interface{}, error) {
	if msg.Kind != SN_ON_CHANGEONLIST{
		return nil, errors.New("message Kind indicates not a SN_ON_CHANGEONLIST")
	}
	
	// in case of concurrent issue, only its ON should change
	if msg.Origin != MsgPasser.ServerIP {
		var newONList map[string]string
		err := ParseRcvInterfaces(msg, &newONList)
		if err != nil {
			fmt.Println("In RcvSNCHangeONList: ")
			return nil, err
		}	
		MsgPasser.ONHostlist = newONList
	
		return newONList, nil
	} else {
		return "Haha! I am SN!", nil
	} 
}


func RcvSnLoadUpdate(msg *Message) (interface{}, error) {
	if msg.Kind != SN_SN_LOADUPDATE{
		return nil, errors.New("message Kind indicates not a SN_SN_LOADUPDATE")
	}

	var load int
	err := ParseRcvInterfaces(msg, &load)
	if err != nil {
		fmt.Println("In RcvSnLoadUpdate: ")
		return nil, err
	}
	
	MsgPasser.SNLoadlist[msg.Origin] = load
	
	newMCastMsg := new(MultiCastMessage)
	newMCastMsg.NewMCastMsgwithData("", SN_SN_LOADMERGE, len(MsgPasser.ONHostlist))
	newMCastMsg.HostList = MsgPasser.SNHostlist
	newMCastMsg.Origin = MsgPasser.ServerIP
	newMCastMsg.Seqnum = atomic.AddInt32(&MsgPasser.SeqNum, 1)
	MsgPasser.SendMCast(newMCastMsg)
	
	return msg, nil
}


func RcvSnLoadMerge(msg *Message) (interface{}, error) {
	if msg.Kind != SN_SN_LOADMERGE {
		return nil, errors.New("message Kind indicates not a SN_SN_LOADMERGE")
	}
	
	var load int
	err := ParseRcvInterfaces(msg, &load)
	if err != nil {
		fmt.Println("In RcvSnLoadMerge: ")
		return nil, err
	}
	
	MsgPasser.SNLoadlist[msg.Origin] = load
	fmt.Println("Current Load Info:")
	for k,_ := range MsgPasser.SNLoadlist {
		fmt.Println(k, MsgPasser.SNLoadlist[k])
	}
	
	return msg, nil
}


// one bootstraping SN get this meesage and send out update_list msgs
func RcvSnJoin(msg *Message) (interface{}, error) {
	if msg.Kind != SN_SN_JOIN {
		return nil, errors.New("message Kind indicates not a SN_SN_JOIN")
	}
	
	var ip string
	err := ParseRcvInterfaces(msg, &ip)
	if err != nil {
		fmt.Println("In RcvSnJoin: ")
		return nil, err
	}
	
	if strings.EqualFold(ip, msg.Src) == false {
		return nil, errors.New("message Src Doesn't match IP address sent")
	}

	
	// TODO: export and send back db data, when response, add it to loadlist with its actual ONlist length
	// (It may not be 1 when bootstrapSN is called in SN failure) and send out Loadlist update
	
	/* a new super node has tried to join , add him to our list and multicast that 
	 * a new node has joined, and everyone should update their lists
	 */
	hostlist := make(map[string]string)
	hostlist[ip] = ip
	for k,_ := range MsgPasser.SNHostlist {
		hostlist[k] = MsgPasser.SNHostlist[k]
	}
	
	newMCastMsg := new(MultiCastMessage)
	newMCastMsg.NewMCastMsgwithData("", SN_SN_LISTUPDATE, hostlist)
	newMCastMsg.HostList = hostlist
	newMCastMsg.Origin = MsgPasser.ServerIP
	newMCastMsg.Seqnum = atomic.AddInt32(&MsgPasser.SeqNum, 1)
	MsgPasser.SendMCast(newMCastMsg)
	
	return ip, nil
}


// every SN get this meesage, update their SNlist and send out list merge
func RcvSnListUpdate(msg *Message) (interface{}, error) {
	if msg.Kind != SN_SN_LISTUPDATE {
		return nil, errors.New("message Kind indicates not a SN_SN_LISTUPDATE")
	}
	
	var hostlist map[string]string
	err := ParseRcvInterfaces(msg, &hostlist)
	if err != nil {
		fmt.Println("In RcvSnListUpdate: ")
		return nil, err
	}
	
	/* merge the hostlist with current SNlist */
	for k,_ := range hostlist {
		MsgPasser.SNHostlist[k] = hostlist[k]
	}
	
	newMCastMsg := new(MultiCastMessage)
	newMCastMsg.NewMCastMsgwithData("", SN_SN_LISTMERGE, MsgPasser.SNHostlist)
	newMCastMsg.HostList = MsgPasser.SNHostlist
	newMCastMsg.Origin = MsgPasser.ServerIP
	newMCastMsg.Seqnum = atomic.AddInt32(&MsgPasser.SeqNum, 1)
	MsgPasser.SendMCast(newMCastMsg)

	return hostlist, nil
}

func RcvSnListMerge(msg *Message) (interface{}, error) {
	if msg.Kind != SN_SN_LISTMERGE {
		return nil, errors.New("message Kind indicates not a SN_SN_LISTMERGE")
	}
	
	var hostlist map[string]string
	err := ParseRcvInterfaces(msg, &hostlist)
	if err != nil {
		fmt.Println("In RcvSnListMerge: ")
		return nil, err
	}
	
	/* merge the hostlist with current SNlist */
	for k,_ := range hostlist {
		MsgPasser.SNHostlist[k] = hostlist[k]
	}
	
	return hostlist, nil
}