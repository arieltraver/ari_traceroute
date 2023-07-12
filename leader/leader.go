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

const BITSETSIZE uint = 16 //2^16 bits --> one bit for each possible IP, sliced by first two bytes.
const MONITORS int = 5 //number of chunks to divide file into
const CHUNKS int = 10
var allIPs *set.SafeSet
var unlockPlease []chan bool
var ipTable []*ipRange //here is where the global stop sets are stored
var seenRanges *seenMap //keeps track of IPs and which has seen what

//a pair: who's using an IP range (locked for concurrency), and also that range (locked)
type ipRange struct {
	prefixes [][]byte //must be the same length as stops, 1-1 correspondence.
	currentProbe  string
	stops *set.Set
	lock sync.Mutex
}

func (i *ipRange) Size() int {
	i.lock.Lock()
	defer i.lock.Unlock()
	l := i.stops.
	return l
}

type Leader int

type ResultArgs struct {
	NewGSS *set.SafeSet
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


//each id is associated with an index in the table.
//the table records which probes have already hit which addresses.
type seenMap struct {
	rangesSeenBy map[string]*set.IntSet
	lock sync.Mutex
}

//given the id of a probe, finds an unseen range and returns its prefixes and stop set.
func findNewRange(id string) ([][]byte, []*set.BitSet, int, error) {
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
			prefixesToProbe := thisRange.prefixes;
			thisRange.lock.Unlock()
			return prefixesToProbe, stopSet, index, nil
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
	thisRange.currentProbe = "" //no id associated here anymore
	args.NewGSS.lock.Lock()
	thisRange.stops.UnionWith(args.NewGSS.Set()) //register new (hop, dest) pairs to this range of IPs
	args.NewGSS.lock.
	allIPs.UnionWith(args.News) //register all new, never-before-seen nodes
	seenRanges.lock.Lock()
	defer seenRanges.lock.Unlock()
	seenRanges.rangesSeenBy[args.Id].Remove(args.Index) //done w this range!

	//TODO register new edges in some kind of graph data structure

	unlockPlease[args.Index] <- true //request to unlock this set, a routine is listening.
	reply.Ok = true
	return nil
}

func (*Leader) GetIPs(args IpArgs, reply *IpReply) error {
	ips, index, _ := findNewRange(args.ProbeId)
	//TODO error handling
	reply.Ips = ips //node gets this
	reply.Index = index
	fmt.Println("index selected:", index, "for", args.ProbeId)
	go waitOnProbe(args.ProbeId, index) //wait for probe to either time out, or finish.
	return nil
}

func waitOnProbe(probeId string, index int) error {
	probeTimer := time.NewTimer(60 * time.Second)
	for {
		select {
		case <- unlockPlease[index]: //second http request occured, result stored
			fmt.Println(allIPs.ToCSV()) //TODO remove, this is test
			//TODO call function which frees up the range again
		case <- probeTimer.C:
			log.Println("probe took too long")
			go freeRange(index)
			return errors.New("probe timeout")
		}
	}
}


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

func test() {
	stopz := make([]*set.IPSet, 65535) //255: 16 bits. ranges specified per byte
	for i, _ := range(stopz) {
		stopz[i] = set.NewIPSet(BITSETSIZE)
		/*so what we are doing here
		is adding this special data structure
		(a bitset with index directly correlating to IP address)
		*/
	}
	seen := make(map[string]*set.IntSet)
	seenRanges = &seenMap{rangesSeenBy:seen} //TODO make this readable
	allIPs = set.NewSafeSet()
	unlockPlease = make([]chan bool, len(ipTable))
	for i, _ := range(unlockPlease) {
		unlockPlease[i] = make(chan bool, 1)
	}


	go connect("localhost:4000")

}

func main() {
	stopz := make([]*set.SafeIPSet, 255) //limit to a range.
	for i, _ := range(stopz) {
		stopz[i] = set.NewSafeIPSet(BITSETSIZE)
	}

	test();
	time.Sleep(600 * time.Second)
}