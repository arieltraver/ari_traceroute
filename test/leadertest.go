package main
import (
	"fmt"
	"net/rpc"
	"errors"
	"log"
	"time"
	"github.com/arieltraver/ari_traceroute/set"
	//"os"
)

type ResultArgs struct {
	NewGSS *set.SafeSet
	News *set.StringSet
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

//selects a range of IP addresses for a monitor to probe.
//sends this (currently in the form of a string array, will compress) to the probe
//TODO: also send the stop set associated with that range to the probe
func getIPRange(leader *rpc.Client, id string) int {
	arguments := IpArgs {
		ProbeId:id,
	}
	reply := IpReply{}
	err := leader.Call("Leader.GetIPs", arguments, &reply)
	if err != nil {
		log.Fatal(err)
	}
	for _, ip := range(reply.Ips) {
		fmt.Println(ip)
	}
	return reply.Index
}

func sendIPRange(leader *rpc.Client, index int, id string) {
	newNodes := set.NewStringSet()
	newGSS := set.NewSafeStringSet()
	newNodes.Add("node1_" + id)
	newNodes.Add("node2_" + id)
	newGSS.Add([]byte{byte(123), byte(22), byte(4), byte(200)})
	newGSS.Add([]byte{byte(1), byte(220), byte(43), byte(10)})
	fmt.Println(newGSS.ToCSV())
	arguments := ResultArgs{NewGSS:newGSS, News:newNodes, Id:id, Index:index}
	reply := ResultReply{}
	err := leader.Call("Leader.TransferResults", arguments, &reply)
	if err != nil {
		log.Fatal(err)
	}
	if reply.Ok {
		fmt.Println("Transfer success")
	}
}

func testAll(id string) {
	address := "localhost:4000"
	leader, err := dialLeader(address)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("connected to:", address)

	index := getIPRange(leader, id)
	time.Sleep(10 * time.Second) //turn this on to see if range is freed successfully
	sendIPRange(leader, index, id)
}

func main() {
	/**
	if len(os.Args) <= 1 {
		fmt.Println("usage: go run leadertest.go {id}")
	}
	id := os.Args[1]
	testAll(id)
	**/
	//set.TestNoRoutine()
	set.TestRoutines()
}