package persistentsocial

import "sort"

// Simplex represents a k-dimensional simplex as an ordered list of vertex IDs.
type Simplex struct {
	Vertices []string
	Dim      int
}

// NewSimplex creates a simplex from vertex labels, sorting them for canonical representation.
func NewSimplex(vertices ...string) Simplex {
	sorted := make([]string, len(vertices))
	copy(sorted, vertices)
	sort.Strings(sorted)
	return Simplex{Vertices: sorted, Dim: len(vertices) - 1}
}

// Key returns a canonical string key for deduplication.
func (s Simplex) Key() string {
	// vertices are already sorted
	result := ""
	for i, v := range s.Vertices {
		if i > 0 {
			result += "|"
		}
		result += v
	}
	return result
}

// SimplicialComplex is a collection of simplices.
type SimplicialComplex struct {
	Simplices map[string]Simplex
}

// NewSimplicialComplex creates an empty simplicial complex.
func NewSimplicialComplex() *SimplicialComplex {
	return &SimplicialComplex{Simplices: make(map[string]Simplex)}
}

// Add adds a simplex and all its faces to the complex.
func (sc *SimplicialComplex) Add(s Simplex) {
	sc.Simplices[s.Key()] = s
	// Add all faces recursively
	if s.Dim >= 1 {
		for i := range s.Vertices {
			face := make([]string, 0, len(s.Vertices)-1)
			face = append(face, s.Vertices[:i]...)
			face = append(face, s.Vertices[i+1:]...)
			fs := NewSimplex(face...)
			sc.Add(fs)
		}
	}
}

// Has checks if a simplex exists in the complex.
func (sc *SimplicialComplex) Has(s Simplex) bool {
	_, ok := sc.Simplices[s.Key()]
	return ok
}

// Dimension returns the maximum dimension of any simplex in the complex.
func (sc *SimplicialComplex) Dimension() int {
	maxDim := -1
	for _, s := range sc.Simplices {
		if s.Dim > maxDim {
			maxDim = s.Dim
		}
	}
	return maxDim
}

// CountByDimension returns the number of simplices at each dimension.
func (sc *SimplicialComplex) CountByDimension() map[int]int {
	counts := make(map[int]int)
	for _, s := range sc.Simplices {
		counts[s.Dim]++
	}
	return counts
}
