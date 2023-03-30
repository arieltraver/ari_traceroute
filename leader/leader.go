package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"
)

var MONITORS int = 5 //number of chunks to divide file into
var CHUNKS int = 10

// a locked channel
type SafeChan struct {
	ranges chan int
	lock sync.Mutex
}

func awaitPop(sc *SafeChan) int {
	sc.lock.Lock()
	c := <- sc.ranges
	sc.lock.Unlock()
	return c
}

func awaitPush(sc *SafeChan, i int) {
	sc.lock.Lock()
	sc.ranges <- i
	sc.lock.Unlock()
}

//blocks till a slice is available for processing, puts it back if seen already
func popCheckPush(sc *SafeChan, seen []bool) int {
	sc.lock.Lock()
	i := <- sc.ranges
	if seen[i] {
		sc.ranges <- i //put it back, we've seen it before.
		sc.lock.Unlock()
		return -1
	} else {
		seen[i] = true
		sc.lock.Unlock()
		return i //this is the chunk to work on
	}
}

// Talks to a single remote monitor.
func handleConnection(c net.Conn, sc *SafeChan, wait *sync.WaitGroup, start int) {
	defer wait.Done()
	count := 0
	seen := make([]bool, CHUNKS) //which chunks have we seen
	for {
		if count >= CHUNKS {
			c.Close()
			return
		}
		//loop repeats as it digs through s for an unread slice
		s := popCheckPush(sc, seen) //take a number off, have we seen it? if not, put it back
		if s >= 0 { //we have not seen it
			count ++
			fmt.Println(s)
			start = s
			time.Sleep(time.Millisecond * 50)
			awaitPush(sc, start) //push the last number onto the channel
		}
	}
}

// Waits for new connections on port (specified by net.Listener). Serves each
// worker with a differecnt goroutine.
func waitOnConnections(listener net.Listener, complete chan bool) {
	var wait sync.WaitGroup
	c := make(chan int, CHUNKS + 1) //store free chunks in a queue
	sc := &SafeChan{ranges:c} //lock to prevent data races

	//connect to all monitors
	for i := 0; i < MONITORS; i++ {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("failed connection")
		} else { //if one connection fails you can have more
			fmt.Println("new host joining:", conn.RemoteAddr())
			wait.Add(1)
			go handleConnection(conn, sc, &wait, i) // each client served by a different routine
		}
	}

	//add chunks to the channel once all hosts are connected
	for i := 0; i < CHUNKS; i++ {
		c <- i
	}

	wait.Wait()
	//give an IP range to each monitor
	//wait for all IPs to complete their range
	//cycle through so that all monitors see all ranges
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
