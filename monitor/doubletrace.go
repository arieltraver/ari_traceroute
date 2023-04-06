package monitor

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net"
	"sync"
	"syscall"
	"time"
	"github.com/arieltraver/ari_traceroute/set"
)

const DEFAULT_PORT int = 33434
const DEFAULT_MAX_HOPS = 64
const DEFAULT_FIRST_HOP = 1
const DEFAULT_TIMEOUT_MS = 500
const DEFAULT_RETRIES = 3
const DEFAULT_PACKET_SIZE = 52
const FLOOR = 6
const CEILING = 12

type Monitor struct{
	GSS *set.SafeSet
}

//doubletree addon from paper, helps prevent overburdening destinations
func (options *TracerouteOptions) SetMaxHopsRandom(floor int, ceiling int) {
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	i := r1.Intn(ceiling - floor)
	i += floor
	options.maxHops = i
}

//setter
func (options *TracerouteOptions) SetMaxHops(maxHops int) {
	options.maxHops = maxHops
}


// Return the first non-loopback address as a 4 byte IP address. This address
// is used for sending packets out.
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
	err = errors.New("you do not appear to be connected to the Internet")
	return
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

// TracrouteOptions type
type TracerouteOptions struct {
	port       int
	maxHops    int
	firstHop   int
	timeoutMs  int
	retries    int
	packetSize int
}

func (options *TracerouteOptions) Port() int {
	if options.port == 0 {
		options.port = DEFAULT_PORT
	}
	return options.port
}

func (options *TracerouteOptions) SetPort(port int) {
	options.port = port
}

func (options *TracerouteOptions) MaxHops() int {
	if options.maxHops == 0 {
		options.SetMaxHopsRandom(5, 10)
	}
	return options.maxHops
}

func (options *TracerouteOptions) FirstHop() int {
	if options.firstHop == 0 {
		options.firstHop = DEFAULT_FIRST_HOP
	}
	return options.firstHop
}

func (options *TracerouteOptions) SetFirstHop(firstHop int) {
	options.firstHop = firstHop
}

func (options *TracerouteOptions) TimeoutMs() int {
	if options.timeoutMs == 0 {
		options.timeoutMs = DEFAULT_TIMEOUT_MS
	}
	return options.timeoutMs
}

func (options *TracerouteOptions) SetTimeoutMs(timeoutMs int) {
	options.timeoutMs = timeoutMs
}

func (options *TracerouteOptions) Retries() int {
	if options.retries == 0 {
		options.retries = DEFAULT_RETRIES
	}
	return options.retries
}

func (options *TracerouteOptions) SetRetries(retries int) {
	options.retries = retries
}

func (options *TracerouteOptions) PacketSize() int {
	if options.packetSize == 0 {
		options.packetSize = DEFAULT_PACKET_SIZE
	}
	return options.packetSize
}

func (options *TracerouteOptions) SetPacketSize(packetSize int) {
	options.packetSize = packetSize
}

// TracerouteHop type
type TracerouteHop struct {
	Success     bool
	Address     [4]byte
	Host        string
	N           int
	ElapsedTime time.Duration
	TTL         int
}

func addressString(add [4]byte) string {
	return fmt.Sprintf("%v.%v.%v.%v", add[0], add[1], add[2], add[3])
}

func (hop *TracerouteHop) AddressString() string {
	return fmt.Sprintf("%v.%v.%v.%v", hop.Address[0], hop.Address[1], hop.Address[2], hop.Address[3])
}

func (hop *TracerouteHop) HostOrAddressString() string {
	hostOrAddr := hop.AddressString()
	if hop.Host != "" {
		hostOrAddr = hop.Host
	}
	return hostOrAddr
}

// TracerouteResult type
type TracerouteResult struct {
	DestinationAddress [4]byte
	Hops               []TracerouteHop
}

func notify(hop TracerouteHop, channels []chan TracerouteHop) {
	for _, c := range channels {
		c <- hop
	}
}

func closeNotify(channels []chan TracerouteHop) {
	for _, c := range channels {
		close(c)
	}
}

func sendProbes(GSS *safeSet, ips []string) {
	NewNodes := NewSafeSet()
	LSS := NewSafeSet()
	var wg sync.WaitGroup
	wg.Add(len(ips)) //one thread per IP
	for _, ip := range(ips){
		fmt.Println("probing", ip)
		go probeAddr(&wg, NewNodes, GSS, LSS, ip)
	}
	wg.Wait()
}

