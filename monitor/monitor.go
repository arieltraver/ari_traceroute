//used tutorial here: https://www.linode.com/docs/guides/developing-udp-and-tcp-clients-and-servers-in-go/

package main

import (
	"io"
	"log"
	"net"
	"os"
	"fmt"
)

func checkErrConn(c net.Conn, err error) {
	if err != nil {
		c.Close()
		log.Fatal(err)
	}
}

// Sends the "ready" keyword to the leader.
func sayhi(c net.Conn) {
	block := make(chan int, 1)
	_, err := io.WriteString(c, "hi\n") //send text to your connection
	checkErrConn(c, err)
	fmt.Println(<-block) //blocks forever, testing purposes
}

//connect to host
func main() {
	args := os.Args
	if len(args) <= 1 {
		log.Fatal("please provide host:port to connect to")
	}
	conn := args[1]
	c, err := net.Dial("tcp", conn) // connect to host:port
	if err != nil { log.Fatal(err) }
	defer c.Close() // make sure it closes

	sayhi(c)
}
