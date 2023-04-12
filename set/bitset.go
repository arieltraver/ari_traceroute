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

// BitSet type cotains library bitset and hasher function
type BitSet struct {
	bs     bitset.Bitset
	hasher bloomfilter.Hash
}

// NewBitSet constructor for BitSet with an array of m bits
func NewBitSet(m uint) *BitSet {
	return &BitSet{bs: bitset.New(m), hasher: bloomfilter.SHA1} //this edit from Ari
}

// Add element to bitset
func (bs *BitSet) Add(elem []byte) {
	bs.bs.Set(bs.hasher(elem)[0] % bs.bs.Len())
}

// Check element in bitset
func (bs *BitSet) Check(elem []byte) bool {
	return bs.bs.IsSet(bs.hasher(elem)[0] % bs.bs.Len())
}

// Union of two bitsets
func (bs *BitSet) Union(that interface{}) (float64, error) {
	other, ok := that.(*BitSet)
	if !ok {
		return bs.getCount(), bloomfilter.ErrImpossibleToTreat
	}

	bs.bs.Union(bs.bs, other.bs)
	return bs.getCount(), nil
}

func (bs *BitSet) getCount() float64 {
	return float64(bs.bs.Count()) / float64(bs.bs.Len())
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
	bs *BitSet
	lock sync.Mutex
}

func NewSafeBitSet(m uint) *SafeBitSet {
	bitst := &SafeBitSet{bs:NewBitSet(m), lock:sync.Mutex{}}
	return bitst
}