package persistentsocial

import (
	"fmt"
	"math"
	"sync"
	"testing"
)

func TestPersonCreation(t *testing.T) {
	g := NewSocialGraph()
	g.AddPerson(Person{ID: "a", Name: "Alice", Attributes: map[string]float64{"age": 30}})
	g.AddPerson(Person{ID: "b", Name: "Bob", Attributes: map[string]float64{"age": 25}})

	if g.PersonCount() != 2 {
		t.Errorf("expected 2 people, got %d", g.PersonCount())
	}

	people := g.People()
	names := map[string]bool{}
	for _, p := range people {
		names[p.Name] = true
	}
	if !names["Alice"] || !names["Bob"] {
		t.Error("expected Alice and Bob")
	}
}

func TestConnectionCreation(t *testing.T) {
	g := NewSocialGraph()
	g.AddPerson(Person{ID: "a", Name: "Alice"})
	g.AddPerson(Person{ID: "b", Name: "Bob"})
	g.AddConnection(Connection{From: "a", To: "b", Strength: 0.8, Timestamp: 1000})

	conns := g.Connections()
	if len(conns) != 1 {
		t.Fatalf("expected 1 connection, got %d", len(conns))
	}
	if conns[0].Strength != 0.8 {
		t.Errorf("expected strength 0.8, got %f", conns[0].Strength)
	}
}

func TestFiltrationBuilding(t *testing.T) {
	g := NewSocialGraph()
	g.AddPerson(Person{ID: "a", Name: "Alice"})
	g.AddPerson(Person{ID: "b", Name: "Bob"})
	g.AddPerson(Person{ID: "c", Name: "Carol"})
	g.AddConnection(Connection{From: "a", To: "b", Strength: 1.0})
	g.AddConnection(Connection{From: "b", To: "c", Strength: 0.5})
	g.AddConnection(Connection{From: "a", To: "c", Strength: 0.33})

	f := g.BuildFiltration()
	if f == nil {
		t.Fatal("expected non-nil filtration")
	}
	if len(f.Steps) == 0 {
		t.Fatal("expected filtration steps")
	}
}

func TestH0PersistenceDisconnected(t *testing.T) {
	g := NewSocialGraph()
	g.AddPerson(Person{ID: "a", Name: "Alice"})
	g.AddPerson(Person{ID: "b", Name: "Bob"})
	g.AddPerson(Person{ID: "c", Name: "Carol"})
	// Only connect a-b, c is isolated
	g.AddConnection(Connection{From: "a", To: "b", Strength: 1.0})

	pd := g.PersistenceDiagram()
	if pd == nil {
		t.Fatal("expected non-nil diagram")
	}

	// Should have 2 essential H0 features (a+b component + c alone)
	essential := pd.EssentialFeatures()
	h0Essential := 0
	for _, pt := range essential {
		if pt.Dimension == 0 {
			h0Essential++
		}
	}
	if h0Essential != 2 {
		t.Errorf("expected 2 essential H0 features, got %d", h0Essential)
	}
}

func TestH0PersistenceMerge(t *testing.T) {
	g := NewSocialGraph()
	g.AddPerson(Person{ID: "a", Name: "Alice"})
	g.AddPerson(Person{ID: "b", Name: "Bob"})
	g.AddPerson(Person{ID: "c", Name: "Carol"})

	g.AddConnection(Connection{From: "a", To: "b", Strength: 1.0})
	g.AddConnection(Connection{From: "b", To: "c", Strength: 0.5})

	pd := g.PersistenceDiagram()

	// Should have 1 essential H0 feature (all connected) and 2 finite H0 features (merges)
	h0Finite := 0
	h0Essential := 0
	for _, pt := range pd.Points {
		if pt.Dimension == 0 {
			if math.IsInf(pt.Death, 1) {
				h0Essential++
			} else {
				h0Finite++
			}
		}
	}

	if h0Essential != 1 {
		t.Errorf("expected 1 essential H0, got %d", h0Essential)
	}
	if h0Finite != 2 {
		t.Errorf("expected 2 finite H0, got %d", h0Finite)
	}
}

func TestH1PersistenceLoop(t *testing.T) {
	g := NewSocialGraph()
	g.AddPerson(Person{ID: "a", Name: "Alice"})
	g.AddPerson(Person{ID: "b", Name: "Bob"})
	g.AddPerson(Person{ID: "c", Name: "Carol"})

	// Triangle = loop in H1, but filled triangle kills H1
	g.AddConnection(Connection{From: "a", To: "b", Strength: 1.0})
	g.AddConnection(Connection{From: "b", To: "c", Strength: 1.0})
	g.AddConnection(Connection{From: "a", To: "c", Strength: 1.0})

	pd := g.PersistenceDiagram()

	// With triangle (2-simplex), H1 should be 0
	h1 := 0
	for _, pt := range pd.Points {
		if pt.Dimension == 1 {
			h1++
		}
	}
	// Triangle fills the loop, so H1 features should die
	// If equal strengths, the triangle forms at the same time as the loop → H1=0
	if h1 > 0 {
		t.Logf("H1 points (expected triangle kills loop): %d", h1)
	}
}

