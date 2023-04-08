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
	idToIndex map[string]int
	seenRanges []*set.IntSet //TODO replace with int set
	lock sync.Mutex
}

//given the id of a probe, finds an unseen range for it.
func (s *seenMap) findNewRange(id string) ([]string, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	i := s.idToIndex[id]
	ranges := s.seenRanges[i]
	//TODO: empty set check.
	//TODO make set interface, enumerable, etc etc
	for index := range(ranges.Mp) {
		thisRange := ipTable[index]
		thisRange.lock.Lock()
		if thisRange.currentProbe == "" {
			thisRange.currentProbe = id
			freeIps := thisRange.ips
			thisRange.lock.Unlock()
			return freeIps, nil
		}
		thisRange.lock.Unlock()
	}
	return nil, os.ErrClosed //TODO replace with custom error
}

//accepts results of a trace from a node.
func (*Leader) TranserResults(args ResultArgs, reply *ResultReply) error {
	//TODO: put the results in the global data structure etc...

	thisRange := ipTable[args.index]
	thisRange.lock.Lock()
	defer thisRange.lock.Unlock()

	//check if this node actually was registered with this range.
	rangeOwner := thisRange.currentProbe
	if rangeOwner != args.id {
		reply.ok = false
		ipTable[args.index].lock.Unlock()
		return os.ErrInvalid //TODO replace with custom error.
	}
	thisRange.currentProbe = "" //no id associated here anymore

	thisRange.stops.UnionWith(args.NewGSS)
	allIPS.UnionWith(args.News)

	unlockPlease[args.index] <- true //request to unlock this set.
	return nil
}

func (*Leader) GetIPs(args IpArgs, reply *IpReply) error {
	probeId := args.probeId
	
	//TODO: reserve a section of IPS, put it in reply, spawn lock process, etc
	return nil
}


