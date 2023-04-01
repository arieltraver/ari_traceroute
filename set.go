//reference: https://www.davidkaya.com/sets-in-golang/
package main
import (
	"sync"
	"fmt"
	"strings"
)


type set struct {
	//using struct{} because an empty struct takes up 0 bytes
	mp map[string]struct{}
}

var exists = struct{}{}

func NewSet() *set {
	s := &set{}
	s.mp = make(map[string]struct{})
	return s
}

/*unlock the set*/
func (s *set) Contains(item string) bool {
	_, ok := s[item]
	return ok
}

func (s *set) Add(item string) {
	s.mp[item] = exists
}

func (s *set) Remove(item string) {
	delete(s.mp, item)
}

//expands the set into its union with another set
func (s1 *set) UnionWith(s2 *set) {
	for key, _ := range(s2.mp) {
		s1.mp[key] = exists
	}
}

//reduces the set to its intersection with another set
func (s1 *set) IntersectWith(s2 *set) {
	for key, _ := range(s2.mp) {
		if !(s1.Contains(key)) {
			s1.Remove(key)
		}
	}
}

//returns a new set which is the union of two sets
func Union(s1 *set, s2 *set) *set {
	s3 := NewSet()
	for key, _ := range(s1.mp) {
		s3.mp[key] = exists
	}
	for key, _ := range(s2.mp) {
		s3.mp[key] = exists
	}
	return s3
}

func Intersection(s1 *set, s2 *set) *set {
	s3 := NewSet()
	for key, _ := range(s1.mp) {
		if s2.Contains(key) {
			s3.mp[key] = exists
		}
	}
	return s3
}

//returns a new set that is s1 U s2'
func IntersectionComplement(s1 *set, s2 *set) *set {
	s3 := NewSet()
	for key, _ := range(s1.mp) {
		if !s2.Contains(key) {
			s3.mp[key] = exists
		}
	}
	return s3
}

//turns a set into a CSV
func (s *set) ToCSV() string {
	str := &strings.Builder{}
	for item, _ := range(s.mp) {
		str.WriteString(item + ",")
	}
	str.WriteRune('\n')
	return str.String()
}

type safeSet struct {
	st *set
	lock sync.Mutex
}

func NewSafeSet() *safeSet {
	st := NewSet()
	sf := &safeSet{st:st}
	return sf
}

func (s *safeSet) Add(item string) {
	s.lock.Lock()
	s.st.Add(item)
	s.lock.Unlock()
}

func (s *safeSet) Remove(item string) {
	s.lock.Lock()
	s.st.Remove(item)
	s.lock.Unlock()
}

func (s *safeSet) Contains(item string) bool {
	s.lock.Lock()
	t := s.st.Contains(item)
	s.lock.Unlock()
	return t
}

func (s1 *safeSet) UnionWith(s2 *safeSet) {
	s1.lock.Lock()
	s2.lock.Lock()
	s1.st.UnionWith(s2.st)
	s1.lock.Unlock()
	s2.lock.Unlock()
}

func (s1 *safeSet) IntersectWith(s2 *safeSet) {
	s1.lock.Lock()
	s2.lock.Lock()
	s1.st.IntersectWith(s2.st)
	s1.lock.Unlock()
	s2.lock.Unlock()
}

func SafeUnion(s1 *safeSet, s2 *safeSet) *safeSet {
	s3 := NewSafeSet()
	s1.lock.Lock()
	s2.lock.Lock()
	s3.lock.Lock()
	s3.st = Union(s1.st, s2.st)
	s1.lock.Unlock()
	s2.lock.Unlock()
	s3.lock.Unlock()
	return s3
}

func (s *safeSet) ToCSV() string {
	s.lock.Lock()
	str := s.st.ToCSV()
	s.lock.Unlock()
	return str
}

func setTests() {
	fruits := NewSafeSet()
	vegetables := NewSafeSet()
	fruits.Add("apple")
	fruits.Add("banana")
	fruits.Add("tomato")

	//contains test
	if (fruits.Contains("banana")) {
		fmt.Println("banana is a fruit")
	} else {
		fmt.Println("why isn't banana a fruit?")
	}

	//remove null test
	fruits.Remove("onion") //nothing should happen

	vegetables.Add("lettuce")
	vegetables.Add("carrot")
	vegetables.Add("tomato")
	vegetables.Add("pizza")

	//remove non null test
	vegetables.Remove("pizza")
	if (!(vegetables.Contains("pizza"))){
		fmt.Println("pizza is not a vegetable")
	} else {
		fmt.Println("pizza is a vegetable???")
	}

	//union (new set) test
	ediblePlants := SafeUnion(vegetables, fruits)
	fmt.Print(ediblePlants.ToCSV())

	//intersection (in place) test
	vegetables.IntersectWith(fruits)
	fmt.Println("this is both a fruit and a vegetable:")
	fmt.Print(vegetables.ToCSV()) //tomato, 

	//union (in place) test
	ediblePlants.Remove("tomato")
	ediblePlants.UnionWith(fruits) //bring back tomato
	fmt.Print(ediblePlants.ToCSV()) //apple banana lettuce carrot tomato
}