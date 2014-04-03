package superNode

import (
	"encoding/gob"
	"fmt"
	"msg"
	"net"
)

// TODO superNode Thread,
// function 1: when receiving supernode "global_ranking" change, merge lists, update and send to all its peers
// function 2: when receiving on "submit success", update global ranking list and send to its peers and other SN if needed
// function 3: when recieve "sign_in" from ON check available SN and send msg to it, to let him register the on or send back sign in fail.
// function 4: connect with an ON and send all msg. userlist, local_info, global ranking.

var mp = msg.MsgPasser
var rankList [msg.RankNum]msg.Rank
var scoreMap map[string]msg.Rank
var RankNum int

func SuperNodeThread(serverPort int) {
	// Get the list of other super node
	//parseConfigFile()

	RankNum = msg.RankNum
	scoreMap = make(map[string]msg.Rank)

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
	data, err := msg.Handlers[m.Kind].Decode(m)
	if err != nil {
		return
	}
	fmt.Println((*m).String())
	//fmt.Println(str)

	//updateGlobalRankList(str)

	switch m.Kind {
	case msg.SN_RANK:
		updateGlobalRankList(data.([msg.RankNum]msg.Rank))
	case msg.SN_ON_SUBMIT:
		submitSuccessFromON(data.(msg.Rank))
	case msg.SN_ON_SIGNIN:
		signInFromON(m)
	}
}

/*func globalRankForSN(m *msg.Message) {
	tmpRankList := m.Data
	if len(tmpRankList) != RankNum {
		fmt.Println("SuperNode: Message data of rank list length incorrect")
	}

	updateGlobalRankList(tmpRankList)
}*/

func updateGlobalRankList(tmpRankList [msg.RankNum]msg.Rank) {
	rankList = tmpRankList

	for _, singleRank := range tmpRankList {
		r := &singleRank
		scoreMap[r.Username] = singleRank
	}
}

func submitSuccessFromON(singleRank msg.Rank) {
	r := &singleRank

	scoreMap[r.Username] = singleRank

	i := RankNum - 1
	if i >= 0 && r.Score > (&rankList[i]).Score {
		tmpRank := rankList[i]
		rankList[i] = singleRank
		if i < RankNum-1 {
			rankList[i+1] = tmpRank
		}
		i--
	}
}

func signInFromON(msg *msg.Message) {

}

func InitiateWithON() {

}
