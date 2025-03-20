package main

import "fmt"

// Create a set using a map with struct{} as the value type
type Set map[int]struct{}

func (s Set) String() string{
    ret := "{"
    if s.Size()>0{
        for k := range s{
            ret += fmt.Sprint(k)+", "
        }
        ret = ret[:len(ret)-2]
    }
    return ret+"}"
}

// Add an element to the set
func (s Set) Add(element int) {
    s[element] = struct{}{}
}

// Remove an element from the set
func (s Set) Remove(element int) {
    delete(s, element)
}

// Check if an element is in the set
func (s Set) Contains(element int) bool {
    _, exists := s[element]
    return exists
}

// Get the size of the set
func (s Set) Size() int {
    return len(s)
}

// Convert set to a slice (for iterating or displaying elements)
func (s Set) ToSlice() []int {
    elements := make([]int, 0, len(s))
    for key := range s {
        elements = append(elements, key)
    }
    return elements
}


//##############################################################################
//              Functions used by Set of tuples ([]int) in DP
func containsSlice(slices [][]int, target []int) bool {
	for _, s := range slices {
		if equalSlices(s, target) {
			return true
		}
	}
	return false
}

// Helper function to check if two slices are equal
func equalSlices(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// Convert slice to string for use as a map key
func sliceToString(s []int) string {
	return fmt.Sprint(s)
}