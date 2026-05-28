package persistentsocial

import (
	"math"
	"sort"
)

// PersistencePoint represents a point in a persistence diagram.
type PersistencePoint struct {
	Birth     float64
	Death     float64
	Dimension float64
}

// PersistenceDiagram is the output of persistence computation.
type PersistenceDiagram struct {
	Points []PersistencePoint
}

// EssentialFeatures returns points with Death = +Inf (features that never die).
func (pd *PersistenceDiagram) EssentialFeatures() []PersistencePoint {
	var essential []PersistencePoint
	for _, p := range pd.Points {
		if math.IsInf(p.Death, 1) {
			essential = append(essential, p)
		}
	}
	return essential
}

// TotalPersistence returns sum of (death - birth) for all finite points.
func (pd *PersistenceDiagram) TotalPersistence() float64 {
	total := 0.0
	for _, p := range pd.Points {
		if !math.IsInf(p.Death, 1) {
			total += p.Death - p.Birth
		}
	}
	return total
}

// BottleneckDistance computes the bottleneck distance between two diagrams.
func (pd *PersistenceDiagram) BottleneckDistance(other *PersistenceDiagram) float64 {
	return pd.WassersteinDistance(other, math.Inf(1))
}

// WassersteinDistance computes the p-Wasserstein distance between two diagrams.
func (pd *PersistenceDiagram) WassersteinDistance(other *PersistenceDiagram, p float64) float64 {
	// Match points by dimension
	dims := make(map[float64]struct{})
	for _, pt := range pd.Points {
		dims[pt.Dimension] = struct{}{}
	}
	for _, pt := range other.Points {
		dims[pt.Dimension] = struct{}{}
	}

	total := 0.0
	for dim := range dims {
		total += math.Pow(wassersteinDim(pd.Points, other.Points, dim, p), math.Min(p, 1.0))
	}
	if p < math.Inf(1) {
		return math.Pow(total, 1.0/math.Min(p, 1.0))
	}
	return total
}

func wassersteinDim(pts1, pts2 []PersistencePoint, dim float64, p float64) float64 {
	var a, b []PersistencePoint
	for _, pt := range pts1 {
		if pt.Dimension == dim {
			a = append(a, pt)
		}
	}
	for _, pt := range pts2 {
		if pt.Dimension == dim {
			b = append(b, pt)
		}
	}

	n := len(a)
	m := len(b)
	size := n + m
	if size == 0 {
		return 0
	}

	// Build cost matrix for bipartite matching
	// a[i] can match b[j], or diagonal (death self at cost dist to diagonal)
	diag := func(pt PersistencePoint) float64 {
		mid := (pt.Birth + pt.Death) / 2.0
		if math.IsInf(pt.Death, 1) {
			return math.Inf(1)
		}
		return math.Abs(pt.Birth-mid) + math.Abs(pt.Death-mid)
	}

	cost := func(ai, bj int) float64 {
		if ai < n && bj < m {
			if math.IsInf(a[ai].Death, 1) && math.IsInf(b[bj].Death, 1) {
				return math.Abs(a[ai].Birth - b[bj].Birth)
			}
			if math.IsInf(a[ai].Death, 1) || math.IsInf(b[bj].Death, 1) {
				return math.Inf(1)
			}
			return math.Max(math.Abs(a[ai].Birth-b[bj].Birth), math.Abs(a[ai].Death-b[bj].Death))
		}
		if ai < n && bj >= m {
			return diag(a[ai])
		}
		if ai >= n && bj < m {
			return diag(b[bj])
		}
		return 0 // diagonal-to-diagonal
	}

	// Hungarian algorithm for small sizes, greedy for large
	if size <= 64 {
		return hungarianWasserstein(cost, n, m, p)
	}
	return greedyWasserstein(cost, n, m, p)
}

func greedyWasserstein(cost func(int, int) float64, n, m int, p float64) float64 {
	size := n + m
	used := make([]bool, size)
	costs := make([]float64, 0, size)

	for i := 0; i < size; i++ {
		bestJ := -1
		bestC := math.Inf(1)
		for j := 0; j < size; j++ {
			if used[j] {
				continue
			}
			c := cost(i, j)
			if c < bestC {
				bestC = c
				bestJ = j
			}
		}
		if bestJ >= 0 {
			used[bestJ] = true
			costs = append(costs, bestC)
		}
	}

	if math.IsInf(p, 1) {
		maxC := 0.0
		for _, c := range costs {
			if c > maxC {
				maxC = c
			}
		}
		return maxC
	}

	total := 0.0
	for _, c := range costs {
		total += math.Pow(c, p)
	}
	return math.Pow(total, 1.0/p)
}

