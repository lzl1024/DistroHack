import (
	"encoding/json"
	"fmt"
	"net"
)

func handleConnectionFromPeers(listener, net.Listener) {
	for {
		conn, err := listener.Accept()

		if err != nil {
			fmt.Println("Accept Error:", err.Error())
			return
		}

		// new thread to handle request from peers
		go handleConnectionFromPeersThread(conn)
	}
}

func handleConnectionFromPeersThread(conn, net.Conn) {

}