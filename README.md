# persistent-social

**Topological data analysis for social networks in Go — Vietoris-Rips filtration, persistent homology, Wasserstein distances, and mobility/stratification metrics.**

Computes the topological fingerprint of social networks. People are vertices, relationships are edges with strength weights. Builds a Vietoris-Rips filtration, computes H⁰ and H¹ persistence, and extracts interpretable metrics: mobility score (how transient relationships are), stratification index (how many permanent structures exist), and community count.

## What This Gives You

- **Social graph construction** — people with attributes, weighted edges
- **Vietoris-Rips filtration** — growing epsilon threshold on edge weights
- **Persistent homology** — H⁰ (communities) and H¹ (structural holes/bridges)
- **Wasserstein distance** — compare two social networks topologically
- **Mobility score** — average persistence length (transient vs stable ties)
- **Stratification index** — fraction of essential (permanent) features
- **Betti numbers** — community count and loop count at any threshold

## Quick Start

```go
package main

import ps "github.com/SuperInstance/persistent-social"

func main() {
    g := ps.NewSocialGraph()
    g.AddPerson(ps.Person{ID: "alice", Name: "Alice", Attributes: map[string]float64{"age": 30}})
    g.AddPerson(ps.Person{ID: "bob", Name: "Bob", Attributes: map[string]float64{"age": 25}})
    g.AddEdge(ps.Edge{From: "alice", To: "bob", Weight: 0.8})

    // Compute persistence
    edges := g.Edges()
    filt := ps.BuildRipsFiltration(edges, 2) // max dim = 2 (triangles)
    pd := ps.ComputePersistence(filt)

    // Metrics
    betti := ps.BettiNumbersFromDiagram(pd, 0.5)
    mobility := ps.MobilityScoreFromDiagram(pd)
    strat := ps.StratificationIndexFromDiagram(pd)

    fmt.Printf("Betti: %v, Mobility: %.3f, Stratification: %.3f\n", betti, mobility, strat)
}
```

## Installation

```bash
go get github.com/SuperInstance/persistent-social
```

## Testing

```bash
go test ./...
```

## How It Fits

Part of the SuperInstance ecosystem:

- **[persistent-sheaf](https://github.com/SuperInstance/persistent-sheaf)** — Rust persistent sheaf cohomology
- **[topology-lab](https://github.com/SuperInstance/topology-lab)** — Interactive WASM visualization
- **persistent-social** — Social network TDA in Go (this repo)

## License

MIT
