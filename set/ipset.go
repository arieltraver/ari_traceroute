//SPECIAL DATA STRUCTURE: the IP set
//SIZE: 32x32
//subnet slices are smaller.

package set

import (
	"sync"
	"errors"
	"strings"
	"fmt"
	"strconv"
)

// IPSet type cotains library IPSet and Hasher function
type IPSet struct {
	BS []bool
}

func bytesToUint32(byts []byte) uint32 {
	if(len(byts)) == 1 {
		return uint32(byts[0])
	}
	var n uint32 = 0;
	for i := len(byts)-1; i>0; i-- {
		n += uint32(byts[i])
		n = n << 8
	}
	n+= uint32(byts[0])
	return n
}

func uint32ToBytes(n uint32) []byte{
	fmt.Println("n is", n)
	bytez := make([]byte, 4)
	mask := uint32(255)
	for i := 0; i < 4; i++ {
		n2 := n & mask //last 8 bits
		fmt.Println("n2 is", n2)
		bytez[i] = uint8(n2);
		n = n >> 8 //shift it down
		fmt.Println("n is", n, "after shift")
	}
	fmt.Println(bytez)
	return bytez
}

func uint32ToString(n uint32) string {
	bytez := uint32ToBytes(n)
	s := &strings.Builder{}
	for _, b := range(bytez[:3]) {
		s.WriteString(strconv.Itoa(int(b)))
		s.WriteString(".")
	}
	s.WriteString(strconv.Itoa(int(bytez[3])))
	return s.String()
}

// NewIPSet constructor for IPSet with an array of m bits
func NewIPSet(numBits uint) *IPSet {
	bs := make([]bool, uint32(1) << numBits-1)
	return &IPSet{BS:bs}
}

func NewIPv4Set() *IPSet {
	return NewIPSet(32)
}

// Add IP to IPSet based on itself
func (BS *IPSet) Add(elem []byte) {
	var i = bytesToUint32(elem)
	fmt.Println("index is",i);
	BS.BS[i] = true
}

func (BS *IPSet) REmove(elem []byte) {
	var i = bytesToUint32(elem)
	BS.BS[i] = false
}

// Check element in IPSet
func (BS *IPSet) Check(elem []byte) bool {
	var i = bytesToUint32(elem)
	return BS.BS[i]
}



// Union of two IPSets
func (BS *IPSet) Union(BS2 *IPSet) (error) {
	if len(BS.BS) != len(BS2.BS) {
		return errors.New("need same size for union")
	}
	for i, _ := range(BS.BS) {
		BS.BS[i] = BS.BS[i] && BS2.BS[i]
	}
	return nil
}

func (BS *IPSet) getCount() uint32 {
	var count uint32
	for _, val := range(BS.BS) {
		if val {
			count+=1
		}
	}
	return count
}

func (s *Set) ToIPSet() *IPSet{
	b := NewIPv4Set() //TODO: research the right size for this.
	mp := s.Mp //TODO make set iterable
	for item := range(mp) {
		b.Add([]byte(item))
	}
	return b
}


//following section by Ari:
type SafeIPSet struct {
	IPS *IPSet
	lock sync.Mutex
}


func NewSafeIPSet(m uint) *SafeIPSet {
	bitst := &SafeIPSet{IPS:NewIPSet(m), lock:sync.Mutex{}}
	return bitst
}

// Union of safe IPSet and regular IPSet
func (sBs *SafeIPSet) Union(that interface{}) (uint32, error) {
	sBs.lock.Lock()
	defer sBs.lock.Unlock()
	other, ok := that.(*IPSet)
	if !ok {
		return sBs.IPS.getCount(), errors.New("not an ip set")
	}
	sBs.IPS.Union(other)
	return sBs.IPS.getCount(), nil
}

func (sBs *SafeIPSet) Wipe(newsize uint) {
	sBs.lock.Lock()
	defer sBs.lock.Unlock()
	sBs.IPS = NewIPSet(newsize)
}

func (sBs *SafeIPSet) AddString(s string) {
	sBs.lock.Lock()
	defer sBs.lock.Unlock()
	sBs.IPS.Add([]byte(s))
}

func (sBs *SafeIPSet) CheckString(s string) bool {
	sBs.lock.Lock()
	defer sBs.lock.Unlock()
	return sBs.IPS.Check([]byte(s))
}

func (IPS *IPSet) ToCSV() string {
	s := &strings.Builder{}
	for i, val := range(IPS.BS) {
		if val {
			s.WriteString(uint32ToString(uint32(i)))
			s.WriteString(",")
		}
	}
	return s.String()
}

func (SPS *SafeIPSet) ToCSV() string {
	SPS.lock.Lock();
	defer SPS.lock.Unlock();
	return SPS.IPS.ToCSV();
}