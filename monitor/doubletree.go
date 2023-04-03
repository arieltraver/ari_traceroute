package main
import (
	"github.com/aeden/traceroute"
	"github.com/arieltraver/ari_traceroute/set"
	"math/rand"
	"time"
	"syscall"
	"net"
	"errors"
	"log"
	"sync"
)

var RANDOMIZED bool = true
var PORT = 50000 //replace later with something to find open ports
var TIMEOUT int64 = 1
var PACKETSIZE int = 10

func setMaxProbe(max int) int {
	if RANDOMIZED {
		s1 := rand.NewSource(time.Now().UnixNano())
		r1 := rand.New(s1)
		i := r1.Intn(max)
		return i
	}
	return max
}

func setUpSockets() (int, int, error) {
	//select non loopback address
	socketAdd, err := socketAddr()
	if err != nil {
		return -1, -1, err
	}
	// Set up the socket to receive inbound packets
	recvSocket, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_ICMP)
	if err != nil {
		return -1, -1, err
	}

	// Set up the socket to send packets out.
	sendSocket, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, syscall.IPPROTO_UDP)
	if err != nil {
		return -1, -1, err
	}
	//bind the socket to listen for ICMP
	syscall.Bind(recvSocket, &syscall.SockaddrInet4{Port: PORT, Addr: socketAdd})

	return recvSocket, sendSocket, nil
}


func sendProbes(GSS *set.SafeSet, LSS *set.SafeSet, ips []string) {
	NewNodes := set.NewSafeSet()
	source, err := socketAddr()
	if err != nil {
		log.Fatal(err)
	}
	var wg sync.WaitGroup
	sendSock, recSock, err := setUpSockets() //set up sockets for sending and listening
	//do we need a new port/socket for each probe? probably since we don't want to mix things
	if err != nil 
		log.Fatal(err)
	}
	wg.Add(len(ips)) //one thread per IP
	for _, ip := range(ips){
		go probeAddr(&wg, &NewNodes, GSS, LSS, sendSock, recSock, ip, source)
	}
}

func probeAddr(wg *sync.WaitGroup, NewNodes *set.SafeSet, GSS *set.SafeSet, LSS *set.SafeSet, sendSock int, recSock int, ip string, source [4]byte) {
	defer wg.Done()
	probeForward(source, sendSock, recSock, ip, LSS, GSS)
	probeBack(source, sendSock, recSock, ip, LSS, GSS)
}

/*
func probe(source...)
-randomly (or not) determine hopcount
-lasthop = source
-for i = 1; i <= hopcount; i++ {
	code, hop := probeForward(...)
	save hop information
	seenLinks[lasthop.address + hop.address] = true //sort addresses to make unique
	lasthop = hop
	if code == 1: //node in GSS
		register this ended the probe?
		hopcount = i
		break;
	else if code == 2: //node is destination
		hopcount = i
		register destination reached ended probe
		break;
	else: //more to go
		seenNodes[hop.address] = true //add this node to graph
}
-for i = (new)hopcount; i >= 1; i-- {
	code, hop := probeBackward(...)
}
*/

//returns -1 if it encounters an error, 0 if the node is new, 1 if it hits a global stop, 2 if it hits dest
func probeForward(source [4] byte, sendSock int, recSock int, dest string, port int, GSS *set.SafeSet, LSS *set.SafeSet) (int, *traceroute.TracerouteHop) {
	destAdd, err := destAddr(dest)
	if err != nil {
		log.Println(err)
		return -1, nil
	}
	//convert timeout
	tv := syscall.NsecToTimeval(1000 * 1000 * TIMEOUT)
	//set up time interval to wait for response
	syscall.SetsockoptTimeval(recSock, syscall.SOL_SOCKET, syscall.SO_RCVTIMEO, &tv)

	start := time.Now()
	//send an emtpy UDP packet to the destination
	syscall.Sendto(sendSock, []byte{0x0}, 0, &syscall.SockaddrInet4{Port: port, Addr: destAdd})

	//receive a response
	var p = make([]byte, PACKETSIZE)
	n, from, err := syscall.Recvfrom(recSock, p, 0)
	elapsed := time.Since(start)
	if err != nil {
		log.Println(err)
		return -1, nil
	} else {
		//save the result in an object
		addr := from.(*syscall.SockaddrInet4).Addr
		hop := traceroute.TracerouteHop{Success: true, Address: addr, N: n, ElapsedTime: elapsed, TTL: ttl}
		//DNS lookup of the IP
		currHost, err := net.LookupAddr(hop.AddressString())
		if err == nil {
			hop.Host = currHost[0]
		}
		if addr == source {
			return 2, &hop //reached destination
		}
		hopDest := string(destAdd[:]) + string(addr[:]) //for hashing purposes
		if !GSS.Contains(hopDest) {
			GSS.Add(hopDest)
			return 0, &hop //new node and not yet at destination
		} else {
			return 1, &hop //node already in GSS
		}
	}
}

func probeBack(source [4]byte, sendSock int, recSock int, dest string, ttl int, timeout int64, port int, packetSize int, GSS map[string]bool, LSS map[string]bool) (int, *traceroute.TracerouteHop) {
	destAdd, err := destAddr(dest)
	if err != nil {
		log.Println(err)
		return -1, nil
	}
	//convert timeout
	tv := syscall.NsecToTimeval(1000 * 1000 * timeout)
	//set up time interval to wait for response
	syscall.SetsockoptTimeval(recSock, syscall.SOL_SOCKET, syscall.SO_RCVTIMEO, &tv)

	start := time.Now()
	//send an emtpy UDP packet to the destination
	syscall.Sendto(sendSock, []byte{0x0}, 0, &syscall.SockaddrInet4{Port: port, Addr: destAdd})

	//receive a response
	var p = make([]byte, packetSize)
	n, from, err := syscall.Recvfrom(recSock, p, 0)
	elapsed := time.Since(start)
	if err != nil {
		log.Println(err)
		return -1, nil
	} else {
		//save the result in an object
		addr := from.(*syscall.SockaddrInet4).Addr
		hop := traceroute.TracerouteHop{Success: true, Address: addr, N: n, ElapsedTime: elapsed, TTL: ttl}
		//DNS lookup of the IP
		currHost, err := net.LookupAddr(hop.AddressString())
		if err == nil {
			hop.Host = currHost[0]
		}
		if addr == source {
			return 2, &hop //reached destination
		}
		hopSource := string(destAdd[:]) + string(source[:]) //for hashing purposes
		LSS[hopSource] = true //place in local stop set
		GSS[hopSource] = true //place in global stop set
		return 0, &hop
	}
}

// Given a host name convert it to a 4 byte IP address.
func destAddr(dest string) (destAddr [4]byte, err error) {
	addrs, err := net.LookupHost(dest)
	if err != nil {
		return
	}
	addr := addrs[0]

	ipAddr, err := net.ResolveIPAddr("ip", addr)
	if err != nil {
		return
	}
	copy(destAddr[:], ipAddr.IP.To4())
	return
}

//from original traceroute. selects a non-loopback prefix to be used as source IP
func socketAddr() (addr [4]byte, err error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return
	}
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if len(ipnet.IP.To4()) == net.IPv4len {
				copy(addr[:], ipnet.IP.To4())
				return
			}
		}
	}
	err = errors.New("you do not appear to be connected to the internet")
	return
}