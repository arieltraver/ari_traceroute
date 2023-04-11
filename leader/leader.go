package main

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
	return nil, -1, errors.New("no free IPs")
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
		return errors.New("ips in use by other probe")
	}
	thisRange.currentProbe = "" //no id associated here anymore

	thisRange.stops.UnionWith(args.NewGSS) //register new (hop, dest) pairs to this range of IPs
	allIPS.UnionWith(args.News) //register all new, never-before-seen nodes
	//TODO register new edges in some kind of graph data structure

	unlockPlease[args.Index] <- true //request to unlock this set, a routine is listening.
	return nil
}

func (*Leader) GetIPs(args IpArgs, reply *IpReply) error {
	probeId := args.ProbeId
	ips, index, _ := findNewRange(probeId)
	//TODO error handling
	reply.Ips = ips //node gets this
	fmt.Println(index)
	//go waitOnProbe(args.probeId, index) //wait for probe to either time out, or finish.
	return nil
}

func waitOnProbe(probeId string, index int) error {
	probeTimer := time.NewTimer(60 * time.Second)
	for {
		select {
		case <- unlockPlease[index]: //second http request occured, result stored
			return nil
		case <- probeTimer.C:
			log.Println("probe took too long")
			return errors.New("probe timeout")
		}
	}
}

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
	ips1 := []string {
		"192.124.249.164", //bugsincyberspace.com
		"129.186.120.3", //bugguide.net
		"172.67.199.120", //buglife.org.uk
	}
	ipRange1 := &ipRange{ips:ips1[:], stops:set.NewSafeSet(), currentProbe:""}

	ips2 := []string {
		"13.35.83.221", //code.org
		"104.18.8.221", //codeacademy.com
		"76.223.115.82", //w3schools.com
	}
	ipRange2 := &ipRange{ips:ips2[:], stops:set.NewSafeSet(), currentProbe:""}

	ipTable = []*ipRange{ipRange1, ipRange2} //add ips to global data structure

	go connect("localhost:4000")

}

func main() {
	test();
	time.Sleep(600 * time.Second)
}