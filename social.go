package persistentsocial

import (
	"math"
	"sort"
	"sync"
)

// Person represents an individual in a social network.
type Person struct {
	ID         string
	Name       string
	Attributes map[string]float64
}

// Connection represents a weighted, timestamped connection between two people.
type Connection struct {
	From      string
	To        string
	Strength  float64
	Timestamp int64
}

// SocialGraph represents a social network.
type SocialGraph struct {
	mu          sync.RWMutex
	people      map[string]Person
	connections []Connection
}

// NewSocialGraph creates an empty social graph.
func NewSocialGraph() *SocialGraph {
	return &SocialGraph{
		people: make(map[string]Person),
	}
}

// AddPerson adds a person to the graph.
func (g *SocialGraph) AddPerson(p Person) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.people[p.ID] = p
}

// AddConnection adds a connection to the graph.
func (g *SocialGraph) AddConnection(c Connection) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.connections = append(g.connections, c)
}

// People returns all people in the graph.
func (g *SocialGraph) People() []Person {
	g.mu.RLock()
	defer g.mu.RUnlock()
	result := make([]Person, 0, len(g.people))
	for _, p := range g.people {
		result = append(result, p)
	}
	return result
}

// Connections returns all connections.
func (g *SocialGraph) Connections() []Connection {
	g.mu.RLock()
	defer g.mu.RUnlock()
	result := make([]Connection, len(g.connections))
	copy(result, g.connections)
	return result
}

// PersonCount returns the number of people.
func (g *SocialGraph) PersonCount() int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return len(g.people)
}

// BuildFiltration builds a Vietoris-Rips filtration from the social graph.
// Edge weight = 1/strength (stronger connections form earlier in the filtration).
func (g *SocialGraph) BuildFiltration() *Filtration {
	g.mu.RLock()
	defer g.mu.RUnlock()

	var edges []Edge
	for _, c := range g.connections {
		w := 1.0 / c.Strength
		if c.Strength == 0 {
			w = math.Inf(1)
		}
		edges = append(edges, Edge{From: c.From, To: c.To, Weight: w})
	}

	return BuildRipsFiltration(edges, 2)
}

// PersistenceDiagram computes the full persistence diagram (H0 + H1).
func (g *SocialGraph) PersistenceDiagram() *PersistenceDiagram {
	g.mu.RLock()
	defer g.mu.RUnlock()

	vertices := make([]string, 0, len(g.people))
	for id := range g.people {
		vertices = append(vertices, id)
	}
	sort.Strings(vertices)

	f := g.buildFiltrationUnsafe()
	return ComputePersistence(f, vertices)
}

// BettiNumbers returns [β₀, β₁] for the graph.
func (g *SocialGraph) BettiNumbers() []int {
	g.mu.RLock()
	defer g.mu.RUnlock()

	vertices := make([]string, 0, len(g.people))
	for id := range g.people {
		vertices = append(vertices, id)
	}

	// Compute β₀ = number of connected components via union-find
	uf := newUnionFind(vertices)
	for _, c := range g.connections {
		if c.Strength > 0 {
			uf.Union(c.From, c.To)
		}
	}
	b0 := uf.ComponentCount()

	// β₁ = |edges| - |vertices| + β₀ (Euler formula for 1-complex)
	edgeSet := make(map[string]bool)
	for _, c := range g.connections {
		if c.Strength > 0 {
			s := NewSimplex(c.From, c.To)
			edgeSet[s.Key()] = true
		}
	}
	b1 := len(edgeSet) - len(vertices) + b0
	if b1 < 0 {
		b1 = 0
	}

	return []int{b0, b1}
}

// MobilityScore measures how much the topology changes across the filtration.
// High mobility = many short-lived features = dynamic network.
func (g *SocialGraph) MobilityScore() float64 {
	pd := g.PersistenceDiagram()
	if len(pd.Points) == 0 {
		return 0
	}

	total := 0.0
	finiteCount := 0
	for _, pt := range pd.Points {
		if !math.IsInf(pt.Death, 1) {
			total += pt.Death - pt.Birth
			finiteCount++
		}
	}

	if finiteCount == 0 {
		return 0
	}
	return total / float64(len(pd.Points))
}

// StratificationIndex measures how layered/hierarchical the network is.
// Based on the ratio of essential H0 features to total features.
func (g *SocialGraph) StratificationIndex() float64 {
	pd := g.PersistenceDiagram()
	if len(pd.Points) == 0 {
		return 0
	}

	essential := len(pd.EssentialFeatures())
	return float64(essential) / float64(len(pd.Points))
}

func (g *SocialGraph) buildFiltrationUnsafe() *Filtration {
	var edges []Edge
	for _, c := range g.connections {
		w := 1.0 / c.Strength
		if c.Strength == 0 {
			w = math.Inf(1)
		}
		edges = append(edges, Edge{From: c.From, To: c.To, Weight: w})
	}
	return BuildRipsFiltration(edges, 2)
}

// CompareNetworks computes the Wasserstein distance between two social graphs' topologies.
func CompareNetworks(g1, g2 *SocialGraph) float64 {
	pd1 := g1.PersistenceDiagram()
	pd2 := g2.PersistenceDiagram()
	return pd1.WassersteinDistance(pd2, 2.0)
}
