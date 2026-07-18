package simgen

import (
	"math"
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
)

func pt(y int, m time.Month, d int, v float64) marketdata.Point {
	return marketdata.Point{Date: time.Date(y, m, d, 0, 0, 0, 0, time.UTC), Close: v}
}

func TestAnchorShapeHitsAnchorsAndFollowsShape(t *testing.T) {
	// Monthly total-return anchors: +10% in January, -10% in February.
	anchors := []marketdata.Point{
		pt(2020, 1, 1, 100),
		pt(2020, 2, 1, 110),
		pt(2020, 3, 1, 99),
	}
	// Daily price shape with its own arbitrary wiggle and a drift that does
	// NOT match the anchors (as a dividend-less price index would).
	shape := []marketdata.Point{
		pt(2020, 1, 2, 50), pt(2020, 1, 10, 52), pt(2020, 1, 20, 47), pt(2020, 1, 31, 51),
		pt(2020, 2, 3, 50), pt(2020, 2, 14, 46), pt(2020, 2, 28, 48),
		pt(2020, 3, 2, 49),
	}

	got := anchorShape(anchors, shape)

	// One point per shape date inside the anchor span, anchors hit exactly at
	// the first shape date on/after each anchor date.
	if len(got) != 8 {
		t.Fatalf("len = %d, want 8: %+v", len(got), got)
	}
	if got[0].Close != 100 || !got[0].Date.Equal(shape[0].Date) {
		t.Errorf("first = %+v, want the first anchor level 100 at the first shape date", got[0])
	}
	if got[4].Close != 110 || !got[4].Date.Equal(pt(2020, 2, 3, 0).Date) {
		t.Errorf("February anchor: got %+v, want 110 at 2020-02-03", got[4])
	}
	if got[7].Close != 99 || !got[7].Date.Equal(pt(2020, 3, 2, 0).Date) {
		t.Errorf("March anchor: got %+v, want 99 at 2020-03-02", got[7])
	}
	// Within a month, day-to-day moves follow the shape times one constant
	// residual factor (the dividend accrual spread evenly per step).
	resid := func(i int) float64 {
		return (got[i].Close / got[i-1].Close) / (shape[i].Close / shape[i-1].Close)
	}
	for i := 2; i <= 4; i++ { // January segment: steps 1..4 share one residual
		if math.Abs(resid(i)-resid(1)) > 1e-12 {
			t.Errorf("January residual not constant: step %d %.15f vs %.15f", i, resid(i), resid(1))
		}
	}
	for i := 6; i <= 7; i++ { // February segment
		if math.Abs(resid(i)-resid(5)) > 1e-12 {
			t.Errorf("February residual not constant: step %d %.15f vs %.15f", i, resid(i), resid(5))
		}
	}
}

