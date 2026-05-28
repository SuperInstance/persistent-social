package persistentsocial

// WassersteinDistance computes the p-Wasserstein distance between two persistence diagrams.
func WassersteinDistance(pd1, pd2 *PersistenceDiagram, p float64) float64 {
	if pd1 == nil {
		pd1 = &PersistenceDiagram{}
	}
	if pd2 == nil {
		pd2 = &PersistenceDiagram{}
	}
	return pd1.WassersteinDistance(pd2, p)
}
