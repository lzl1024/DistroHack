package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
	"msg"
)

// main thread for active connection
func activeTest() {
	sendoutGlobalRanking()
	
	fmt.Printf("ActiveTest\n")
}


// send out the updated global info to application
func sendoutGlobalRanking() {
	// fake global_ranking
	global_ranking := []msg.User_record{}
	global_ranking = append(global_ranking,
		msg.User_record{UserName: "2", Score: 3, Ctime: time.Now()})
	global_ranking = append(global_ranking,
		msg.User_record{UserName: "1", Score: 2, Ctime: time.Now()})

	data, _ := json.Marshal(global_ranking)

	// send data out
	sendout(msg.App_url+"hacks/update_rank/", string(data))

}

// truely send out date
func sendout(urlAddress string, data string) {
	_, err := http.PostForm(urlAddress,
		url.Values{"data": {data}})

	if err != nil {
		fmt.Println("Post failure: " + urlAddress + "," + data)
	}
}
