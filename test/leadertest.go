package main
import (
	"fmt"
	"net/rpc"
	"errors"
	"log"
	"time"
	"github.com/arieltraver/ari_traceroute/set"
)

type ResultArgs struct {
	NewGSS *set.Set
	News *set.Set
	Id string
	Index int
}

type ResultReply struct {
	Ok bool
}

type IpArgs struct {
	ProbeId string
}

type IpReply struct {
	Ips []string
	Index int
}


func dialLeader(address string) (*rpc.Client, error) {
	client, err := rpc.DialHTTP("tcp", address)
	if err != nil {
		connectTimer := time.NewTimer(800 * time.Millisecond)
		for {
			select {
			case <- connectTimer.C:
				return nil, errors.New("failed to connect within time limit")
			default:
				client, err = rpc.DialHTTP("tcp", address)
				if err == nil {
					return client, nil
				} else {
					return nil, err
				}
			}
		}
	}
	return client, nil
}

func getIPRange(leader *rpc.Client) int {
	arguments := IpArgs {
		ProbeId:"test",
	}
	reply := IpReply{}
	err := leader.Call("Leader.GetIPs", arguments, &reply)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("no error")
	for _, ip := range(reply.Ips) {
		fmt.Println(ip)
	}
	return reply.Index
}

func sendIPRange(leader *rpc.Client, index int) {
	newNodes := set.NewSet()
	newGSS := set.NewSet()
	newNodes.Add("node1")
	newNodes.Add("node2")
	newGSS.Add("example1")
	newGSS.Add("example2")
	arguments := ResultArgs{NewGSS:newGSS, News:newNodes, Id:"test", Index:index}
	reply := ResultReply{}
	err := leader.Call("Leader.TransferResults", arguments, &reply)
	if err != nil {
		log.Fatal(err)
	}
	if reply.Ok {
		fmt.Println("Transfer success")
	}
}

func testAll() {
	address := "localhost:4000"
	leader, err := dialLeader(address)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("connected to:", address)

	index := getIPRange(leader)
	sendIPRange(leader, index)

}

func main() {
	testAll()
}