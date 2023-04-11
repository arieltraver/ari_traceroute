package main
import (
	"fmt"
	"net/rpc"
	"errors"
	"log"
	"time"
	//"github.com/arieltraver/ari_traceroute/set"
)

type IpArgs struct {
	ProbeId string
}

type IpReply struct {
	Ips []string
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

func getIpRange(leader *rpc.Client) {
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
}

func testGet() {
	address := "localhost:4000"
	leader, err := dialLeader(address)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("connected to:", address)
	getIpRange(leader)
}

func main() {
	testGet()
}