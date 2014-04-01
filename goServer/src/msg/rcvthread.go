package msg

import (
	"fmt"
	"net"
	"encoding/gob"
)

func rcvthread(mp *Messagepasser, conn net.Conn) {
	fmt.Println("Started recevier thread\n")
	
	var msg Message
	for {
		decoder := gob.NewDecoder(conn)
		err := decoder.Decode(&msg)
		if err != nil {
			fmt.Println("error while decoding: ", err)
			conn.Close()
			break
		}
		/* add this to the list of connections */
		fmt.Println(msg.String())
	}
}

