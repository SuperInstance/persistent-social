package persistentsocial

import "sort"

// Filtration represents a filtered simplicial complex where simplices
// are added at increasing distance thresholds.
type Filtration struct {
	Steps []FiltrationStep
}

// FiltrationStep holds simplices added at a specific threshold.
type FiltrationStep struct {
	Threshold float64
	Simplices []Simplex
}

// BuildRipsFiltration constructs a Vietoris-Rips filtration from edge weights.
// edges: list of (from, to, weight) tuples. Vertices are extracted from edges.
// maxDim: maximum simplex dimension (0=vertices, 1=edges, 2=triangles, etc.)
func BuildRipsFiltration(edges []Edge, maxDim int) *Filtration {
	if len(edges) == 0 {
		return &Filtration{}
	}

	// Collect unique thresholds
	thresholdSet := make(map[float64]struct{})
	vertexSet := make(map[string]struct{})
	for _, e := range edges {
		thresholdSet[e.Weight] = struct{}{}
		vertexSet[e.From] = struct{}{}
		vertexSet[e.To] = struct{}{}
	}

	thresholds := make([]float64, 0, len(thresholdSet))
	for t := range thresholdSet {
		thresholds = append(thresholds, t)
	}
	sort.Float64s(thresholds)

	// Build adjacency at each threshold
	f := &Filtration{}

	for _, thresh := range thresholds {
		// Active edges at this threshold
		activeEdges := make(map[string][]string) // vertex -> neighbors
		edgeKeys := make(map[string]bool)

		for _, e := range edges {
			if e.Weight <= thresh {
				activeEdges[e.From] = append(activeEdges[e.From], e.To)
				activeEdges[e.To] = append(activeEdges[e.To], e.From)
				s := NewSimplex(e.From, e.To)
				edgeKeys[s.Key()] = true
			}
		}

		var simplices []Simplex

		// Add edges at this threshold
		for k := range edgeKeys {
			// Reconstruct simplex from key — we stored edge simplices above
			for _, e := range edges {
				if e.Weight <= thresh {
					s := NewSimplex(e.From, e.To)
					if s.Key() == k {
						simplices = append(simplices, s)
						break
					}
				}
			}
		}

		// Build higher-dimensional simplices (triangles for dim >= 2)
		if maxDim >= 2 {
			// Find all triangles: for each vertex pair with edge, check common neighbors
			seen := make(map[string]bool)
			vertices := make([]string, 0, len(vertexSet))
			for v := range vertexSet {
				vertices = append(vertices, v)
			}
			sort.Strings(vertices)

			for _, v1 := range vertices {
				n1, ok := activeEdges[v1]
				if !ok {
					continue
				}
				for _, v2 := range n1 {
					if v2 <= v1 {
						continue
					}
					n2, ok := activeEdges[v2]
					if !ok {
						continue
					}
					// Common neighbors form triangles with v1,v2
					for _, v3 := range n2 {
						if v3 <= v2 {
							continue
						}
						// Check if v1-v3 edge exists
						for _, n := range n1 {
							if n == v3 {
								tri := NewSimplex(v1, v2, v3)
								if !seen[tri.Key()] {
									seen[tri.Key()] = true
									simplices = append(simplices, tri)
								}
							}
						}
					}
				}
			}
		}

		f.Steps = append(f.Steps, FiltrationStep{
			Threshold: thresh,
			Simplices: simplices,
		})
	}

	return f
}

// Edge represents a weighted edge in a graph.
type Edge struct {
	From, To string
	Weight   float64
}
