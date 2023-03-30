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

// Says hi
func sayhi(c net.Conn) {
	_, err := io.WriteString(c, "hi\n") //send text to your connection
	fmt.Println("hi")
	checkErrConn(c, err)
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
