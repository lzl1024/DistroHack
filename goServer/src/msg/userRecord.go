package msg

import (
	"time"
)

type UserRecord struct {
	UserName string
	Score    int
	Ctime    time.Time
}

const GlobalRankSize = 20
