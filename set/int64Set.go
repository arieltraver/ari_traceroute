
package set

import (
	"sync"
	"errors"
	"strings"
	"fmt"
	"strconv"
)

type Uint64Set struct {
	chunks []uint64
	l uint64 //the length
}

func getBitFromAddr(addr uint64, l uint64) (uint64, uint64) {
	chunk := addr / (l * 64) //find which 64 bit chunk to store it in
	offset := addr % 64 //find which number
	var num uint64 = 1
	num = num << offset //get a bit in the right location, rest is 0s
	return chunk, num
}
/*here, the address (addr) can be the result of a hash function*/
func (uSet *Uint64Set) insertAtAddr(addr uint64) {
	chunk, num := getBitFromAddr(addr, uSet.l)
	uSet.chunks[chunk] |= num //bitwise OR
}

func (uSet *Uint64Set) checkAddr(addr uint64) bool {
	chunk, num := getBitFromAddr(addr, uSet.l)
	return ((uSet.chunks[chunk] ^ num) == 0)
}