// Hungarian algorithm for exact matching
func hungarianWasserstein(cost func(int, int) float64, n, m int, p float64) float64 {
	size := n + m
	// Build cost matrix
	C := make([][]float64, size)
	for i := range C {
		C[i] = make([]float64, size)
		for j := range C {
			C[i][j] = cost(i, j)
		}
	}

	// Simplified Hungarian using potential method
	u := make([]float64, size+1)
	v := make([]float64, size+1)
	pArr := make([]int, size+1)
	way := make([]int, size+1)

	for i := 1; i <= size; i++ {
		pArr[0] = i
		j0 := 0
		minv := make([]float64, size+1)
		used := make([]bool, size+1)
		for j := range minv {
			minv[j] = math.Inf(1)
		}
		for {
			used[j0] = true
			i0 := pArr[j0]
			delta := math.Inf(1)
			j1 := -1
			for j := 1; j <= size; j++ {
				if used[j] {
					continue
				}
				cur := C[i0-1][j-1] - u[i0] - v[j]
				if cur < minv[j] {
					minv[j] = cur
					way[j] = j0
				}
				if minv[j] < delta {
					delta = minv[j]
					j1 = j
				}
			}
			for j := 0; j <= size; j++ {
				if used[j] {
					u[pArr[j]] += delta
					v[j] -= delta
				} else {
					minv[j] -= delta
				}
			}
			j0 = j1
			if pArr[j0] == 0 {
				break
			}
		}
		for j := j0; j != 0; {
			pArr[j] = pArr[way[j]]
			j = way[j]
		}
	}

	// Extract assignment costs
	maxCost := 0.0
	sumCost := 0.0
	for j := 1; j <= size; j++ {
		if pArr[j] != 0 {
			c := C[pArr[j]-1][j-1]
			if math.IsInf(p, 1) {
				if c > maxCost {
					maxCost = c
				}
			} else {
				sumCost += math.Pow(c, p)
			}
		}
	}

	if math.IsInf(p, 1) {
		return maxCost
	}
	return math.Pow(sumCost, 1.0/p)
}

// ComputePersistence computes H0 and H1 persistence from a filtration using
// union-find for H0 and a boundary matrix reduction for H1.
func ComputePersistence(f *Filtration, vertices []string) *PersistenceDiagram {
	if len(f.Steps) == 0 {
		// Handle vertices with no edges
		pd := &PersistenceDiagram{}
		for range vertices {
			pd.Points = append(pd.Points, PersistencePoint{Birth: 0, Death: math.Inf(1), Dimension: 0})
		}
		return pd
	}

	pd := &PersistenceDiagram{}

	// H0 via union-find
	uf := newUnionFind(vertices)
	seen := make(map[string]bool)

	for _, step := range f.Steps {
		for _, sx := range step.Simplices {
			if sx.Dim == 1 {
				a, b := sx.Vertices[0], sx.Vertices[1]
				if seen[sx.Key()] {
					continue
				}
				seen[sx.Key()] = true
				if uf.Find(a) != uf.Find(b) {
					uf.Union(a, b)
					// H0 feature dies at this threshold
					pd.Points = append(pd.Points, PersistencePoint{
						Birth:     0,
						Death:     step.Threshold,
						Dimension: 0,
					})
				}
			}
		}
	}

	// Remaining components are essential H0 features
	rootCount := uf.ComponentCount()
	for i := 0; i < rootCount; i++ {
		pd.Points = append(pd.Points, PersistencePoint{
			Birth:     0,
			Death:     math.Inf(1),
			Dimension: 0,
		})
	}

	// H1 via boundary matrix reduction
	pd.Points = append(pd.Points, computeH1(f)...)

	return pd
}

