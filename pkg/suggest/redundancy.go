package suggest

import "sort"

// Group is a set of holdings that move almost identically and belong to the
// same asset class, effectively one bet held several times.
type Group struct {
	IDs     []string
	Weight  float64 // combined fraction of the portfolio
	MinCorr float64 // weakest pairwise correlation inside the group
}

// Redundancies groups holdings whose daily returns correlate at or above
// threshold and that share an asset class. returns[i] is holding i's
// return series (equal length). Groups of a single asset are omitted;
// the result is ordered by combined weight, descending.
func Redundancies(holdings []Holding, returns [][]float64, threshold float64) []Group {
	n := len(holdings)
	parent := make([]int, n)
	for i := range parent {
		parent[i] = i
	}
	find := func(x int) int {
		for parent[x] != x {
			parent[x] = parent[parent[x]]
			x = parent[x]
		}
		return x
	}
	// Union near-identical, same-class pairs; remember the weakest link.
	minCorr := map[int]float64{}
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			if holdings[i].Meta.AssetClass != holdings[j].Meta.AssetClass {
				continue
			}
			c := Correlation(returns[i], returns[j])
			if c < threshold {
				continue
			}
			ri, rj := find(i), find(j)
			if ri != rj {
				parent[ri] = rj
			}
			root := find(i)
			if v, ok := minCorr[root]; !ok || c < v {
				minCorr[root] = c
			}
		}
	}

	members := map[int][]int{}
	for i := 0; i < n; i++ {
		r := find(i)
		members[r] = append(members[r], i)
	}
	var groups []Group
	for root, idx := range members {
		if len(idx) < 2 {
			continue
		}
		g := Group{MinCorr: minCorr[root]}
		for _, i := range idx {
			g.IDs = append(g.IDs, holdings[i].ID)
			g.Weight += holdings[i].Weight
		}
		groups = append(groups, g)
	}
	sort.Slice(groups, func(a, b int) bool { return groups[a].Weight > groups[b].Weight })
	return groups
}
