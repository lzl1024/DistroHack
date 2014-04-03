package superNode

import (
	"encoding/gob"
	"fmt"
	"msg"
	"net"
	"util"
)

// TODO superNode Thread,
// function 1: when receiving supernode "global_ranking" change, merge lists, update and send to all its peers
// function 2: when receiving on "submit success", update global ranking list and send to its peers and other SN if needed
// function 3: when recieve "sign_in" from ON check available SN and send msg to it, to let him register the on or send back sign in fail.
// function 4: connect with an ON and send all msg. userlist, local_info, global ranking.

var mp = msg.MsgPasser
var rankList [msg.GlobalRankSize]msg.UserRecord
var scoreMap map[string]msg.UserRecord

var Global_ranking = []msg.UserRecord{}
var Local_info = map[string]msg.UserRecord{}

func SuperNodeThread(serverPort int) {
	// Get the list of other super node
	//parseConfigFile()

	fmt.Println("aaa")

	scoreMap = make(map[string]msg.UserRecord)

	// Get the tcpAddr
	/*fmt.Println("SuperNode: Started superNode server thread")
	service := fmt.Sprint(":", serverPort)
	tcpAddr, err := net.ResolveTCPAddr("tcp", service)
	if err != nil {
		fmt.Println("SuperNode: Unrecoverable error trying to start go supernode server")
		c <- err
		return
	}*/

	// Create Listener
	/*listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		fmt.Println("SuperNode: Unrecoverable error trying to start listening on server ", err)
		c <- err
		return
	}*/

	// Create new receive thread when a new msg arrives
	/*for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("SuperNode: error accepting connection...continuing")
			continue
		}
		go rcvThreadForSN(mp, conn)
	}*/

}

func rcvThreadForSN(mp *msg.Messagepasser, conn net.Conn) {
	fmt.Println("SuperNode: Started super node recevier thread\n")
	var tcpconn *net.TCPConn
	var ok bool
	var err error
	var msg msg.Message

	tcpconn, ok = conn.(*net.TCPConn)
	if ok {
		err = tcpconn.SetLinger(0)
		if err != nil {
			fmt.Println("SuperNode: cannot set linger options")
		}
	}

	decoder := gob.NewDecoder(conn)
	for {
		err := decoder.Decode(&msg)
		if err != nil {
			fmt.Println("SuperNode: error while decoding: ", err)
			conn.Close()
			break
		}
		superNodeMsgDoAction(&msg)
	}
}

func superNodeMsgDoAction(m *msg.Message) {
	data, err := msg.Handlers[m.Kind](m)
	if err != nil {
		return
	}
	fmt.Println((*m).String())
	//fmt.Println(str)

	//updateGlobalRankList(str)

	switch m.Kind {
	case msg.SN_RANK:
		updateGlobalRankList(data.([msg.GlobalRankSize]msg.UserRecord))
	case msg.SN_PBLSUCCESS:
		pblSuccessFromON(data.(msg.UserRecord))
	case msg.SN_SIGNIN:
		signInFromON(m, data.(map[string]string))
	}
}

/*func globalRankForSN(m *msg.Message) {
	tmpRankList := m.Data
	if len(tmpRankList) != RankNum {
		fmt.Println("SuperNode: Message data of rank list length incorrect")
	}

	updateGlobalRankList(tmpRankList)
}*/

func updateGlobalRankList(tmpRankList [msg.GlobalRankSize]msg.UserRecord) {
	rankList = tmpRankList

	for _, userRecord := range tmpRankList {
		r := &userRecord
		scoreMap[r.UserName] = userRecord
	}
}

func pblSuccessFromON(userRecord msg.UserRecord) {
	r := &userRecord

	scoreMap[r.UserName] = userRecord

	i := msg.GlobalRankSize - 1
	if i >= 0 && r.Score > (&rankList[i]).Score {
		tmpRank := rankList[i]
		rankList[i] = userRecord
		if i < msg.GlobalRankSize-1 {
			rankList[i+1] = tmpRank
		}
		i--
	}
}

func signInFromON(m *msg.Message, signInMsg map[string]string) {
	// check database and send back SignInAck
	// TODO: update number of node register, send by heartbeat

	backMsg := util.DatabaseSignIn(signInMsg["username"], signInMsg["password"])

	backData := map[string]string{
		"user":   signInMsg["username"],
		"status": backMsg,
	}

	sendoutMsg := new(msg.Message)
	err := sendoutMsg.NewMsgwithData(m.Src, msg.SIGNINACK, backData)
	if err != nil {
		fmt.Println(err)
	}

	// send message to SN
	mp.Send(sendoutMsg, false)
}

func InitiateWithON() {

}
