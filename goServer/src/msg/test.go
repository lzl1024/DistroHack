package msg

import (
	"fmt"
	"net"
	"sync/atomic"
	"strconv"
	"time"
)

func TestMessagePasser() {
	channel := make(chan error)

	go clientTestThread(MsgPasser, channel)
	value := <-channel
	fmt.Println(value)
}

func clientTestThread(mp *Messagepasser, c chan error) {
	//testPublishSuccess(mp, c)
	//testGlobalRank(mp, c)

	//testMulticast(mp, c)
	//mp.SNHostlist["10.0.1.17"] = "10.0.1.17"
	//mp.ONHostlist["10.0.1.17"] = "10.0.1.17"
	testDLock(mp, c)
}

func testConstructSNList(mp *Messagepasser, c chan error) {
	mp.SNHostlist["128.2.13.134"] = "128.2.13.134"
	mp.SNHostlist["128.2.13.133"] = "128.2.13.133"
}

func testDLock(mp *Messagepasser, c chan error) {
	for {
		var option string
		var opt int
		fmt.Println("Enter what you want to do:")
		fmt.Scanf("%s", &option)
		opt,_ = strconv.Atoi(option)
		if opt == 1 {
			go optionLock()
		} else {
			continue
		}
	}
}

func optionLock() {
	fmt.Println("Trying to get DLock()")
	DLock.Lock()
	time.Sleep(time.Duration(6) * time.Second)
	fmt.Println("Unlocking the DLock()")
	DLock.Unlock()
}

func testMulticast(mp *Messagepasser, c chan error) {
	for {
		var ip string
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
		mp.Send(msg1)

		time.Sleep(20)

		msg2 := new(Message)
		userData := UserRecord{
			"akaila",
			100,
			time.Now(),
		}
		err = msg2.NewMsgwithData(ip, ON_SN_PBLSUCCESS, userData)
		if err != nil {
			fmt.Println("Reached here:", err)
			continue
		}
		mp.Send(msg2)


		msg3 := new(MultiCastMessage)
		msg3.NewMCastMsgwithData(ip, STRING, "Sending MCAST")
		hostlist := make(map[string]string)
		hostlist["10.0.0.2"] = "10.0.0.2"
		hostlist["128.2.13.133"]= "128.2.13.133"
		hostlist["128.2.13.134"] = "128.2.13.134"
		msg3.Origin = mp.ServerIP
		msg3.HostList = hostlist
		msg3.Seqnum = atomic.AddInt32(&mp.SeqNum, 1)
		mp.SendMCast(msg3)
	}
}

func testSuperNodeONSIGNIN(mp *Messagepasser, c chan error) {
	var ip string
	fmt.Println("Enter which super node you would like to connect to")
	fmt.Scanf("%s", &ip)
	if net.ParseIP(ip) == nil {
		fmt.Println("Invalid ip: try again")
		return
	}

	msg1 := new(Message)
	data := make(map[string]string)
	data["username"] = "kb24"
	data["password"] = "nddndd"

	err := msg1.NewMsgwithData(ip, ON_SN_SIGNIN, data)
	if err != nil {
		fmt.Println(err)
		return
	}
	mp.Send(msg1)
}

func testPublishSuccess(mp *Messagepasser, c chan error) {
	var ip string

	fmt.Println("Enter which super node you would like to connect to")
	fmt.Scanf("%s", &ip)
	if net.ParseIP(ip) == nil {
		fmt.Println("Invalid ip: try again")
		return
	}

	msg1 := new(Message)
	userRecord := new(UserRecord)
	userRecord.NewUserRecord("kb24", 1, time.Now().Add(mp.drift))

	err := msg1.NewMsgwithData(ip, ON_SN_PBLSUCCESS, *userRecord)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Test: Before publish success")
	mp.Send(msg1)
}

func testGlobalRank(mp *Messagepasser, c chan error) {
	var ip string

	fmt.Println("Enter which super node you would like to connect to")
	fmt.Scanf("%s", &ip)
	if net.ParseIP(ip) == nil {
		fmt.Println("Invalid ip: try again")
		return
	}

	var testRankList [GlobalRankSize]UserRecord
	userRecord := new(UserRecord)
	userRecord.NewUserRecord("kb24", 1, time.Now().Add(mp.drift))
	testRankList[0] = *userRecord

	msg3 := new(MultiCastMessage)
	msg3.NewMCastMsgwithData(ip, SN_SN_RANK, testRankList)
	hostlist := make(map[string]string)
	hostlist["128.237.218.95"] = "128.237.218.95"
	hostlist["128.2.13.133"] = "128.2.13.133"
	hostlist["128.2.13.134"] = "128.2.13.134"
	msg3.Origin = mp.ServerIP
	msg3.HostList = hostlist
	msg3.Seqnum = atomic.AddInt32(&mp.SeqNum, 1)
	mp.SendMCast(msg3)

}
