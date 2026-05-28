package persistentsocial

import "math"

// BettiNumbersFromDiagram computes Betti numbers from a persistence diagram.
func BettiNumbersFromDiagram(pd *PersistenceDiagram, threshold float64) []int {
	if pd == nil {
		return []int{0, 0}
	}
	b0 := 0
	b1 := 0
	for _, pt := range pd.Points {
		born := pt.Birth <= threshold
		alive := math.IsInf(pt.Death, 1) || pt.Death > threshold
		if born && alive {
			if pt.Dimension == 0 {
				b0++
			} else if pt.Dimension == 1 {
				b1++
			}
		}
	}
	return []int{b0, b1}
}

// MobilityScoreFromDiagram computes mobility from a persistence diagram directly.
func MobilityScoreFromDiagram(pd *PersistenceDiagram) float64 {
	if pd == nil || len(pd.Points) == 0 {
		return 0
	}
	total := 0.0
	finite := 0
	for _, pt := range pd.Points {
		if !math.IsInf(pt.Death, 1) {
			total += pt.Death - pt.Birth
			finite++
		}
	}
	if finite == 0 {
		return 0
	}
	return total / float64(len(pd.Points))
}

// StratificationIndexFromDiagram computes stratification from a persistence diagram.
func StratificationIndexFromDiagram(pd *PersistenceDiagram) float64 {
	if pd == nil || len(pd.Points) == 0 {
		return 0
	}
	return float64(len(pd.EssentialFeatures())) / float64(len(pd.Points))
}

// AveragePersistence returns the average persistence lifetime.
func AveragePersistence(pd *PersistenceDiagram) float64 {
	if pd == nil || len(pd.Points) == 0 {
		return 0
	}
	total := 0.0
	count := 0
	for _, pt := range pd.Points {
		if !math.IsInf(pt.Death, 1) {
			total += pt.Death - pt.Birth
			count++
		}
	}
	if count == 0 {
		return 0
	}
	return total / float64(count)
}