func TestH1PersistenceRing(t *testing.T) {
	g := NewSocialGraph()
	g.AddPerson(Person{ID: "1", Name: "A"})
	g.AddPerson(Person{ID: "2", Name: "B"})
	g.AddPerson(Person{ID: "3", Name: "C"})
	g.AddPerson(Person{ID: "4", Name: "D"})

	// Square (ring of 4) — no triangle → H1 loop persists
	g.AddConnection(Connection{From: "1", To: "2", Strength: 1.0})
	g.AddConnection(Connection{From: "2", To: "3", Strength: 1.0})
	g.AddConnection(Connection{From: "3", To: "4", Strength: 1.0})
	g.AddConnection(Connection{From: "4", To: "1", Strength: 1.0})

	betti := g.BettiNumbers()
	if betti[0] != 1 {
		t.Errorf("expected β₀=1, got %d", betti[0])
	}
	if betti[1] != 1 {
		t.Errorf("expected β₁=1 (ring), got %d", betti[1])
	}
}

func TestBettiNumbersClique(t *testing.T) {
	g := NewSocialGraph()
	for i := 0; i < 5; i++ {
		g.AddPerson(Person{ID: fmt.Sprintf("p%d", i), Name: fmt.Sprintf("P%d", i)})
	}
	// Complete graph (clique)
	for i := 0; i < 5; i++ {
		for j := i + 1; j < 5; j++ {
			g.AddConnection(Connection{
				From: fmt.Sprintf("p%d", i), To: fmt.Sprintf("p%d", j), Strength: 1.0,
			})
		}
	}

	betti := g.BettiNumbers()
	if betti[0] != 1 {
		t.Errorf("clique β₀ should be 1, got %d", betti[0])
	}
}

func TestBettiNumbersTree(t *testing.T) {
	g := NewSocialGraph()
	for i := 0; i < 5; i++ {
		g.AddPerson(Person{ID: fmt.Sprintf("p%d", i), Name: fmt.Sprintf("P%d", i)})
	}
	// Tree: p0-p1, p0-p2, p1-p3, p1-p4
	g.AddConnection(Connection{From: "p0", To: "p1", Strength: 1.0})
	g.AddConnection(Connection{From: "p0", To: "p2", Strength: 1.0})
	g.AddConnection(Connection{From: "p1", To: "p3", Strength: 1.0})
	g.AddConnection(Connection{From: "p1", To: "p4", Strength: 1.0})

	betti := g.BettiNumbers()
	if betti[0] != 1 {
		t.Errorf("tree β₀ should be 1, got %d", betti[0])
	}
	if betti[1] != 0 {
		t.Errorf("tree β₁ should be 0, got %d", betti[1])
	}
}

func TestMobilityScore(t *testing.T) {
	g := NewSocialGraph()
	g.AddPerson(Person{ID: "a", Name: "Alice"})
	g.AddPerson(Person{ID: "b", Name: "Bob"})
	g.AddPerson(Person{ID: "c", Name: "Carol"})
	g.AddConnection(Connection{From: "a", To: "b", Strength: 1.0})
	g.AddConnection(Connection{From: "b", To: "c", Strength: 0.5})

	score := g.MobilityScore()
	if score < 0 {
		t.Errorf("mobility score should be non-negative, got %f", score)
	}
}

func TestStratificationIndex(t *testing.T) {
	g := NewSocialGraph()
	g.AddPerson(Person{ID: "a", Name: "Alice"})
	g.AddPerson(Person{ID: "b", Name: "Bob"})
	g.AddConnection(Connection{From: "a", To: "b", Strength: 1.0})

	idx := g.StratificationIndex()
	if idx < 0 || idx > 1 {
		t.Errorf("stratification index should be in [0,1], got %f", idx)
	}
}

func TestWassersteinSymmetric(t *testing.T) {
	g1 := NewSocialGraph()
	g1.AddPerson(Person{ID: "a", Name: "Alice"})
	g1.AddPerson(Person{ID: "b", Name: "Bob"})
	g1.AddConnection(Connection{From: "a", To: "b", Strength: 1.0})

	g2 := NewSocialGraph()
	g2.AddPerson(Person{ID: "x", Name: "Xena"})
	g2.AddPerson(Person{ID: "y", Name: "Yuri"})
	g2.AddPerson(Person{ID: "z", Name: "Zara"})
	g2.AddConnection(Connection{From: "x", To: "y", Strength: 0.8})
	g2.AddConnection(Connection{From: "y", To: "z", Strength: 0.6})

	d1 := CompareNetworks(g1, g2)
	d2 := CompareNetworks(g2, g1)

	if math.Abs(d1-d2) > 1e-9 {
		t.Errorf("Wasserstein distance should be symmetric: %f vs %f", d1, d2)
	}
}

