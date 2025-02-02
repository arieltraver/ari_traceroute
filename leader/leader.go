package main

//IPSet: a bitset in which an IP address (or a list of bytes) maps directly to an array index
//This gives us a set of size 2^(32/(2^n)) where n is the number of IP address chunsk you leave out
//For example, a monitor could be given [111 200 301] and it will map all IPS with those first 3 parts, "111.200.300.(...)"

import (
	"net/rpc"
	"net/http"
	"sync"
	"github.com/arieltraver/ari_traceroute/set"
	"time"
	"log"
	"errors"
	"fmt"
)

const MONITORS int = 5 //number of chunks to divide file into
const CHUNKS int = 10
var allIPs *set.SafeSet
var unlockPlease []chan bool
var ipTable []*ipRange //here is where the global stop sets are stored
var seenRanges *seenMap //keeps track of IPs and which has seen what

//a pair: who's using an IP range (locked for concurrency), and also that range (locked)
type ipRange struct {
	addresses [][4]byte //must be the same length as stops, 1-1 correspondence.
	currentProbe  string
	stops *set.StringSet
	lock sync.Mutex
}
/*
func (i *ipRange) Size() int {
	i.lock.Lock()
	defer i.lock.Unlock()
	l := i.stops.
	return l
}*/

type Leader int

type ResultArgs struct {
	NewGSS *set.StringSet
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
	Ips [][4]byte
	Stops *set.StringSet
	Index int
	Ok bool
}

//each id is associated with an index in the table.
//the table records which probes have already hit which addresses.
type seenMap struct {
	rangesSeenBy map[string]*set.IntSet
	lock sync.Mutex
}

//given the id of a probe, finds an unseen range and returns its ip's and stop set.
func findNewRange(id string) ([][4]byte, *set.StringSet, int, error) {
	seenRanges.lock.Lock()
	seenRanges.lock.Unlock() //TODO: lock the range but not the whole table.
	//TODO: empty set check. return error "{id} has seen all ip ranges"
	
	indexes, ok := seenRanges.rangesSeenBy[id]

	//TODO: make registration separate?
	if !ok { //this probe is New, register seen ip ranges for it
		r := set.NewIntSet()
		l := len(ipTable)
		for i := 0; i < l; i++ {
			r.Add(i)
		}
		seenRanges.rangesSeenBy[id] = r
		indexes = seenRanges.rangesSeenBy[id]
	}

	//TODO make set iteratable
	for index, _ := range(indexes.Mp) {
		thisRange := ipTable[index] //check if each unseen range is in use
		thisRange.lock.Lock()
		if thisRange.currentProbe == "" { //no current owner
			thisRange.currentProbe = id //new owner
			stopSet := thisRange.stops //copy range of IP addresses from table
			addressesToProbe := thisRange.addresses;
			thisRange.lock.Unlock()
			return addressesToProbe, stopSet, index, nil
		}
		thisRange.lock.Unlock()
	}
	//everything in use
	return nil, nil, -1, errors.New("no free IPs")
}

//accepts results of a trace from a node.
func (*Leader) TransferResults(args ResultArgs, reply *ResultReply) error {
	
	thisRange := ipTable[args.Index] //look in the table for the ip range
	thisRange.lock.Lock()
	defer thisRange.lock.Unlock()

	//check if this node actually was registered with this range.
	rangeOwner := thisRange.currentProbe
	if rangeOwner != args.Id {
		reply.Ok = false
		if rangeOwner == "" {
			return errors.New("you took too long")
		}
		return errors.New("ips in use by other probe")
	}
	fmt.Println("result:", args.NewGSS.ToCSV()) //TODO: remove this test
	thisRange.currentProbe = "" //no id associated here anymore
	thisRange.stops.UnionWith(args.NewGSS) //register new (hop, dest) pairs to this range of IPs
	allIPs.UnionWith(args.News) //register all new, never-before-seen nodes
	seenRanges.lock.Lock()
	defer seenRanges.lock.Unlock()
	seenRanges.rangesSeenBy[args.Id].Remove(args.Index) //done w this range!

	//TODO register new edges in some kind of graph data structure

	unlockPlease[args.Index] <- true //request to unlock this set, a routine is listening.
	reply.Ok = true
	return nil
}

//RPC which assigns a range of IP's to a monitor, depending on which are free.
func (*Leader) GetIPs(args IpArgs, reply *IpReply) error {
	ips, stops, index, er := findNewRange(args.ProbeId)
	if er != nil {
		reply.Ok = false
		return errors.New("could not find new IP range for that node.")
	}
	fmt.Println("ip range is:", ips)
	reply.Ips = ips //node gets this
	reply.Stops = stops
	reply.Index = index
	reply.Ok = true
	fmt.Println("index selected:", index, "for", args.ProbeId)
	go waitOnProbe(args.ProbeId, index) //wait for probe to either time out, or finish.
	return nil
}

/*Waits for a probe to return. If it doesn't return in time, it frees up its range.*/
func waitOnProbe(probeId string, index int) error {
	probeTimer := time.NewTimer(90 * time.Second)
	for {
		select {
		case <- unlockPlease[index]: //second http request occured, result stored
			fmt.Println(allIPs.ToCSV()) //TODO remove, this is test
		case <- probeTimer.C:
			log.Println("probe took too long")
			go freeRange(index) //free the range, change the id in case the probe comes back later
			return errors.New("probe timeout")
		}
	}
}

//frees up a range lent to a monitor that timed out, removing that monitor's id.
func freeRange(index int) {
	thisRange := ipTable[index] //look in the table for the ip range
	thisRange.lock.Lock()
	defer thisRange.lock.Unlock()
	thisRange.currentProbe = ""
}


//set up http server
func connect(port string) {
	api := new(Leader)
	err := rpc.Register(api)
	if err != nil {
		log.Fatal("error registering the RPCs", err)
	}
	rpc.HandleHTTP()
	go http.ListenAndServe(port, nil)
	log.Printf("serving rpc on port " + port)
}

func test(numRanges int) {
	ipTable = make([]*ipRange,numRanges)
	for i := 0; i < numRanges; i++ {
		b := make([][4]byte, 1)
		b[0] = [4]byte{byte(i), byte(i), byte(i), byte(i)}
		stopz := set.NewStringSet()
		ipTable[i] = &ipRange{addresses:b, stops:stopz, currentProbe:""}
	}
	seen := make(map[string]*set.IntSet)
	seenRanges = &seenMap{rangesSeenBy:seen} //TODO make this readable
	allIPs = set.NewSafeStringSet()
	unlockPlease = make([]chan bool, numRanges)
	for i, _ := range(unlockPlease) {
		unlockPlease[i] = make(chan bool, 1)
	}
	go connect("localhost:4000")
	time.Sleep(120 * time.Second)

}

func main() {
	test(10)
}