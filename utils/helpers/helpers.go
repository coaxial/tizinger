// Package helpers provides various functions to achieve small operations that
// are useful through the codebase and across packages.
package helpers

// Uniq removes duplicate ints in a slice of ints and returns the result.
func Uniq(s []int) (ds []int) {
	// note: using s []int because a slice is already a pointer to an
	// array, so it won't copy the whole slice.
	seen := make(map[int]bool)
	for _, el := range s {
		if !seen[el] {
			seen[el] = true
			ds = append(ds, el)
		}
	}
	return ds
}
