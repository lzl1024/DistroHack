package msg

import (
	"strconv"
	"time"
)

type UserRecord struct {
	UserName string
	Score    int
	Ctime    time.Time
}

const GlobalRankSize = 20

func (userRecord *UserRecord) NewUserRecord(userName string, score int, ctime time.Time) {
	userRecord.UserName = userName
	userRecord.Score = score
	userRecord.Ctime = ctime
}

func (userRecord UserRecord) String() string {
	s := "UserName: " + userRecord.UserName + " " + " Score: " + strconv.Itoa(userRecord.Score) + " time: " +
		userRecord.Ctime.String()
	return s
}
