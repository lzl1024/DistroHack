package msg

import (
	"sync"
	"fmt"
	"time"
	"sync/atomic"
	"strings"
)

const (
	RELEASED = iota
	HOLD
	WANTED
)

type DsLock struct {
	ilock sync.Mutex
	ackMap map[string]bool
	voted bool
	status int
	counter int
	condvar *sync.Cond
	owner string

}

func (lock *DsLock) Init() {
	lock.status = RELEASED
	lock.counter = 0
	lock.condvar = sync.NewCond(&lock.ilock)
	lock.voted = false
	lock.ackMap = make(map[string]bool)
}

func (lock *DsLock) LockInternal() {
	lock.ilock.Lock()
}

func (lock *DsLock) UnlockInternal() {
	lock.ilock.Unlock()
}

func (lock *DsLock) Lock() {
	lock.ilock.Lock()
	lock.counter++
	if lock.status == HOLD || lock.status == WANTED {
		lock.condvar.Wait()
		lock.status = HOLD
		fmt.Println("Lock: I have got the lock slyly")
		lock.ilock.Unlock()
		return
	}
	lock.status = WANTED

	for lock.status != HOLD {
		newMCastMsg := new(MultiCastMessage)
		newMCastMsg.NewMCastMsgwithData("", SN_SNLOCKREQ, "Lock Request")
		newMCastMsg.HostList = MsgPasser.SNHostlist
		newMCastMsg.Origin = MsgPasser.ServerIP
		newMCastMsg.Seqnum = atomic.AddInt32(&MsgPasser.SeqNum, 1)
		MsgPasser.SendMCast(newMCastMsg)
		lock.ilock.Unlock()
		time.Sleep(time.Duration(3) * time.Second)
		lock.ilock.Lock()
	}
	lock.owner = MsgPasser.ServerIP
	lock.ilock.Unlock()
	fmt.Println("Lock: I have got the lock")
}

func RcvSnLockReq (msg *Message) (interface{}, error) {
	DLock.ilock.Lock()
	if DLock.voted == true || DLock.status == HOLD {
		/* do nothing, the person will retry after a bit */
	} else {
		DLock.voted = true
		DLock.owner = msg.Origin
		m := new(Message)
		m.NewMsgwithData(msg.Origin, SN_SNLOCKACK, "Lock Acked")
		err := MsgPasser.Send(m)
		if err != nil {
			DLock.voted = false
			DLock.owner = ""
		}
	}
	DLock.ilock.Unlock()
	return msg, nil
}

func RcvSnLockAck (msg *Message) (interface{}, error) {
	DLock.ilock.Lock()
	DLock.ackMap[msg.Src] = true
	var rcvdAll bool
	rcvdAll = true
	for k,_ := range MsgPasser.SNHostlist {
		val, ok := DLock.ackMap[k]
		if !ok || val == false {
			rcvdAll = false
			break;
		}
	}
	if rcvdAll == true {
		for k,_ := range DLock.ackMap {
			delete(DLock.ackMap, k)
		}
		DLock.status = HOLD
	}
	DLock.ilock.Unlock()
	return msg, nil
}

func (lock *DsLock) Unlock() {
	lock.ilock.Lock()
	lock.counter--
	if lock.counter != 0 {
		lock.condvar.Signal()
		lock.ilock.Unlock()
		fmt.Println("Unlock: I have released the lock slyly")
		return;
	}
	/* here counter == 0 */
	lock.status = RELEASED
	newMCastMsg := new(MultiCastMessage)
	newMCastMsg.NewMCastMsgwithData("", SN_SNLOCKREL, "Lock Released")
	newMCastMsg.HostList = MsgPasser.SNHostlist
	newMCastMsg.Origin = MsgPasser.ServerIP
	newMCastMsg.Seqnum = atomic.AddInt32(&MsgPasser.SeqNum, 1)
	MsgPasser.SendMCast(newMCastMsg)
	fmt.Println("Unlock: I have released the lock")
	lock.ilock.Unlock()
}

func RcvSnLockRel (msg *Message) (interface{}, error) {
	DLock.ResetLock(msg.Origin)
	return msg, nil
}

func (lock *DsLock)ResetLock(owner string) {
	DLock.ilock.Lock()
	if strings.EqualFold(owner, lock.owner) == true {
		lock.voted = false
		lock.owner = ""
	}
	DLock.ilock.Unlock()
}
