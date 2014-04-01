package msg

import (
	"fmt"
)

const SERVER_PORT = 5989

func TestMessagePasser() {
	channel := make(chan error)

	go Clienthread(MsgPasser, channel)
	value := <- channel
	fmt.Println(value)
}

