package web

import (
	"fmt"
	"slices"

	"github.com/bpineau/pofo/pkg/decumul"
	"github.com/bpineau/pofo/pkg/scenario"
)

// fanSeed is the seed every wealth fan is simulated with, and therefore the
// one /api/path replays a sample of it from.
const fanSeed = 7

// FutureYear is one row of an opened future (decumul.LedgerYear for the
// page): every amount in real euros of year 0, Return a FRACTION.
type FutureYear struct {
	Year       int     `json:"year"`       // plan year, 0-based
	Age        int     `json:"age"`        // the first member's age that year
	Return     float64 `json:"ret"`        // the growth sleeve's real return
	Open       float64 `json:"open"`       // wealth at the start of the year, buffer included
	Income     float64 `json:"income"`     // pension and side income received
	Spent      float64 `json:"spent"`      // net delivered by the portfolio
	FromBuffer float64 `json:"fromBuffer"` // the part of Spent drawn from the cash buffer
	Sold       float64 `json:"sold"`       // gross sales of the growth sleeve, tax included
	Tax        float64 `json:"tax"`        // tax paid on those sales
	Buffer     float64 `json:"buffer"`     // the cash buffer at the end of the year
	Close      float64 `json:"close"`      // wealth at the end of the year
	Cut        bool    `json:"cut"`        // delivered below the planned standard
	Broke      bool    `json:"broke"`      // at or after the year the money ran out
}

// FutureResult is one simulated future opened year by year: the sample a
// reader clicked on a wealth fan, replayed from its own draw (Plan.DrawPath)
// through the kernel's ledger (Plan.RunLedger). Note is set, and Years empty,
// when the request names no fan or no path of it.
type FutureResult struct {
	Note     string       `json:"note,omitempty"` // why nothing was opened; "" otherwise
	Model    string       `json:"model"`          // the fan's return model
	Index    int          `json:"index"`          // the path index replayed
	Ruined   bool         `json:"ruined"`         // the money ran out
	RuinYear int          `json:"ruinYear"`       // the first unfunded year; -1 when none
	Need     float64      `json:"need"`           // the planned net spending per year, euros
	Years    []FutureYear `json:"years"`          // one row per plan year
}

// fanPlan is the plan a fan of the named model is simulated on, base (the
// request's pr.plan()) on that model's source and the annual kernel: the one
// construction Paths and OpenFuture share, so a replayed sample is the path
// the chart drew, bit for bit.
func fanPlan(base decumul.Plan, sources []namedSource, name string) (decumul.Plan, bool) {
	ns, ok := pickModel(sources, name)
	if !ok {
		return decumul.Plan{}, false
	}
	p := base
	p.Monthly = false
	p.Source = ns.source
	return p, true
}

// OpenFuture replays sample PathIndex of the PathModel fan and returns its
// ledger. The path index is the key the fan handed out (Fan.Samples), never a
// rank: ranks tie wherever several futures end at zero.
func OpenFuture(pr Params, panel *scenario.Panel) FutureResult {
	if pr.NPaths == 0 {
		pr.NPaths = 2000
	}
	res := FutureResult{Model: pr.PathModel, Index: pr.PathIndex, RuinYear: -1, Need: pr.NeedAnnual}
	if !slices.Contains(fanModels, pr.PathModel) {
		res.Note = fmt.Sprintf("no wealth fan is called %q", pr.PathModel)
		return res
	}
	if pr.PathIndex < 0 || pr.PathIndex >= pr.NPaths {
		res.Note = fmt.Sprintf("path %d is not one of the %d simulated", pr.PathIndex, pr.NPaths)
		return res
	}
	pl, ok := fanPlan(pr.plan(), modelSources(pr, panel), pr.PathModel)
	if !ok {
		res.Note = fmt.Sprintf("the %s model is not available here", pr.PathModel)
		return res
	}
	seq, lives := pl.DrawPath(fanSeed, pr.PathIndex)
	path, led := pl.RunLedger(seq, lives)
	res.Ruined, res.RuinYear = path.Ruined, path.RuinYear
	age := int(pr.age())
	res.Years = make([]FutureYear, len(led))
	for i, y := range led {
		res.Years[i] = FutureYear{
			Year: y.Year, Age: age + y.Year, Return: y.Return, Open: y.Open, Income: y.Income,
			Spent: y.Spent, FromBuffer: y.FromBuffer, Sold: y.Sold, Tax: y.Tax,
			Buffer: y.Buffer, Close: y.Close, Cut: y.Cut, Broke: y.Broke,
		}
	}
	return res
}