func TestMSCIWorldBlendsDailyShape(t *testing.T) {
	// 400 monthly net-TR anchors from 1969-12, steady +1%/month.
	net := &marketdata.Series{Symbol: "MSCIWORLD-USD", Name: "MSCI World net TR"}
	d := time.Date(1969, 12, 1, 0, 0, 0, 0, time.UTC)
	v := 100.0
	for range 400 { // through 2003-03
		net.Points = append(net.Points, marketdata.Point{Date: d, Close: v})
		d = d.AddDate(0, 1, 0)
		v *= 1.01
	}
	// Flat daily price shape (dividend-less) covering 1990 through the net end.
	shape := &marketdata.Series{Symbol: "^990100-USD-STRD"}
	sd := time.Date(1990, 1, 2, 0, 0, 0, 0, time.UTC)
	for i := range 4900 {
		shape.Points = append(shape.Points, marketdata.Point{Date: sd.AddDate(0, 0, i), Close: 50})
	}
	f := fakeFetcher{"MSCIWORLD-USD": net, "^990100-USD-STRD": shape}
	b := msciWorld(0, func(Fetcher, time.Time) (*marketdata.Series, error) {
		t.Fatal("fallback must not run")
		return nil, nil
	})

	got, err := b(f, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	// Denser than monthly: daily points in the shape era on top of the
	// monthly prefix.
	if len(got.Points) < 4000 {
		t.Fatalf("points = %d, want daily density in the shape era", len(got.Points))
	}
	// The monthly prefix (1969-12 -> 1990) is preserved as-is in front.
	if p := got.Points[0]; !p.Date.Equal(net.Points[0].Date) || math.Abs(p.Close-100) > 1e-9 {
		t.Errorf("first = %+v, want the 1969-12 anchor at 100", p)
	}
	// A shape-era anchor keeps its exact net-TR level (fee 0 here): March
	// 1990 is 243 months after 1969-12. alignMonthEnd snaps the monthly anchor
	// onto the shape's last day of the month, so the level lands on 1990-03-31,
	// not the anchor's own first-of-month label.
	want := 100 * math.Pow(1.01, 243)
	at := time.Date(1990, 3, 31, 0, 0, 0, 0, time.UTC)
	rate, _, ok := got.At(at)
	if !ok || math.Abs(rate-want)/want > 1e-9 {
		t.Errorf("1990-03 level = %v, want %v (net anchor preserved)", rate, want)
	}
}

func TestShapedSeriesTruncatedShapeKeepsAnchors(t *testing.T) {
	// A shape that stops short of the anchors' end must not silently drop
	// the recent anchors. With no second anchor boundary inside the shape
	// there is nothing to blend: the series stays monthly.
	anchors := &marketdata.Series{Points: []marketdata.Point{
		pt(2020, 1, 1, 100), pt(2020, 2, 1, 110), pt(2020, 3, 1, 99),
	}}
	shape := &marketdata.Series{Points: []marketdata.Point{
		pt(2020, 1, 2, 50), pt(2020, 1, 20, 51),
	}}
	got := shapedSeries(anchors, shape)
	if len(got.Points) != 3 || got.Points[2].Close != 99 {
		t.Errorf("truncated shape: got %+v, want the anchors unchanged", got.Points)
	}
}

func TestShapedSeriesKeepsAnchorsAfterShapeEnd(t *testing.T) {
	// A shape covering only the early era (like the daily Treasury yields,
	// which only matter before the real fund starts) blends that era and
	// appends the remaining anchors at their own cadence, continuing the
	// same index level seamlessly.
	anchors := &marketdata.Series{Points: []marketdata.Point{
		pt(2020, 1, 1, 100), pt(2020, 2, 1, 110), pt(2020, 3, 1, 99), pt(2020, 4, 1, 105),
	}}
	shape := &marketdata.Series{Points: []marketdata.Point{
		pt(2020, 1, 2, 50), pt(2020, 1, 10, 52), pt(2020, 1, 20, 47), pt(2020, 2, 3, 50),
	}}
	got := shapedSeries(anchors, shape)
	// The January anchor spliced back in front + blended January (4 shape
	// dates) + the March and April anchors.
	if len(got.Points) != 7 {
		t.Fatalf("len = %d, want 7: %+v", len(got.Points), got.Points)
	}
	if p := got.Points[0]; p.Close != 100 || !p.Date.Equal(pt(2020, 1, 1, 0).Date) {
		t.Errorf("front = %+v, want the January anchor 100 at 2020-01-01", p)
	}
	if p := got.Points[4]; p.Close != 110 || !p.Date.Equal(pt(2020, 2, 3, 0).Date) {
		t.Errorf("blend end = %+v, want the February anchor level 110 at 2020-02-03", p)
	}
	if p := got.Points[5]; p.Close != 99 || !p.Date.Equal(pt(2020, 3, 1, 0).Date) {
		t.Errorf("first kept anchor = %+v, want 99 at 2020-03-01", p)
	}
	if p := got.Points[6]; p.Close != 105 {
		t.Errorf("last kept anchor = %+v, want 105", p)
	}
}

func TestAnchorShapeDegenerate(t *testing.T) {
	anchors := []marketdata.Point{pt(2020, 1, 1, 100), pt(2020, 2, 1, 110)}
	// No shape at all: the anchors come back unchanged.
	if got := anchorShape(anchors, nil); len(got) != 2 || got[0].Close != 100 {
		t.Errorf("empty shape: got %+v, want the anchors unchanged", got)
	}
	// Shape entirely outside the anchor span: same.
	outside := []marketdata.Point{pt(2021, 5, 3, 42), pt(2021, 5, 4, 43)}
	if got := anchorShape(anchors, outside); len(got) != 2 || got[1].Close != 110 {
		t.Errorf("disjoint shape: got %+v, want the anchors unchanged", got)
	}
}

func TestAnchorShapeSkipsAnchorsBeforeShape(t *testing.T) {
	// Anchors 1969->; shape only from mid-1970: the output starts at the
	// first anchor the shape covers, earlier anchors are the caller's
	// business (ExtendBack prepends them seamlessly).
	anchors := []marketdata.Point{
		pt(1969, 12, 1, 100), pt(1970, 1, 1, 101), pt(1970, 2, 1, 103), pt(1970, 3, 1, 99),
	}
	shape := []marketdata.Point{
		pt(1970, 2, 2, 10), pt(1970, 2, 16, 11), pt(1970, 3, 2, 10.5),
	}

	got := anchorShape(anchors, shape)

	if len(got) != 3 {
		t.Fatalf("len = %d, want 3 (shape span only): %+v", len(got), got)
	}
	if got[0].Close != 103 || !got[0].Date.Equal(shape[0].Date) {
		t.Errorf("first = %+v, want the 1970-02 anchor level 103 at 1970-02-02", got[0])
	}
	if got[2].Close != 99 {
		t.Errorf("last = %+v, want the 1970-03 anchor level 99", got[2])
	}
}