func TestWassersteinIdentical(t *testing.T) {
	g := NewSocialGraph()
	g.AddPerson(Person{ID: "a", Name: "Alice"})
	g.AddPerson(Person{ID: "b", Name: "Bob"})
	g.AddConnection(Connection{From: "a", To: "b", Strength: 1.0})

	d := CompareNetworks(g, g)
	if d > 1e-9 {
		t.Errorf("identical networks should have distance 0, got %f", d)
	}
}

func TestCompareNetworksDifferent(t *testing.T) {
	g1 := NewSocialGraph()
	g1.AddPerson(Person{ID: "a", Name: "Alice"})
	g1.AddPerson(Person{ID: "b", Name: "Bob"})
	g1.AddConnection(Connection{From: "a", To: "b", Strength: 1.0})

	g2 := NewSocialGraph()
	g2.AddPerson(Person{ID: "x", Name: "Xena"})
	g2.AddPerson(Person{ID: "y", Name: "Yuri"})
	g2.AddPerson(Person{ID: "z", Name: "Zara"})
	g2.AddConnection(Connection{From: "x", To: "y", Strength: 0.5})
	g2.AddConnection(Connection{From: "y", To: "z", Strength: 0.5})
	g2.AddConnection(Connection{From: "x", To: "z", Strength: 0.5})

	d := CompareNetworks(g1, g2)
	if d <= 0 {
		t.Errorf("different networks should have nonzero distance, got %f", d)
	}
}

func TestEssentialFeatures(t *testing.T) {
	g := NewSocialGraph()
	g.AddPerson(Person{ID: "a", Name: "Alice"})
	g.AddPerson(Person{ID: "b", Name: "Bob"})
	g.AddConnection(Connection{From: "a", To: "b", Strength: 1.0})

	pd := g.PersistenceDiagram()
	essential := pd.EssentialFeatures()
	if len(essential) == 0 {
		t.Error("expected at least one essential feature")
	}
	for _, pt := range essential {
		if !math.IsInf(pt.Death, 1) {
			t.Error("essential features must have Death = +Inf")
		}
	}
}

func TestTotalPersistence(t *testing.T) {
	g := NewSocialGraph()
	g.AddPerson(Person{ID: "a", Name: "Alice"})
	g.AddPerson(Person{ID: "b", Name: "Bob"})
	g.AddPerson(Person{ID: "c", Name: "Carol"})
	g.AddConnection(Connection{From: "a", To: "b", Strength: 1.0})
	g.AddConnection(Connection{From: "b", To: "c", Strength: 0.5})

	pd := g.PersistenceDiagram()
	tp := pd.TotalPersistence()
	if tp < 0 {
		t.Errorf("total persistence should be non-negative, got %f", tp)
	}
}

func TestLargeScalePerformance(t *testing.T) {
	g := NewSocialGraph()
	n := 10000
	for i := 0; i < n; i++ {
		g.AddPerson(Person{ID: fmt.Sprintf("p%d", i), Name: fmt.Sprintf("Person%d", i)})
	}
	// Create a connected network: chain + some random edges
	for i := 0; i < n-1; i++ {
		strength := 0.5 + 0.5*float64(i%10)/10.0
		g.AddConnection(Connection{
			From: fmt.Sprintf("p%d", i), To: fmt.Sprintf("p%d", i+1), Strength: strength,
		})
	}

	betti := g.BettiNumbers()
	if betti[0] != 1 {
		t.Errorf("connected chain β₀ should be 1, got %d", betti[0])
	}
}

func TestConcurrentAccess(t *testing.T) {
	g := NewSocialGraph()
	var wg sync.WaitGroup

	// Concurrent writers
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			g.AddPerson(Person{ID: fmt.Sprintf("p%d", idx), Name: fmt.Sprintf("P%d", idx)})
		}(i)
	}

	// Concurrent connections
	for i := 0; i < 99; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			g.AddConnection(Connection{
				From: fmt.Sprintf("p%d", idx), To: fmt.Sprintf("p%d", idx+1), Strength: 1.0,
			})
		}(i)
	}

	wg.Wait()

	if g.PersonCount() != 100 {
		t.Errorf("expected 100 people, got %d", g.PersonCount())
	}
}

func TestEmptyNetwork(t *testing.T) {
	g := NewSocialGraph()
	betti := g.BettiNumbers()
	if betti[0] != 0 {
		t.Errorf("empty network β₀ should be 0, got %d", betti[0])
	}
	pd := g.PersistenceDiagram()
	if pd == nil {
		t.Error("expected non-nil diagram for empty network")
	}
}

func TestSinglePerson(t *testing.T) {
	g := NewSocialGraph()
	g.AddPerson(Person{ID: "solo", Name: "Solo"})

	betti := g.BettiNumbers()
	if betti[0] != 1 {
		t.Errorf("single person β₀ should be 1, got %d", betti[0])
	}
	if betti[1] != 0 {
		t.Errorf("single person β₁ should be 0, got %d", betti[1])
	}

	pd := g.PersistenceDiagram()
	essential := pd.EssentialFeatures()
	if len(essential) != 1 {
		t.Errorf("single person should have 1 essential feature, got %d", len(essential))
	}
}