func probeAddr(wg *sync.WaitGroup, NewNodes *safeSet, GSS *safeSet, LSS *safeSet, ip string) {
	defer wg.Done()
	options := &TracerouteOptions{}
	options.SetMaxHopsRandom(FLOOR, CEILING)
	sourceAddr, err := socketAddr() //possible cause of glitch
	if err != nil {
		log.Fatal(err) //Todo: replace with non fatal err & return
	}
	forward := make(chan TracerouteHop, options.maxHops)
	forwardHops, err := probeForward(sourceAddr, GSS, ip, options, forward)
	if err != nil {
		log.Fatal(err)
	}
	backward := make(chan TracerouteHop, options.maxHops)
	//
	_, err = probeBackwards(sourceAddr,forwardHops.Hops, GSS, LSS, options, backward)
	if err != nil {
		log.Fatal(err) //TODO: do not crash the whole program if one trace fails.
	}

	//TODO: check for null nodes.
	//add all new nodes to the set
	for _, hop := range(forwardHops.Hops) {
		NewNodes.Add(hop.AddressString())
	}
	//TODO: add new (sorted) ip,ip edges to the edge set.
}





// Traceroute uses the given dest (hostname) and options to execute a traceroute
// from your machine to the remote host.
//
// Outbound packets are UDP packets and inbound packets are ICMP.
//
// Returns a TracerouteResult which contains an array of hops. Each hop includes
// the elapsed time and its IP address.
func probeForward(socketAddr [4]byte, GSS *safeSet, dest string, options *TracerouteOptions, c ...chan TracerouteHop) (result TracerouteResult, err error) {
	fmt.Println("probe forward")
	result.Hops = make([]TracerouteHop, 0, options.maxHops) //prevent resizing
	destAddr, err := destAddr(dest)
	result.DestinationAddress = destAddr
	if err != nil {
		return
	}

	timeoutMs := (int64)(options.TimeoutMs())
	tv := syscall.NsecToTimeval(1000 * 1000 * timeoutMs)

	ttl := 0
	retry := 0
	for {

		ttl += 1
		//log.Println("TTL: ", ttl)
		start := time.Now()

		// Set up the socket to receive inbound packets
		recvSocket, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_ICMP)
		if err != nil {
			log.Fatal(err)
		}

		// Set up the socket to send packets out.
		sendSocket, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, syscall.IPPROTO_UDP)
		if err != nil {
			log.Fatal(err)
		}

		/*
		THIS IS WHERE PARIS TRACEROUTE MODIFICATIONS CAN BE MADE
		using: syscall.Setsockopt
		*/

		// This sets the current hop TTL
		syscall.SetsockoptInt(sendSocket, 0x0, syscall.IP_TTL, ttl)
		// This sets the timeout to wait for a response from the remote host
		syscall.SetsockoptTimeval(recvSocket, syscall.SOL_SOCKET, syscall.SO_RCVTIMEO, &tv)

		defer syscall.Close(recvSocket)
		defer syscall.Close(sendSocket)

		// Bind to the local socket to listen for ICMP packets
		syscall.Bind(recvSocket, &syscall.SockaddrInet4{Port: options.Port(), Addr: socketAddr})

		/*
		 :In UDP probes, it is the checksum field. This requires manipulating the payload
		to yield the desired checksum, as packets with an incorrect checksum are liable
		to be discarded."
		GOAL: replace "[]byte{0x0}" with a modified payload that keeps the checksum constant
			  print out the checksum each time
		*/
		// Send a single null byte UDP packet
		syscall.Sendto(sendSocket, []byte{0x0}, 0, &syscall.SockaddrInet4{Port: options.Port(), Addr: destAddr})

		var p = make([]byte, options.PacketSize())
		n, from, err := syscall.Recvfrom(recvSocket, p, 0)
		elapsed := time.Since(start)
		if err == nil {
			currAddr := from.(*syscall.SockaddrInet4).Addr

			hop := TracerouteHop{Success: true, Address: currAddr, N: n, ElapsedTime: elapsed, TTL: ttl}

			// TODO: this reverse lookup appears to have some standard timeout that is relatively
			// high. Consider switching to something where there is greater control.
			currHost, err := net.LookupAddr(hop.AddressString())
			if err == nil {
				hop.Host = currHost[0]
			}

			notify(hop, c)
			
			retry = 0

			hopDestString := hop.AddressString() + "-" + addressString(destAddr)

			// modification added here to stop if it hits node in GSS or LSS
			if ttl > options.MaxHops() || currAddr == destAddr || GSS.Contains(hopDestString) {
				closeNotify(c)
				return result, nil
			}
			result.Hops = append(result.Hops, hop)
			GSS.Add(hopDestString) //add to global stop set
		} else {
			retry += 1
			if retry > options.Retries() {
				notify(TracerouteHop{Success: false, TTL: ttl}, c)
				ttl += 1
				retry = 0
			}

			if ttl > options.MaxHops() {
				closeNotify(c)
				return result, nil
			}
		}

	}
}

