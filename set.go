//reference: https://www.davidkaya.com/sets-in-golang/
package main


type set struct {
	//using struct{} because an empty struct takes up 0 bytes
	mp map[string]struct{}
}

var exists = struct{}{}

func NewSet() *set {
	s := &set{}
	s.m = make(map[string]struct{})
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
func (s1 *set) UnionTo(s2 *set) {
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
