//reference: https://www.davidkaya.com/Sets-in-golang/
package Set
import (
	"sync"
	"fmt"
	"strings"
	"strconv"
)


type Set struct {
	//using struct{} because an empty struct takes up 0 bytes
	mp map[string]struct{}
}

func NewSet() *Set {
	s := &Set{}
	s.mp = make(map[string]struct{})
	return s
}

/*unlock the Set*/
func (s *Set) Contains(item string) bool {
	_, ok := s.mp[item]
	return ok
}

func (s *Set) Add(item string) {
	s.mp[item] = struct{}{}
}

func (s *Set) Remove(item string) {
	delete(s.mp, item)
}

//expands the Set into its union with another Set
func (s1 *Set) UnionWith(s2 *Set) {
	for key, _ := range(s2.mp) {
		s1.mp[key] = struct{}{}
	}
}

//reduces the Set to its intersection with another Set
func (s1 *Set) IntersectWith(s2 *Set) {
	for key, _ := range(s1.mp) {
		if !(s2.Contains(key)) {
			s1.Remove(key)
		}
	}
}

//returns a new Set which is the union of two Sets
func Union(s1 *Set, s2 *Set) *Set {
	s3 := NewSet()
	for key, _ := range(s1.mp) {
		s3.mp[key] = struct{}{}
	}
	for key, _ := range(s2.mp) {
		s3.mp[key] = struct{}{}
	}
	return s3
}

func Intersection(s1 *Set, s2 *Set) *Set {
	s3 := NewSet()
	for key, _ := range(s1.mp) {
		if s2.Contains(key) {
			s3.mp[key] = struct{}{}
		}
	}
	return s3
}

//returns a new Set that is s1 U s2'
func IntersectionComplement(s1 *Set, s2 *Set) *Set {
	s3 := NewSet()
	for key, _ := range(s1.mp) {
		if !s2.Contains(key) {
			s3.mp[key] = struct{}{}
		}
	}
	return s3
}

//turns a Set into a CSV
func (s *Set) ToCSV() string {
	str := &strings.Builder{}
	for item, _ := range(s.mp) {
		str.WriteString(item + ",")
	}
	str.WriteRune('\n')
	return str.String()
}

type SafeSet struct {
	st *Set
	lock sync.Mutex
}

func NewSafeSet() *SafeSet {
	var sf SafeSet
	sf.st = NewSet()
	return &sf
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

func (s *SafeSet) Contains(item string) bool {
	s.lock.Lock()
	defer s.lock.Unlock()
	t := s.st.Contains(item)
	return t
}

func (s1 *SafeSet) UnionWith(s2 *SafeSet) {
	s1.lock.Lock()
	defer s1.lock.Unlock()
	s2.lock.Lock()
	defer s2.lock.Unlock()
	s1.st.UnionWith(s2.st)
}

func (s1 *SafeSet) IntersectWith(s2 *SafeSet) {
	s1.lock.Lock()
	defer s1.lock.Unlock()
	s2.lock.Lock()
	defer s2.lock.Unlock()
	s1.st.IntersectWith(s2.st)
}

func SafeUnion(s1 *SafeSet, s2 *SafeSet) *SafeSet {
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

func (s *SafeSet) ToCSV() string {
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

	//union (new Set) test
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

//to be used with Go's data race testing feature (go test -race Set.go)
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