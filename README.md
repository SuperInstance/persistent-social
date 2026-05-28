# persistent-social

A production-quality Go library implementing **persistent homology** for social network analysis. Pure Go, no external dependencies, goroutine-safe.

## Features

- **Vietoris-Rips filtration** built from social graph connections
- **H0 persistence** via union-find (connected component analysis)
- **H1 persistence** via boundary matrix reduction (loop detection)
- **Betti numbers** (β₀ = components, β₁ = loops)
- **Wasserstein & Bottleneck distance** for comparing network topologies
- **Mobility score** — measures topological dynamism across filtration
- **Stratification index** — measures hierarchical layering
- Handles **10,000+ people** efficiently
- **Goroutine-safe** with `sync.RWMutex`

## Installation

```bash
go get github.com/SuperInstance/persistent-social
```

## Quick Start

```go
package main

import (
    "fmt"
    ps "github.com/SuperInstance/persistent-social"
)

func main() {
    g := ps.NewSocialGraph()

    // Add people
    g.AddPerson(ps.Person{ID: "alice", Name: "Alice"})
    g.AddPerson(ps.Person{ID: "bob", Name: "Bob"})
    g.AddPerson(ps.Person{ID: "carol", Name: "Carol"})

    // Add connections (strength 0-1, higher = stronger)
    g.AddConnection(ps.Connection{From: "alice", To: "bob", Strength: 0.9})
    g.AddConnection(ps.Connection{From: "bob", To: "carol", Strength: 0.6})
    g.AddConnection(ps.Connection{From: "alice", To: "carol", Strength: 0.3})

    // Compute topology
    betti := g.BettiNumbers()
    fmt.Printf("β₀=%d (components), β₁=%d (loops)\n", betti[0], betti[1])

    // Persistence diagram
    pd := g.PersistenceDiagram()
    fmt.Printf("Total persistence: %.4f\n", pd.TotalPersistence())
    fmt.Printf("Essential features: %d\n", len(pd.EssentialFeatures()))

    // Network metrics
    fmt.Printf("Mobility score: %.4f\n", g.MobilityScore())
    fmt.Printf("Stratification index: %.4f\n", g.StratificationIndex())

    // Compare two networks
    g2 := ps.NewSocialGraph()
    // ... build second network ...
    distance := ps.CompareNetworks(g, g2)
    fmt.Printf("Network distance: %.4f\n", distance)
}
```

## API

### Core Types

```go
type Person struct {
    ID         string
    Name       string
    Attributes map[string]float64
}

type Connection struct {
    From, To   string
    Strength   float64
    Timestamp  int64
}
```

### SocialGraph

```go
g := ps.NewSocialGraph()
g.AddPerson(person)
g.AddConnection(connection)

g.BuildFiltration()         // *Filtration
g.PersistenceDiagram()      // *PersistenceDiagram
g.BettiNumbers()            // []int{β₀, β₁}
g.MobilityScore()           // float64
g.StratificationIndex()     // float64
```

### PersistenceDiagram

```go
pd := g.PersistenceDiagram()

pd.TotalPersistence()                        // sum of lifetimes
pd.EssentialFeatures()                        // features that never die
pd.BottleneckDistance(otherDiagram)           // ∞-Wasserstein
pd.WassersteinDistance(otherDiagram, p)       // p-Wasserstein
```

### Network Comparison

```go
distance := ps.CompareNetworks(graph1, graph2) // Wasserstein-2 distance
```

## How It Works

1. **Filtration**: Connections are sorted by weight (1/strength). Simplices are added as the threshold increases — vertices first, then edges, then triangles.

2. **H0 Persistence**: Union-find tracks connected components. When two components merge, a persistence point (birth=0, death=threshold) is recorded.

3. **H1 Persistence**: A boundary matrix is constructed for triangles and reduced using standard column operations. Non-zero paired columns give H1 birth-death pairs.

4. **Metrics**: Betti numbers count living features; mobility and stratification summarize the persistence diagram.

## Performance

- 10,000 people with chain topology: **~60ms** for Betti number computation
- Concurrent-safe: supports parallel reads and writes via `sync.RWMutex`
- Hungarian algorithm for exact Wasserstein distance on small diagrams; greedy approximation for large ones

## License

MIT
