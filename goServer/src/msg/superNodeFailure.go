package msg

import (
	"fmt"
	"time"
	"errors"
)

var ON_ACK_Set map[string]bool // map of ON that want me to be SN
var hasPrefered bool // has grant other ON to be leader
var electionOnGoing bool // whether I am trying to find SN


// main routine handling SN fails
func SNFailure() {
	fmt.Println("Haha! My SN fails")
	
	// use a go routine to find new SN
	go findNewSN()
}


// if new SN not find in some time, try again until find one
// in case of the Highest Priority one fail and block
func findNewSN() {
	// wipe out SuperNodeIP as a marker to indicate whether SN is found
	SuperNodeIP = ""
	electionOnGoing = true

	for SuperNodeIP == "" {
		// initialize election tools
		ON_ACK_Set = make(map[string]bool)
		hasPrefered = false
		
		// I want to be a leader!
		electionMsg := new(Message)
		err := electionMsg.NewMsgwithData("", ON_ON_ELECTION, MsgPasser.ServerIP)
		if err != nil {
			fmt.Println("On SuperNode failure: ", err)
			return
		}
	
		// send message to ONs
		MulticastMsgInGroup(electionMsg, false)
		
		// check election status
		i := 0 
		for i = 0; i <= BusyWaitingTimeOutRound; i++ {
			time.Sleep(BusyWaitingSleepInterval)
			
			if SuperNodeIP != "" || hasPrefered {
				break
			} else {
				// check accept using one by one compare, instead of using the size
				// of the set, in case of other ON fails
				win := true
				for ONname := range MsgPasser.ONHostlist {
					if _, exist := ON_ACK_Set[ONname]; !exist {
						win = false
						break
					}
				}
				
				// I am the Leader! Send msg to my ONs
				if win {
					fmt.Println("I am the Leader!!!")
					leaderMsg := new(Message)
					err := leaderMsg.NewMsgwithData("", ON_ON_LEADER, MsgPasser.ServerIP)
					if err != nil {
						fmt.Println("On Leader notify: ", err)
						return
					}
	
					// send message to ONs
					MulticastMsgInGroup(leaderMsg, false)
					
					// overwrite SuperNodeIP of myself at this time to bootstrap
					SuperNodeIP = MsgPasser.ServerIP
					
					BootStrapSN()
					break;
				}
			}	
		}
				
		// sleep 2.5s in total
		if SuperNodeIP == "" {
			time.Sleep(BusyWaitingSleepInterval * (BusyWaitingTimeOutRound + 10 - time.Duration(i)))
		}
	}
	
	electionOnGoing = false
}


// receive election msg
func RcvONElection(msg *Message) (interface{}, error) {
	if msg.Kind != ON_ON_ELECTION {
		return nil, errors.New("message Kind indicates not a ON_ON_ELECTION")
	}
	
	var ip string
	err := ParseRcvInterfaces(msg, &ip)
	if err != nil {
		fmt.Println("In RcvONElection: ")
		return nil, err
	}
	
	// TODO: directly compare IP now, should be changed to some fix priority number
	// generated randomly when server starts
	if ip >= MsgPasser.ServerIP {
		if ip != MsgPasser.ServerIP {
			hasPrefered = true		
		}
		
		// ack potential leader
		leaderACKMsg := new(Message)
	
		err := leaderACKMsg.NewMsgwithData(msg.Origin, ON_ON_ELECTION_ACK, MsgPasser.ServerIP)
		if err != nil {
			fmt.Println("In RcvONElection:")
			return nil, err
		}
		MsgPasser.Send(leaderACKMsg)
		
	} else if !electionOnGoing {
		// election SN has lower priority than me, and I haven't try to find new
		// election, start a new election
		go findNewSN()
	}
	
	return ip, nil
}


// receive a leader's success message
func RcvONLeader(msg *Message) (interface{}, error) {
	if msg.Kind != ON_ON_LEADER {
		return nil, errors.New("message Kind indicates not a ON_ON_LEADER")
	}
	
	var ip string
	err := ParseRcvInterfaces(msg, &ip)
	if err != nil {
		fmt.Println("In RcvONLeader: ")
		return nil, err
	}
	
	// commit leader's supernode ip
	SuperNodeIP = ip
	
	return ip, nil
}


// get the election ack from other on
func RcvElectionACK(msg *Message) (interface{}, error) {
	if msg.Kind != ON_ON_ELECTION_ACK {
		return nil, errors.New("message Kind indicates not a ON_ON_ELECTION_ACK")
	}
	
	var ip string
	err := ParseRcvInterfaces(msg, &ip)
	if err != nil {
		fmt.Println("In RcvElectionACK: ")
		return nil, err
	}
	
	// collect other ON's IP
	ON_ACK_Set[ip] = true
	
	return ip, nil
}
