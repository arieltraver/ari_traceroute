//reference: https://www.davidkaya.com/sets-in-golang/
package main
import (
	"sync"
	"fmt"
	"strings"
	"strconv"
)


type set struct {
	//using struct{} because an empty struct takes up 0 bytes
	mp map[string]struct{}
}

func NewSet() *set {
	s := &set{}
	s.mp = make(map[string]struct{})
	return s
}

/*unlock the set*/
func (s *set) Contains(item string) bool {
	_, ok := s.mp[item]
	return ok
}

func (s *set) Add(item string) {
	s.mp[item] = struct{}{}
}

func (s *set) Remove(item string) {
	delete(s.mp, item)
}

//expands the set into its union with another set
func (s1 *set) UnionWith(s2 *set) {
	for key, _ := range(s2.mp) {
		s1.mp[key] = struct{}{}
	}
}

//reduces the set to its intersection with another set
func (s1 *set) IntersectWith(s2 *set) {
	for key, _ := range(s1.mp) {
		if !(s2.Contains(key)) {
			s1.Remove(key)
		}
	}
}

//returns a new set which is the union of two sets
func Union(s1 *set, s2 *set) *set {
	s3 := NewSet()
	for key, _ := range(s1.mp) {
		s3.mp[key] = struct{}{}
	}
	for key, _ := range(s2.mp) {
		s3.mp[key] = struct{}{}
	}
	return s3
}

func Intersection(s1 *set, s2 *set) *set {
	s3 := NewSet()
	for key, _ := range(s1.mp) {
		if s2.Contains(key) {
			s3.mp[key] = struct{}{}
		}
	}
	return s3
}

//returns a new set that is s1 U s2'
func IntersectionComplement(s1 *set, s2 *set) *set {
	s3 := NewSet()
	for key, _ := range(s1.mp) {
		if !s2.Contains(key) {
			s3.mp[key] = struct{}{}
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
	var sf safeSet
	sf.st = NewSet()
	return &sf
}

func (s *safeSet) Add(item string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.st.Add(item)
}

func (s *safeSet) Remove(item string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.st.Remove(item)
}

func (s *safeSet) Contains(item string) bool {
	s.lock.Lock()
	defer s.lock.Unlock()
	t := s.st.Contains(item)
	return t
}

func (s1 *safeSet) UnionWith(s2 *safeSet) {
	s1.lock.Lock()
	defer s1.lock.Unlock()
	s2.lock.Lock()
	defer s2.lock.Unlock()
	s1.st.UnionWith(s2.st)
}

func (s1 *safeSet) IntersectWith(s2 *safeSet) {
	s1.lock.Lock()
	defer s1.lock.Unlock()
	s2.lock.Lock()
	defer s2.lock.Unlock()
	s1.st.IntersectWith(s2.st)
}

func SafeUnion(s1 *safeSet, s2 *safeSet) *safeSet {
	s3 := NewSafeSet()
	s1.lock.Lock()
	defer s1.lock.Unlock()
	s2.lock.Lock()
	defer s2.lock.Unlock()
	s3.lock.Lock()
	defer s3.lock.Unlock()
	s3.st = Union(s1.st, s2.st)
	return s3
}

func (s *safeSet) ToCSV() string {
	s.lock.Lock()
	defer s.lock.Unlock()
	str := s.st.ToCSV()
	return str
}

func testSets() {
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
	fmt.Print(ediblePlants.ToCSV()) //lettuce carrot tomato apple banana (no order)

	//intersection (in place) test
	vegetables.IntersectWith(fruits)
	fmt.Println("this is both a fruit and a vegetable:")
	fmt.Print(vegetables.ToCSV()) //tomato, 

	//union (in place) test
	ediblePlants.Remove("tomato")
	ediblePlants.UnionWith(fruits) //bring back tomato
	fmt.Print(ediblePlants.ToCSV()) //apple banana lettuce carrot tomato (no order)
}

//to be used with Go's data race testing feature (go test -race set.go)
func testSetsConcurrent() {
	s1 := NewSafeSet()
	s2 := NewSafeSet()
	s3 := NewSafeSet()
	var wg sync.WaitGroup
	wg.Add(10)
	for i := 0; i < 10; i ++ {
		go testRoutine(s1, s2, s3, &wg)
	}
	wg.Wait()
	fmt.Println("done")
}

//tries every function, many concurrent adds/removes
func testRoutine(s1 *safeSet, s2 *safeSet, s3 *safeSet, wg *sync.WaitGroup){
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

func main() {
	testSetsConcurrent()
}