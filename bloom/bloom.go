package bloom

import (
	"github.com/arieltraver/ari_traceroute/set"
)

//some helper functions to work with go bloom filters
//makes data transfer across the web easier


//TODO:
//consider. reworking everything such that GSS is always a bitset
//

func setToBloom(s *set.Set) *set.BitSet{
	//size chosen using this: https://hur.st/bloomfilter/?n=4000&p=1.0E-7&m=&k=
	b := set.NewBitSet(134191) //TODO: research the right size for this.
	
	mp := s.Mp //TODO make set iterable
	for item := range(mp) {
		b.Add([]byte(item))
	}
	return b
}