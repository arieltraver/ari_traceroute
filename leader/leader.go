package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"sync"
)

var MONITORS int = 5 //number of chunks to divide file into

/*
// a locked file, from which data will be sent to workers
type SafeRanges struct {
	ranges []bool
	lock sync.Mutex
}

func findOpen(global *SafeRanges, local []bool) int {
	global.lock.Lock()
	ranges := global.ranges
	for i, open := range(ranges) {
		if open && local[i] {
			ranges[i] = false
			local[i] = false
			global.lock.Unlock()
			return i
		}
	}
	global.lock.Unlock()
	return -1
}
*/

// Talks to a single remote worker. Upon receiving a "ready" keyword, if there
// are remaining file chunks, sends the worker a "map words" keyword and waits to
// receive "ok map" confirmation keyword, both through sendJobname().
// Upon receiving the worker's confirmation, grabs a file chunk and sends
// it to the worker. If there are no file chunks left, communicates this
// through a channel, writes "DONE" keyword to workers, closes the connection,
// and returns.
func handleConnection(c net.Conn, data chan int, wait *sync.WaitGroup) {
	defer wait.Done()
	count := 0
	alldone := make(chan bool, 1)
	for {
		select {
		case <- data:
			fmt.Println("...")
			count++
			if count >= MONITORS {
				alldone <- true
			}
			wait.Done() //wait will be incremented every time data is filled
		case <- alldone:
			c.Close()
			return
		}
	}
}

// Waits for new connections on port (specified by net.Listener). Serves each
// worker with a different goroutine.
func waitOnConnections(listener net.Listener, complete chan bool) {
	var wait sync.WaitGroup

	//allocate a channel for each worker
	//new ranges of IPs will be sent on this channel
	dataRanges := make([]chan int, MONITORS)
	for i, _ := range dataRanges {
		dataRanges[i] = make(chan int, 1)
	}

	//connect to all monitors and assign each a channel
	for _, data := range(dataRanges) {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("failed connection")
		} else { //if one connection fails you can have more
			fmt.Println("new host joining:", conn.RemoteAddr())
			go handleConnection(conn, data, &wait) // each client served by a different routine
		}
	}

	//give an IP range to each monitor
	//wait for all IPs to complete their range
	//cycle through so that all monitors see all ranges
	for i := 0; i < MONITORS; i++ {
		j := i
		for _, c := range(dataRanges) {
			wait.Add(1)
			c <- j
			j++
			if j >= MONITORS {
				j = 0
			}
		}
		wait.Wait() //wait for all monitors to complete their section
	}
	complete <- true
}


func main() {
	arguments := os.Args
	if len(arguments) < 2 {
		fmt.Println("Usage: 'leader port'")
		return
	}
	PORT := ":" + arguments[1]
	listener, err := net.Listen("tcp4", PORT)
	if err != nil {log.Fatal(err)}
	fmt.Println("listening on port", arguments[1])

	complete := make(chan bool, 1)
	// initialize a global file
	go waitOnConnections(listener, complete)

	if <-complete {
		fmt.Println("done")
	}


}
