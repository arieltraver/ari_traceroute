//reference: https://www.davidkaya.com/Sets-in-golang/
package set
import (
	"sync"
	"strings"
	"strconv"
	"fmt"
)

type Set interface {
	Add(any)
	Remove(any)
	Contains(any) bool
	ToCSV() string
	Wipe()
	UnionWith(Set)
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

func (s *IntSet) Add(i any) {
	s.Mp[i.(int)] = struct{}{}
}

func (s *IntSet) Remove(i any) {
	delete(s.Mp, i.(int))
}

func (s *IntSet) Wipe() {
	m := make(map[int]struct{})
	s.Mp = m
}

func (s1 *IntSet) UnionWith(s2 Set) {
	for key, _ := range(s2.(*IntSet).Mp) {
		s1.Mp[key] = struct{}{}
	}
}

func (s *IntSet) Contains(i any) bool {
	_, ok := s.Mp[i.(int)]
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
func (s *StringSet) Contains(item any) bool {
	_, ok := s.Mp[item.(string)]
	return ok
}

func (s *StringSet) Wipe() {
	m := make(map[string]struct{})
	s.Mp = m
}

func (s *StringSet) Add(item any) {
	s.Mp[item.(string)] = struct{}{}
}

func (s *StringSet) Remove(item any) {
	delete(s.Mp, item.(string))
}

//expands the Set into its union with another Set
func (s1 *StringSet) UnionWith(s2 Set) {
	for key, _ := range(s2.(*StringSet).Mp) {
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

func NewSafeStringSet() *SafeSet {
	strset := NewStringSet()
	ss := &SafeSet{st:strset}
	return ss
}

func NewSafeIntSet() *SafeSet {
	intSet := NewIntSet()
	ss := &SafeSet{st:intSet}
	return ss
}

func (s *SafeSet) Add(item any) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.st.Add(item)
}

func (s *SafeSet) Remove(item any) {
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


func (s *SafeSet) Set() Set {
	s.lock.Lock()
	defer s.lock.Unlock()
	st := s.st
	return st
}

//replaces the set with a new eMpty set
func (s *SafeSet) Wipe() {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.st.Wipe()
}

func TestNoRoutine() {
	strs := NewStringSet()
	strs.Add("ab")
	strs.Add("cd")
	strs.Remove("cd")
	fmt.Println(strs.ToCSV())
	fmt.Println("-- should be: ab\n")
	ints := NewIntSet()
	ints.Add(1)
	ints.Add(2)
	ints.Remove(1)
	fmt.Println(ints.ToCSV())
	fmt.Println("-- should be:", 2, "\n")
	safeStrs := NewSafeStringSet()
	safeInts := NewSafeIntSet()
	safeStrs.Add("lame")
	safeStrs.Add("cool")
	safeStrs.Remove("lame")
	fmt.Println(safeStrs.ToCSV())
	fmt.Println("-- should be: cool\n")
	safeInts.Add(11)
	safeInts.Add(22)
	safeInts.Remove(22)
	fmt.Println(safeInts.ToCSV())
	fmt.Println("-- should be: 11\n")

	safeStrs.UnionWith(strs)
	fmt.Println(safeStrs.ToCSV())
	fmt.Println("-- should be: ab, cool or cool, ab")
}

func TestRoutines() {
	//Does a bunch of add and delete operations.
	//To be tested with data race checker.
	ss := NewSafeStringSet()
	wg := &sync.WaitGroup{}
	wg.Add(3)
	go func() {
		defer wg.Done()
		wg2 := &sync.WaitGroup{}
		wg2.Add(100)
		for i := 0; i < 100; i++ {
			go func(num int){
				ss.Add(strconv.Itoa(num))
				wg2.Done()
			}(i);
		}
		wg2.Wait()
	}();
	go func() {
		defer wg.Done()
		wg3 := &sync.WaitGroup{}
		wg3.Add(100)
		for i := 0; i < 100; i++ {
			go func(num int){
				ss.Remove(strconv.Itoa(num))
				wg3.Done()
			}(i);
		}
		wg3.Wait()
	}();
	go func() {
		defer wg.Done()
		strSet := NewStringSet();
		strSet.Add("WOAH NICE")
		strSet.Add("UNION")
		ss.UnionWith(strSet)
	}()
	wg.Wait()
	fmt.Println(ss.ToCSV())
}