/*
unlike forwards route discovery, backwards goes from probe to each hop.
this records routes between each hop and the probe, with the probe as destination.
each(hop, probe) address pair is added to both GSS and LSS.
*/
func probeBackwards(socketAddr [4]byte, forwardHops []TracerouteHop, LSS *safeSet, GSS *safeSet, options *TracerouteOptions, c ...chan TracerouteHop) (result TracerouteResult, err error) {
	source := addressString(socketAddr)
	result.Hops = make([]TracerouteHop, 0, len(forwardHops)) //prevent resizing

	timeoutMs := (int64)(options.TimeoutMs())
	tv := syscall.NsecToTimeval(1000 * 1000 * timeoutMs)

	retry := 0
	currentHop := len(forwardHops) - 1
	for {
		hopAddr := forwardHops[currentHop].Address //probe the address
		// Set up the socket to receive inbound packets
		recvSocket, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_ICMP)
		if err != nil {
			return result, err
		}

		// Set up the socket to send packets out.
		sendSocket, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, syscall.IPPROTO_UDP)
		if err != nil {
			return result, err
		}

		/*
		THIS IS WHERE PARIS TRACEROUTE MODIFICATIONS CAN BE MADE
		using: syscall.Setsockopt
		*/

		// set current hop ttl to die when it reaches destination
		syscall.SetsockoptInt(sendSocket, 0x0, syscall.IP_TTL, currentHop + 1)
		// This sets the timeout to wait for a response from the remote host
		syscall.SetsockoptTimeval(recvSocket, syscall.SOL_SOCKET, syscall.SO_RCVTIMEO, &tv)

		defer syscall.Close(recvSocket)
		defer syscall.Close(sendSocket)

		// Bind to the local socket to listen for ICMP packets
		syscall.Bind(recvSocket, &syscall.SockaddrInet4{Port: options.Port(), Addr: hopAddr})

		/*
		planned modification: keep checksum constant using payload?...
		 :In UDP probes, it is the checksum field. This requires manipulating the payload
		to yield the desired checksum, as packets with an incorrect checksum are liable
		to be discarded."
		GOAL: replace "[]byte{0x0}" with a modified payload that keeps the checksum constant
			  print out the checksum each time
		*/
		// Send a single null byte UDP packet
		start := time.Now()
		syscall.Sendto(sendSocket, []byte{0x0}, 0, &syscall.SockaddrInet4{Port: options.Port(), Addr: hopAddr})

		var p = make([]byte, options.PacketSize())
		n, from, err := syscall.Recvfrom(recvSocket, p, 0)
		elapsed := time.Since(start)
		if err == nil {
			currAddr := from.(*syscall.SockaddrInet4).Addr

			hop := TracerouteHop{Success: true, Address: currAddr, N: n, ElapsedTime: elapsed, TTL: currentHop + 1}

			// TODO: this reverse lookup appears to have some standard timeout that is relatively
			// high. Consider switching to something where there is greater control.
			currHost, err2 := net.LookupAddr(hop.AddressString())
			if err2 == nil {
				hop.Host = currHost[0]
			}

			notify(hop, c)

			result.Hops = append(result.Hops, hop)
			GSS.Add(addressString(hopAddr) +"-"+ source) //modification: add to GSS while probing back
			LSS.Add(addressString(hopAddr) +"-" + source) //add to LSS while probing back

			currentHop-=1
			retry = 0

			if currentHop <= 0 {
				closeNotify(c)
				return result, nil
			}
		} else {
			retry += 1
			if retry > options.Retries() {
				notify(TracerouteHop{Success: false, TTL: currentHop}, c)
				currentHop -= 1
				retry = 0
			}
		}

	}
}

func testJustProbes(addr string) {
	testGSS := NewSafeSet()
	testLSS := NewSafeSet()
	options := &TracerouteOptions{}
	options.SetMaxHopsRandom(FLOOR, CEILING)
	fmt.Println("max hops is", options.maxHops)
	sourceAddr, err := socketAddr()
	if err != nil {
		log.Fatal(err) //Todo: replace with non fatal err & return
	}
	hopChan := make(chan TracerouteHop, options.maxHops)
	forwardResult, err := probeForward(sourceAddr, testGSS, addr, options, hopChan)
	if err != nil {
		log.Fatal(err)
	}
	for _, hop := range(forwardResult.Hops) {
		fmt.Println(hop.AddressString())
	}
	fmt.Println("-----------------")
	backward := make(chan TracerouteHop, options.maxHops)
	backResult, err := probeBackwards(sourceAddr, forwardResult.Hops, testGSS, testLSS, options, backward)
	if err != nil {
		log.Fatal(err) //TODO: do not crash the whole program if one trace fails.
	}
	for _, hop := range(backResult.Hops) {
		fmt.Println(hop.AddressString())
	}

	fmt.Println("-------GSS-------")
	fmt.Print(testGSS.ToCSV())
	fmt.Println("-------LSS-------")
	fmt.Print(testGSS.ToCSV())

}

func testConcurrent() {
	GSS := NewSafeSet()
	ips := []string{"bugsincyberspace.com", "wellesley.edu", "google.com", "github.com"}
	sendProbes(GSS, ips)
	fmt.Println("-------GSS-------")
	fmt.Print(GSS.ToCSV())
}

func main(){
	//testJustProbes("bugsincyberspace.com")
	testConcurrent()
}