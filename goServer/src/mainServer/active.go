package main

import (
	"fmt"
	"time"
)

type user_record struct {
	name string
	score int
	time1 time.Time
} 

var global_ranking_size = 20

var global_ranking []user_record
var local_info map[string]user_record

func activeThread() {
	fmt.Println("aaa")
}
