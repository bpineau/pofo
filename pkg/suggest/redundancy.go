package suggest

import (
	"math"
	"sort"
)

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
	// Union near-identical, same-class pairs, keeping every pair's correlation:
	// the weakest link cannot be tracked as the groups form, because a union
	// moves the root and would orphan what was recorded under the old one (A~B
	// at 0.95 then A~C at 0.99 reported 0.99 as the group's weakest pair). It is
	// computed below, once the membership is final. Every pair inside a group
	// shares its asset class by transitivity, so every one of them is in here.
	corr := map[[2]int]float64{}
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			if holdings[i].Meta.AssetClass != holdings[j].Meta.AssetClass {
				continue
			}
			c := Correlation(returns[i], returns[j])
			corr[[2]int{i, j}] = c
			if c < threshold {
				continue
			}
			if ri, rj := find(i), find(j); ri != rj {
				parent[ri] = rj
			}
		}
	}

	members := map[int][]int{}
	for i := 0; i < n; i++ {
		r := find(i)
		members[r] = append(members[r], i)
	}
	var groups []Group
	for _, idx := range members {
		if len(idx) < 2 {
			continue
		}
		sort.Ints(idx)
		g := Group{MinCorr: math.Inf(1)}
		for a, i := range idx {
			g.IDs = append(g.IDs, holdings[i].ID)
			g.Weight += holdings[i].Weight
			for _, j := range idx[a+1:] {
				if c, ok := corr[[2]int{i, j}]; ok && c < g.MinCorr {
					g.MinCorr = c
				}
			}
		}
		groups = append(groups, g)
	}
	sort.Slice(groups, func(a, b int) bool { return groups[a].Weight > groups[b].Weight })
	return groups
}
