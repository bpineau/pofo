package simgen

import (
	"fmt"
	"math"
	"time"

	"portfodor/pkg/marketdata"
	"portfodor/pkg/metrics"
)

// FitBackcast regresses the real asset's daily returns on the given
// components' returns over their overlap (ordinary least squares with an
// intercept), then rebuilds the whole frame period from the fitted linear
// model — residuals are dropped, so the backcast captures only the
// systematic part of the fund's behaviour. It returns the reconstructed
// index, the in-sample R² and the fitted coefficients.
func FitBackcast(fr *Frame, real *marketdata.Series, ids []string) ([]float64, float64, []float64, error) {
	realRet := make(map[time.Time]float64, len(real.Points))
	for i := 1; i < len(real.Points); i++ {
		realRet[real.Points[i].Date] = real.Points[i].Close/real.Points[i-1].Close - 1
	}
	// Design matrix on overlapping dates: [1, comp1, comp2, …].
	var X [][]float64
	var y []float64
	for k := 1; k < len(fr.Dates); k++ {
		r, ok := realRet[fr.Dates[k]]
		if !ok {
			continue
		}
		row := make([]float64, len(ids)+1)
		row[0] = 1
		for i, id := range ids {
			row[i+1] = fr.Returns[id][k]
		}
		X = append(X, row)
		y = append(y, r)
	}
	if len(y) < 120 {
		return nil, 0, nil, fmt.Errorf("overlap insuffisant pour la régression (%d points)", len(y))
	}
	coef, err := ols(X, y)
	if err != nil {
		return nil, 0, nil, err
	}
	// In-sample R².
	my := metrics.Mean(y)
	var ssRes, ssTot float64
	for i, row := range X {
		pred := 0.0
		for j, c := range coef {
			pred += c * row[j]
		}
		ssRes += (y[i] - pred) * (y[i] - pred)
		ssTot += (y[i] - my) * (y[i] - my)
	}
	r2 := 0.0
	if ssTot > 0 {
		r2 = 1 - ssRes/ssTot
	}
	// Rebuild over the full frame.
	values := make([]float64, len(fr.Dates))
	values[0] = 100
	for k := 1; k < len(fr.Dates); k++ {
		pred := coef[0]
		for i, id := range ids {
			pred += coef[i+1] * fr.Returns[id][k]
		}
		values[k] = values[k-1] * (1 + pred)
	}
	return values, r2, coef, nil
}

// ols solves the normal equations XᵀX·β = Xᵀy by Gaussian elimination with
// partial pivoting; fine for the handful of regressors used here.
func ols(X [][]float64, y []float64) ([]float64, error) {
	k := len(X[0])
	a := make([][]float64, k) // augmented [XᵀX | Xᵀy]
	for i := range a {
		a[i] = make([]float64, k+1)
	}
	for r, row := range X {
		for i := range k {
			for j := range k {
				a[i][j] += row[i] * row[j]
			}
			a[i][k] += row[i] * y[r]
		}
	}
	for col := range k {
		pivot := col
		for r := col + 1; r < k; r++ {
			if math.Abs(a[r][col]) > math.Abs(a[pivot][col]) {
				pivot = r
			}
		}
		a[col], a[pivot] = a[pivot], a[col]
		if math.Abs(a[col][col]) < 1e-12 {
			return nil, fmt.Errorf("matrice singulière (régresseurs colinéaires ?)")
		}
		for r := range k {
			if r == col {
				continue
			}
			f := a[r][col] / a[col][col]
			for j := col; j <= k; j++ {
				a[r][j] -= f * a[col][j]
			}
		}
	}
	coef := make([]float64, k)
	for i := range k {
		coef[i] = a[i][k] / a[i][i]
	}
	return coef, nil
}
