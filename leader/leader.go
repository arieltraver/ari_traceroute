package leader

import (
	"os"
	"sync"
	"github.com/arieltraver/ari_traceroute/set"
)

var MONITORS int = 5 //number of chunks to divide file into
var CHUNKS int = 10
var allIPS *set.SafeSet
var unlockPlease []chan bool
var ipTable []*ipRange 
var seenRanges *seenMap //keeps track of IPs and which has seen what

//a pair: who's using an IP range (locked for concurrency), and also that range (locked)
type ipRange struct {
	ips []string
	currentProbe  string
	stops *set.SafeSet
	lock sync.Mutex
}

type Leader int

type ResultArgs struct {
	NewGSS *set.Set
	News *set.Set
	id string
	index int
}

type ResultReply struct {
	ok bool
}

type IpArgs struct {
	probeId string
}

type IpReply struct {
	ips []string
}


//each id is associated with an index in the table.
//the table records which probes have already hit which addresses.
type seenMap struct {
	rangesSeenBy map[string]*set.IntSet
	lock sync.Mutex
}

//given the id of a probe, finds an unseen range for it.
func findNewRange(id string) ([]string, int, error) {
	seenRanges.lock.Lock()
	seenRanges.lock.Unlock() //TODO: lock the range but not the whole table.
	//TODO: empty set check. return error "{id} has seen all ip ranges"
	indexes := seenRanges.rangesSeenBy[id]

	//TODO make set iteratable
	for index, _ := range(indexes.Mp) {
		thisRange := ipTable[index] //check if each unseen range is in use
		thisRange.lock.Lock()
		if thisRange.currentProbe == "" { //no current owner
			thisRange.currentProbe = id //new owner
			freeIps := thisRange.ips //copy range of IP addresses from table
			thisRange.lock.Unlock()
			return freeIps, index, nil
		}
		thisRange.lock.Unlock()
	}
	//everything in use
	return nil, -1, os.ErrClosed //TODO replace with custom error saying "everything in use"
}

//accepts results of a trace from a node.
func (*Leader) TransferResults(args ResultArgs, reply *ResultReply) error {

	thisRange := ipTable[args.index] //look in the table for the ip range
	thisRange.lock.Lock()
	defer thisRange.lock.Unlock()

	//check if this node actually was registered with this range.
	rangeOwner := thisRange.currentProbe
	if rangeOwner != args.id {
		reply.ok = false
		return os.ErrInvalid //TODO replace with custom error, "ips being probed by another node"
	}
	thisRange.currentProbe = "" //no id associated here anymore

	thisRange.stops.UnionWith(args.NewGSS) //register new (hop, dest) pairs to this range of IPs
	allIPS.UnionWith(args.News) //register all new, never-before-seen nodes
	//TODO register new edges in some kind of graph data structure

	unlockPlease[args.index] <- true //request to unlock this set, a routine is listening.
	return nil
}

func (*Leader) GetIPs(args IpArgs, reply *IpReply) error {
	probeId := args.probeId
	ips, index, _ := findNewRange(probeId)
	//TODO error handling
	reply.ips = ips //node gets this
	go waitOnProbe(args.probeId, index) //wait for probe to either time out, or finish.
	return nil
}

func waitOnProbe(probeId string, index int) error {
	//make timer
	for {
		select {
		case <- unlockPlease[index]: //second http request occured, result stored
			return nil
		}
		//TODO: case <- timer:
			//unlock the range
			//return error "probe took too long"
	}

}


