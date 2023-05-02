//TAKEN FROM: https://github.com/krakendio/bloomfilter/tree/master/bitset
//Needed to take this because their bitset only uses MD5
//Doubletree paper uses Sha-1

// Package bitset implements a bitset based on the bitset library.
//
// It aggregates a hashfunction and implements the Add, Check and Union methods.
package set

import (
	"github.com/krakendio/bloomfilter/v2"
	"github.com/tmthrgd/go-bitset"
	"sync"
)

// BitSet type cotains library bitset and Hasher function
type BitSet struct {
	Bs     bitset.Bitset
	Hasher bloomfilter.Hash
}

// NewBitSet constructor for BitSet with an array of m bits
func NewBitSet(m uint) *BitSet {
	return &BitSet{Bs: bitset.New(m), Hasher: bloomfilter.SHA1} //this edit from Ari
}

// Add element to bitset
func (Bs *BitSet) Add(elem []byte) {
	Bs.Bs.Set(Bs.Hasher(elem)[0] % Bs.Bs.Len())
}

// Check element in bitset
func (Bs *BitSet) Check(elem []byte) bool {
	return Bs.Bs.IsSet(Bs.Hasher(elem)[0] % Bs.Bs.Len())
}

// Union of two bitsets
func (Bs *BitSet) Union(that interface{}) (float64, error) {
	other, ok := that.(*BitSet)
	if !ok {
		return Bs.getCount(), bloomfilter.ErrImpossibleToTreat
	}

	Bs.Bs.Union(Bs.Bs, other.Bs)
	return Bs.getCount(), nil
}

func (Bs *BitSet) getCount() float64 {
	return float64(Bs.Bs.Count()) / float64(Bs.Bs.Len())
}

func (s *Set) ToBitset() *BitSet{
	//size chosen using this: https://hur.st/bloomfilter/?n=4000&p=1.0E-7&m=&k=
	b := NewBitSet(134191) //TODO: research the right size for this.
	
	mp := s.Mp //TODO make set iterable
	for item := range(mp) {
		b.Add([]byte(item))
	}
	return b
}


//following section by Ari:
type SafeBitSet struct {
	bits *BitSet
	lock sync.Mutex
}


func NewSafeBitSet(m uint) *SafeBitSet {
	bitst := &SafeBitSet{bits:NewBitSet(m), lock:sync.Mutex{}}
	return bitst
}

// Union of safe bitset and regular bitset
func (sBs *SafeBitSet) Union(that interface{}) (float64, error) {
	sBs.lock.Lock()
	defer sBs.lock.Unlock()
	other, ok := that.(*BitSet)
	if !ok {
		return sBs.bits.getCount(), bloomfilter.ErrImpossibleToTreat
	}
	sBs.bits.Union(other)
	return sBs.bits.getCount(), nil
}

func (sBs *SafeBitSet) Wipe(newsize uint) {
	sBs.lock.Lock()
	defer sBs.lock.Unlock()
	sBs.bits = NewBitSet(newsize)
}

func (sBs *SafeBitSet) AddString(s string) {
	sBs.lock.Lock()
	defer sBs.lock.Unlock()
	sBs.bits.Add([]byte(s))
}

func (sBs *SafeBitSet) CheckString(s string) bool {
	sBs.lock.Lock()
	defer sBs.lock.Unlock()
	return sBs.bits.Check([]byte(s))
}

func (sBs *SafeBitSet) ToCSV() string {
	return "...implement toCSV"
}