// computeH1 computes H1 persistence points using boundary matrix reduction.
func computeH1(f *Filtration) []PersistencePoint {
	// Collect all edges and triangles across all filtration steps
	type indexedSimplex struct {
		sx        Simplex
		threshold float64
		index     int
	}

	var all []indexedSimplex
	idx := 0
	for _, step := range f.Steps {
		for _, sx := range step.Simplices {
			if sx.Dim == 1 || sx.Dim == 2 {
				all = append(all, indexedSimplex{sx: sx, threshold: step.Threshold, index: idx})
				idx++
			}
		}
	}

	if len(all) == 0 {
		return nil
	}

	// Sort by threshold, then dimension (edges before triangles at same threshold)
	sort.Slice(all, func(i, j int) bool {
		if all[i].threshold != all[j].threshold {
			return all[i].threshold < all[j].threshold
		}
		return all[i].sx.Dim < all[j].sx.Dim
	})

	// Re-index
	for i := range all {
		all[i].index = i
	}

	// Build edge index map
	edgeIdx := make(map[string]int)
	for _, s := range all {
		if s.sx.Dim == 1 {
			edgeIdx[s.sx.Key()] = s.index
		}
	}

	// Build boundary matrix: rows=edges, cols=simplices
	// For triangle, boundary is its 3 edges
	n := len(all)
	boundary := make([]map[int]bool, n)
	for i := range boundary {
		boundary[i] = make(map[int]bool)
	}

	for _, s := range all {
		if s.sx.Dim == 2 {
			// Triangle boundary = 3 edges
			for j := 0; j < 3; j++ {
				face := make([]string, 0, 2)
				face = append(face, s.sx.Vertices[:j]...)
				face = append(face, s.sx.Vertices[j+1:]...)
				fs := NewSimplex(face...)
				if ei, ok := edgeIdx[fs.Key()]; ok {
					boundary[s.index][ei] = true
				}
			}
		}
	}

	// Column reduction (standard algorithm)
	low := make([]int, n) // low(col) = lowest row with 1, or -1
	for i := range low {
		low[i] = -1
	}
	for col := 0; col < n; col++ {
		if all[col].sx.Dim == 2 {
			low[col] = lowestOne(boundary[col])
		}
	}

	// Reduce
	for col := 0; col < n; col++ {
		if low[col] == -1 {
			continue
		}
		for {
			// Find earlier column with same low
			conflict := -1
			for c := 0; c < col; c++ {
				if low[c] == low[col] && low[c] != -1 {
					conflict = c
					break
				}
			}
			if conflict == -1 {
				break
			}
			// XOR columns
			for row := range boundary[conflict] {
				if boundary[col][row] {
					delete(boundary[col], row)
				} else {
					boundary[col][row] = true
				}
			}
			low[col] = lowestOne(boundary[col])
		}
	}

	// Read off H1 pairs
	var points []PersistencePoint
	paired := make(map[int]bool)
	for col := 0; col < n; col++ {
		if low[col] != -1 {
			paired[low[col]] = true
			if all[low[col]].sx.Dim == 1 && all[col].sx.Dim == 2 {
				points = append(points, PersistencePoint{
					Birth:     all[low[col]].threshold,
					Death:     all[col].threshold,
					Dimension: 1,
				})
			}
		}
	}

	return points
}

func lowestOne(col map[int]bool) int {
	lo := -1
	for row := range col {
		if row > lo {
			lo = row
		}
	}
	return lo
}

// union-find
type unionFind struct {
	parent map[string]string
	rank   map[string]int
}

func newUnionFind(vertices []string) *unionFind {
	uf := &unionFind{
		parent: make(map[string]string),
		rank:   make(map[string]int),
	}
	for _, v := range vertices {
		uf.parent[v] = v
		uf.rank[v] = 0
	}
	return uf
}

func (uf *unionFind) Find(x string) string {
	if uf.parent[x] != x {
		uf.parent[x] = uf.Find(uf.parent[x])
	}
	return uf.parent[x]
}

func (uf *unionFind) Union(x, y string) {
	rx, ry := uf.Find(x), uf.Find(y)
	if rx == ry {
		return
	}
	if uf.rank[rx] < uf.rank[ry] {
		rx, ry = ry, rx
	}
	uf.parent[ry] = rx
	if uf.rank[rx] == uf.rank[ry] {
		uf.rank[rx]++
	}
}

func (uf *unionFind) ComponentCount() int {
	count := 0
	for v := range uf.parent {
		if uf.parent[v] == v {
			count++
		}
	}
	return count
}
