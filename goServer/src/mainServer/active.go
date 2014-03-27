package main

import (
	"fmt"
	"net/http"
	"encoding/json"
	"time"
	"net/url"
)

type User_record struct {
	UserName string
    Score   int
    Ctime time.Time
}

var app_url = "http://localhost:8000/"

var global_ranking_size = 20

var global_ranking []User_record
var local_info map[string]User_record

// main thread for active connection
func activeThread() {
	sendoutLocalInfo()
	sendoutGlobalRanking()
}


// send out the updated local info to application
func sendoutLocalInfo() {
	// fake data
	local_info = map[string]User_record{
    	"1": User_record{
        	"1", 1, time.Now(),
    	},
    	"2": User_record{
        	"2", 3, time.Now(),
    	},
    }
    
    data, _ := json.Marshal(local_info)
    
    // send data out
    sendout(app_url+"hacks/update_local/", string(data))
}

// send out the updated global info to application
func sendoutGlobalRanking() {
	// fake global_ranking
    global_ranking = append(global_ranking, 
    	User_record{UserName: "2", Score: 3, Ctime:time.Now()})
    global_ranking = append(global_ranking, 
    	User_record{UserName: "1", Score: 2, Ctime: time.Now()})

    data, _:= json.Marshal(global_ranking)
    
    // send data out
    sendout(app_url+"hacks/update_rank/", string(data))
    
}

// truely send out date
func sendout(urlAddress string, data string) {
	_, err := http.PostForm(urlAddress, 
		url.Values{"data": {data}})
	
	if err != nil {
		fmt.Println("Post failure: " + urlAddress + "," + data)
	}
}
