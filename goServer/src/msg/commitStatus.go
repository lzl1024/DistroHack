package msg

import (
	"time"
	"sync"
)

// sign up 2 phase commit
// As master
type SignUpCommitStatus struct {
	HasAbort		bool			 // has got abort?
	ReadySNIP		map[string]bool  // ready SN ip set	
}
var signUp_requestMap map[string]*SignUpCommitStatus // map of user name who sent signup request

// As worker
var signUp_commit_readySet map[string]bool // map of user name that worker has been ready
var signUp_commitLock sync.Mutex

const sleepInterval = time.Millisecond * time.Duration(50)
const timeOutRound = 40

func (commitStatus *SignUpCommitStatus) NewSignUpCmitStatus() {
	commitStatus.HasAbort = false
	commitStatus.ReadySNIP = make(map[string]bool)
}


// check number of readys get from other SNs
func checkCommitStatus(commitStatusChan chan string, userName string) {
	for i := 0; i <= timeOutRound; i++ {
		time.Sleep(sleepInterval)
		
		// check status fail
		if signUp_requestMap[userName].HasAbort {
			commitStatusChan <- "Abort"
			return
		} else if len(signUp_requestMap[userName].ReadySNIP) == len(MsgPasser.SNHostlist) {
			commitStatusChan <- "Commit"
			return
		}
	}
	
	// timeout: abort
	commitStatusChan <- "Abort"
}