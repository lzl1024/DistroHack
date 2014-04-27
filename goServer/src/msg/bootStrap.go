package msg

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"
	"util"
	"yaml"
)

var IsSN = false

var ListenPortLocal = 4213
var ListenPortPeer = 4214
var ListenPortSuperNode = 4214
var ListenPortHttpDB = 1234
var configSNList = make([]net.TCPAddr, 0)
var Question_URI string
var SNdone = make(chan bool)
var SNbootstrap = make(chan error)

var httpServerReady = false
var JoinAck bool

const Question_file = "https://s3.amazonaws.com/dsconfig/questions.txt"

func DoBootStrap() {
	retries := 0
	var err error
	if IsSN {
		// open the http server to provide database file
		go ConstructHttpServer()
		for retries != 3 {
			fmt.Println("do bootstrap")
			go BootStrapSN()
			err = <-SNbootstrap
			if err.Error() != "" {
				retries++
				fmt.Println("Trying BootStrap Again", retries)
				continue
			}
			break
		}
		if retries == 3 && err.Error() != "" {
			fmt.Println("Max tries on bootstrap done...Failing")
			os.Exit(-1)
		}
	} else {
		BootStrapON()
	}
}

func ReadConfig() error {
	for key, _ := range util.SNConfigNames {
		conn, err := net.DialTimeout("tcp", fmt.Sprint(key, ":", ListenPortSuperNode),
			(time.Duration(5000) * time.Millisecond))
		if err == nil {
			conn.Close()
			tcpAddr, err := net.ResolveTCPAddr("tcp", fmt.Sprint(key, ":", ListenPortSuperNode))
			if err == nil {
				configSNList = append(configSNList, *tcpAddr)
				fmt.Println("Connect to: ", tcpAddr.String())
			} else {
				fmt.Println("ReadConfig dns name fail ", key)
			}
		} else {
			fmt.Println("ReadConfig dial fail ", key)
		}
	}

	fmt.Printf("ReadConfig configSNList len %d\n", len(configSNList))
	return nil
}

func ReadQuestions() error {
	data, err := readWebFile(Question_file)
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

func ConstructHttpServer() {
	fmt.Println("bootstrap: constructHttpServer")
	http.HandleFunc("/hello", func(w http.ResponseWriter, req *http.Request) {
		io.WriteString(w, "hello, world!\n")
	})

	http.HandleFunc("/database", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "/tmp/users.csv")
	})

	httpServerReady = true
	err := http.ListenAndServe(fmt.Sprint(":", ListenPortHttpDB), nil)
	if err != nil {
		fmt.Println("BootStrap createServer ListenAndServe: ", err)
		os.Exit(-1)
	}
}

func loadFileFromHttpServer(ip string) bool {
	out, err := os.Create("/tmp/output.csv")
	if err != nil {
		fmt.Println("bootStrap loadfileFromhttpServer: ", err)
		return false
	}
	defer out.Close()

	resp, err := http.Get(fmt.Sprint("http://", ip, ":", ListenPortHttpDB, "/database"))
	if err != nil {
		fmt.Println("bootStrap loadfileFromhttpServer: ", err)
		return false
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		fmt.Println("testServer: ", err)
		return false
	}

	return true
}

func BootStrapSN() {
	// set isSN to be true in case of ON upgrade
	IsSN = true

	fmt.Println("BootstrapSN ")

	// read the url question from configuration file
	err := ReadQuestions()

	if err != nil {
		fmt.Println("Bootstrap BootStrapSN ReadQuestions Fail")

		err = errors.New(fmt.Sprint("BootStrap BootStrapSN ReadQuestions: ", err))
		SNbootstrap <- err
		return
	}

	// send msg to one SN in the list
	bootStrapMsg := new(Message)
	err = bootStrapMsg.NewMsgwithData("", SN_SN_JOIN, MsgPasser.ServerIP)
	if err != nil {
		fmt.Println("In BootStrapSN: ", err)
		SNbootstrap <- err
		return
	}

	// I am myself's supernode
	SuperNodeIP = MsgPasser.ServerIP

	// select the first entry randomly
	rand.Seed(time.Now().UnixNano())

	listLength := len(configSNList)
	fmt.Println("My configSNList length ", listLength)
	if listLength == 0 {
		SNbootstrap <- errors.New("")
		return
	}

	start := rand.Intn(listLength)
	for i := range configSNList {
		chose := (start + i) % listLength
		// not connect with myself
		if configSNList[chose].String() != MsgPasser.ServerIP {
			bootStrapMsg.Dest, _, _ = net.SplitHostPort(configSNList[chose].String())
			err = MsgPasser.Send(bootStrapMsg)
			if err != nil {
				continue
			} else {
				waitForJoinAck()
				if JoinAck == true {
					fmt.Println("bootstrap send bootstrapMsg to ", configSNList[chose].String())
					break
				} else {
					// delete the fail one from configSNList
					fmt.Println("Cannot can response from bootstrapping SN: ", configSNList[chose].String())
					configSNList = append(configSNList[:chose], configSNList[chose+1:]...)
					listLength = listLength - 1
					if listLength == 0 {
						fmt.Println("Fail to connect with any bootstapping node, try again later...")
						os.Exit(-1)
					}
					continue
				}
			}
		}
	}

	if err != nil {
		SNbootstrap <- err
		return
	}
}

