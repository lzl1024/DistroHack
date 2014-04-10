package msg

import (
	"fmt"
	"net"
	"time"
	"sync/atomic"
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
	testMulticast(mp, c)
}

func testConstructSNList(mp *Messagepasser, c chan error) {
	mp.SNHostlist.PushBack("128.2.13.134")
	mp.SNHostlist.PushBack("128.2.13.133")
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
		err = msg2.NewMsgwithData(ip, SN_PBLSUCCESS, userData)
		if err != nil {
			fmt.Println("Reached here:", err)
			continue
		}
		mp.Send(msg2)

		msg3 := new(MultiCastMessage)
		msg3.NewMCastMsgwithData(ip, STRING, "Sending MCAST")
		hostlist := make([]string, 0)
		hostlist = append(hostlist, "128.237.124.82")
		hostlist = append(hostlist, "128.2.13.133")
		hostlist = append(hostlist, "128.2.13.134")
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

	err := msg1.NewMsgwithData(ip, SN_ONSIGNIN, data)
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

	err := msg1.NewMsgwithData(ip, SN_PBLSUCCESS, *userRecord)
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
	msg3.NewMCastMsgwithData(ip, SN_RANK, testRankList)
	hostlist := make([]string, 0)
	hostlist = append(hostlist, "128.237.220.160")
	hostlist = append(hostlist, "128.2.13.133")
	hostlist = append(hostlist, "128.2.13.134")
	msg3.Origin = mp.ServerIP
	msg3.HostList = hostlist
	mp.SendMCast(msg3)

}
