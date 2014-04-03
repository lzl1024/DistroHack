package msg

import (
	"time"
)

type Rank struct {
	Username string
	Score    int
	Time     time.Time
}

const RankNum = 10