func waitForJoinAck() {
	JoinAck = false
	// time out wait for join ack
	for i := 0; i <= BusyWaitingTimeOutRound; i++ {
		time.Sleep(BusyWaitingSleepInterval)

		// check ack status
		if JoinAck == true {
			break
		}
	}
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
	if listLength == 0 {
		return err
	}

	start := rand.Intn(listLength)
	for i := range configSNList {
		chose := (start + i) % listLength
		bootStrapMsg.Dest, _, _ = net.SplitHostPort(configSNList[chose].String())
		err = MsgPasser.Send(bootStrapMsg)
		if err != nil {
			continue
		} else {
			waitForJoinAck()
			if JoinAck == true {
				break
			} else {
				// delete the fail one from configSNList
				fmt.Println("Cannot can response from bootstrapping SN: ", configSNList[chose].String())
				configSNList = append(configSNList[:chose], configSNList[chose+1:]...)
				listLength = listLength - 1
				if listLength == 0 {
						fmt.Println("Fail to connect with any bootstapping node, try again later...")
						os.Exit(-1)
				}
				continue
			}
		}
	}

	return err
}

func RcvOnJoin(msg *Message) (interface{}, error) {
	if msg.Kind != ON_SN_JOIN {
		return nil, errors.New("message Kind indicates not a ON_SN_JOIN")
	}

	SNHostlistMutex.Lock()
	var snIP string
	if IsSN == false {
		snIP = "0"
	} else {
		// get the SN with the lightest load
		min := (1 << 30)

		for k, _ := range MsgPasser.SNLoadlist {
			if MsgPasser.SNLoadlist[k] < min {
				min = MsgPasser.SNLoadlist[k]
				snIP = k
			}
		}
	}

	/* Send ON the SN IP it should connect to */
	m := new(Message)
	m.NewMsgwithData(msg.Origin, SN_ON_JOIN_ACK, snIP)
	err := MsgPasser.Send(m)
	SNHostlistMutex.Unlock()
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

	if ip == "0" {
		JoinAck = false
		return ip, err
	}

	JoinAck = true
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
	ONHostlistMutex.Lock()
	/* Update ONlist */
	MsgPasser.ONHostlist[ip] = ip

	/* Send MCast to other ONs in the group */
	changeONList := new(Message)
	err = changeONList.NewMsgwithData("", SN_ON_CHANGEONLIST, MsgPasser.ONHostlist)
	ONHostlistMutex.Unlock()
	if err != nil {
		fmt.Println("In RcvSnOnRegister: ")
		return nil, err
	}

	// send message to ONs
	MulticastMsgInGroup(changeONList, false)

	/* Send Load Message to Others */
	newMsg := new(Message)
	ONHostlistMutex.Lock()
	newMsg.NewMsgwithData("", SN_SN_LOADUPDATE, len(MsgPasser.ONHostlist))
	ONHostlistMutex.Unlock()
	if err != nil {
		fmt.Println("In RcvSnOnRegister: ")
		return nil, err
	}
	MulticastMsgInGroup(newMsg, true)

	return msg, nil
}

// all ON get this msg should change their point to new ONlist directly
func RcvSNChangeONList(msg *Message) (interface{}, error) {
	if msg.Kind != SN_ON_CHANGEONLIST {
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
		ONHostlistMutex.Lock()
		MsgPasser.ONHostlist = newONList
		ONHostlistMutex.Unlock()

		return newONList, nil
	} else {
		return "Haha! I am SN!", nil
	}
}

func RcvSnLoadUpdate(msg *Message) (interface{}, error) {
	if msg.Kind != SN_SN_LOADUPDATE {
		return nil, errors.New("message Kind indicates not a SN_SN_LOADUPDATE")
	}

	var load int
	err := ParseRcvInterfaces(msg, &load)
	if err != nil {
		fmt.Println("In RcvSnLoadUpdate: ")
		return nil, err
	}

	SNHostlistMutex.Lock()
	MsgPasser.SNLoadlist[msg.Origin] = load
	SNHostlistMutex.Unlock()
	newMsg := new(Message)
	ONHostlistMutex.Lock()
	newMsg.NewMsgwithData("", SN_SN_LOADMERGE, len(MsgPasser.ONHostlist))
	ONHostlistMutex.Unlock()
	if err != nil {
		fmt.Println("In RcvSnOnRegister: ")
		return nil, err
	}
	MulticastMsgInGroup(newMsg, true)

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

	SNHostlistMutex.Lock()
	MsgPasser.SNLoadlist[msg.Origin] = load
	SNHostlistMutex.Unlock()
	fmt.Println("Current Load Info:")
	for k, _ := range MsgPasser.SNLoadlist {
		fmt.Println(k, MsgPasser.SNLoadlist[k])
	}

	return msg, nil
}

