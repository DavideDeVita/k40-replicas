package main

// Create a set using a map with struct{} as the value type
type Set map[int]struct{}

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