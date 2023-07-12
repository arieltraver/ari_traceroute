//reference: https://www.davidkaya.com/Sets-in-golang/
package set
import (
	"sync"
	"strings"
	"strconv"
)

type Set interface {
	UnionWith(Set)
	Add(any)
	Remove(any)
	Contains(any) bool
	ToCSV() string
	Wipe()
}

//---------IntSet-------------------------------------------------//

type IntSet struct {
	Mp map[int]struct{}
}

func NewIntSet() *IntSet {
	s := &IntSet{}
	s.Mp = make(map[int]struct{})
	return s
}

func (s *IntSet) ToCSV() string {
	str := &strings.Builder{}
	for item, _ := range(s.Mp) {
		str.WriteString(strconv.Itoa(item) + ",")
	}
	str.WriteRune('\n')
	return str.String()
}

func (s *IntSet) Add(i int) {
	s.Mp[i] = struct{}{}
}

func (s *IntSet) Remove(i int) {
	delete(s.Mp, i)
}

func (s *IntSet) Wipe() {
	m := make(map[int]struct{})
	s.Mp = m
}

func (s1 *IntSet) UnionWith(s2 *IntSet) {
	for key, _ := range(s2.Mp) {
		s1.Mp[key] = struct{}{}
	}
}

func (s *IntSet) Contains(i int) bool {
	_, ok := s.Mp[i]
	return ok
}




//---------StringSet-------------------------------------------------//

type StringSet struct {
	//using struct{} because an eMpty struct takes up 0 bytes
	Mp map[string]struct{} //TODO: make this private and make an enumerate method
}

func NewStringSet() *StringSet {
	s := &StringSet{}
	s.Mp = make(map[string]struct{})
	return s
}


func (s *StringSet) Size() int {
	return len(s.Mp)
}

/*unlock the Set*/
func (s *StringSet) Contains(item string) bool {
	_, ok := s.Mp[item]
	return ok
}

func (s *StringSet) Wipe() {
	m := make(map[string]struct{})
	s.Mp = m
}

func (s *StringSet) Add(item string) {
	s.Mp[item] = struct{}{}
}

func (s *StringSet) Remove(item string) {
	delete(s.Mp, item)
}

//expands the Set into its union with another Set
func (s1 *StringSet) UnionWith(s2 *StringSet) {
	for key, _ := range(s2.Mp) {
		s1.Mp[key] = struct{}{}
	}
}

//reduces the Set to its intersection with another Set
func (s1 *StringSet) IntersectWith(s2 *StringSet) {
	for key, _ := range(s1.Mp) {
		if !(s2.Contains(key)) {
			s1.Remove(key)
		}
	}
}

//returns a new Set which is the union of two Sets
func Union(s1 *StringSet, s2 *StringSet) *StringSet {
	s3 := NewStringSet()
	for key, _ := range(s1.Mp) {
		s3.Mp[key] = struct{}{}
	}
	for key, _ := range(s2.Mp) {
		s3.Mp[key] = struct{}{}
	}
	return s3
}

func Intersection(s1 *StringSet, s2 *StringSet) *StringSet {
	s3 := NewStringSet()
	for key, _ := range(s1.Mp) {
		if s2.Contains(key) {
			s3.Mp[key] = struct{}{}
		}
	}
	return s3
}

//returns a new Set that is s1 U s2'
func IntersectionComplement(s1 *StringSet, s2 *StringSet) *StringSet {
	s3 := NewStringSet()
	for key, _ := range(s1.Mp) {
		if !s2.Contains(key) {
			s3.Mp[key] = struct{}{}
		}
	}
	return s3
}

//turns a Set into a CSV
func (s *StringSet) ToCSV() string {
	str := &strings.Builder{}
	for item, _ := range(s.Mp) {
		str.WriteString(item + ",")
	}
	str.WriteRune('\n')
	return str.String()
}




//---------SafeSet-------------------------------------------------//

type SafeSet struct {
	st Set
	lock sync.Mutex
}

func (s *SafeSet) Add(item string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.st.Add(item)
}

func (s *SafeSet) Remove(item string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.st.Remove(item)
}

func (s *SafeSet) Contains(item any) bool {
	s.lock.Lock()
	defer s.lock.Unlock()
	t := s.st.Contains(item)
	return t
}

func (s1 *SafeSet) UnionWith(s2 Set){
	s1.lock.Lock()
	defer s1.lock.Unlock()
	s1.st.UnionWith(s2)
}

func (s *SafeSet) ToCSV() string {
	s.lock.Lock()
	defer s.lock.Unlock()
	str := s.st.ToCSV()
	return str
}


func (s *SafeSet) Set() *Set {
	s.lock.Lock()
	defer s.lock.Unlock()
	st := s.st
	return &st
}

//replaces the set with a new eMpty set
func (s *SafeSet) Wipe() {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.st.Wipe()
}
/*
//tries every function, many concurrent adds/removes
func testRoutine(s1 *SafeSet, s2 *SafeSet, s3 *SafeSet, wg *sync.WaitGroup){
	defer wg.Done()
	for i := 0; i < 5; i++ {
		fmt.Print(s3.ToCSV())
		s1.Add(strconv.Itoa(i))
		s1.Remove(strconv.Itoa(5 - i))
		s1.UnionWith(s2)
		s2.IntersectWith(s1)
		s3 = SafeUnion(s1, s2) //is s3's old memory getting collected?...
		s3.Add("test")
	}
}
*/