// one bootstraping SN get this message and send out update_list msgs
func RcvSnJoin(msg *Message) (interface{}, error) {
	if msg.Kind != SN_SN_JOIN {
		return nil, errors.New("message Kind indicates not a SN_SN_JOIN")
	}

	var ip string
	err := ParseRcvInterfaces(msg, &ip)
	if err != nil {
		fmt.Println("In RcvSnJoin: ", err)
		return nil, err
	}

	if strings.EqualFold(ip, msg.Src) == false {
		return nil, errors.New("message Src Doesn't match IP address sent")
	}

	// make sure http server created
	for !httpServerReady {
		fmt.Println("BootStrap: wait for http server...")
	}

	// export data in db into file on server
	err = util.DatabaseCreateDBFile()
	if err != nil {
		fmt.Println("In RcvSnJoin: ", err)
		return nil, err
	}

	// send out ack msg tell the new sn data is ready on server
	bootStrapMsg := new(Message)

	StartEnd_Lock.Lock()
	returnIP := MsgPasser.ServerIP

	// if I am not SN set error mark
	if IsSN == false {
		returnIP = "0"
	}

	backData := map[string]string{
		"serverIP":  returnIP,
		"startTime": StartTime,
	}
	StartEnd_Lock.Unlock()

	err = bootStrapMsg.NewMsgwithData(ip, SN_SN_JOIN_ACK, backData)
	err = MsgPasser.Send(bootStrapMsg)
	if err != nil {
		fmt.Println("In RcvSnJoin: ")
		return nil, err
	}

	/* a new super node has tried to join , add him to our list and multicast that
	 * a new node has joined, and everyone should update their lists
	 */
	hostlist := make(map[string]string)
	hostlist[ip] = ip
	SNHostlistMutex.Lock()
	for k, _ := range MsgPasser.SNHostlist {
		hostlist[k] = MsgPasser.SNHostlist[k]
	}
	SNHostlistMutex.Unlock()
	newMCastMsg := new(MultiCastMessage)
	newMCastMsg.NewMCastMsgwithData("", SN_SN_LISTUPDATE, hostlist)
	newMCastMsg.HostList = hostlist
	newMCastMsg.Origin = MsgPasser.ServerIP
	newMCastMsg.Seqnum = atomic.AddInt32(&MsgPasser.SeqNum, 1)
	MsgPasser.SendMCast(newMCastMsg)

	return ip, nil
}

func RcvSnJoinAck(msg *Message) (interface{}, error) {
	var err error
	if msg.Kind != SN_SN_JOIN_ACK {
		err = errors.New("message Kind indicates not a SN_SN_JOIN_ACK")
		SNbootstrap <- err
		return nil, err
	}

	var ipWithStartTime map[string]string
	err = ParseRcvInterfaces(msg, &ipWithStartTime)

	if err != nil {
		fmt.Println("In RcvSnJoinAck: ")
		SNbootstrap <- err
		return nil, err
	}

	ip := ipWithStartTime["serverIP"]

	if ip == "0" {
		JoinAck = false
		return nil, err
	}

	JoinAck = true
	// update hack start time
	StartEnd_Lock.Lock()
	StartTime = ipWithStartTime["startTime"]
	StartEnd_Lock.Unlock()

	if strings.EqualFold(ip, msg.Src) == false {
		err = errors.New("message Src Doesn't match IP address sent")
		SNbootstrap <- err
		return nil, err
	}

	// get the file from http Server
	if !loadFileFromHttpServer(ip) {
		err = errors.New("bootStrap: load db file failed")
		SNbootstrap <- err
		return nil, err
	}

	// import the file into database
	err = util.DatabaseLoadDBFile()
	if err != nil {
		fmt.Println("In RcvSnJoinAck: Importing file to DB failed ")
		SNbootstrap <- err
		return nil, err
	}

	/* Send Load Message to Others */
	newMsg := new(Message)
	ONHostlistMutex.Lock()
	newMsg.NewMsgwithData("", SN_SN_LOADUPDATE, len(MsgPasser.ONHostlist))
	ONHostlistMutex.Unlock()
	if err != nil {
		fmt.Println("In RcvSnOnRegister: ")
		return nil, err
	}
	MulticastMsgInGroup(newMsg, true)

	SNbootstrap <- errors.New("")

	return nil, nil
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
	SNHostlistMutex.Lock()
	for k, _ := range hostlist {
		MsgPasser.SNHostlist[k] = hostlist[k]
	}

	newMsg := new(Message)
	newMsg.NewMsgwithData("", SN_SN_LISTMERGE, MsgPasser.SNHostlist)
	SNHostlistMutex.Unlock()
	if err != nil {
		fmt.Println("In RcvSnOnRegister: ")
		return nil, err
	}
	MulticastMsgInGroup(newMsg, true)

	return MsgPasser.SNHostlist, nil
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

	SNHostlistMutex.Lock()
	/* merge the hostlist with current SNlist */
	for k, _ := range hostlist {
		MsgPasser.SNHostlist[k] = hostlist[k]
	}
	SNHostlistMutex.Unlock()
	return MsgPasser.SNHostlist, nil
}
