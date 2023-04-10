package leader
import (
	"fmt"
	"net/rpc"
	"log"
	"time"
)

func dialLeader(address string) *rpc.Client {
	client, err := rpc.DialHTTP("tcp", address)
	if err != nil {
		connectTimer := time.NewTimer(800 * time.Millisecond)
		for {
			select {
			case <- connectTimer.C:
				fmt.Println("failed to connect within time limit")
				return nil
			default:
				client, err = rpc.DialHTTP("tcp", address)
				if err == nil {
					return client
				}
			}
		}
	}
	return client
}