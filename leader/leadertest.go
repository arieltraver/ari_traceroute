package leader
import (
	"fmt"
	"net/rpc"
	"log"
	"time"
	"github.com/arieltraver/ari_traceroute/set"
)

func dialLeader(address string) *rpc.Client {
	client, err := rpc.DialHTTP("tcp", address)
	if err != nil {
		connectTimer := time.NewTimer(800 * time.Millisecond)
		for {
			select {
			case <- connectTimer.C:
				log.Println("failed to connect within time limit")
				return nil
			default:
				client, err = rpc.DialHTTP("tcp", address)
				if err == nil {
					return client
				} else {
					log.Println(err)
				}
			}
		}
	}
	return client
}

func getIpRange(leader *rpc.Client) {
	arguments := IpArgs {
		probeId:"test",
	}
	reply := IpReply{}
	err := leader.Call("RaftNode.RequestVote", arguments, &reply)
	if err != nil {
		return
	}
	for _, ip := range(reply.ips) {
		fmt.Println(ip)
	}
}