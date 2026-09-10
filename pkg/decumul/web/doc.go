// Package web is a thin embedded HTTP UI for pkg/decumul: it serves a
// single page of sliders and, on each change, runs the Monte-Carlo in Go and
// returns chart SVGs and summary cards as JSON. The engine stays in Go; the
// browser only renders. Handler returns a ready-to-mount http.Handler.
//
// With a nil panel it serves the parametric playground (returns from
// mu/sigma/df sliders). With a historical panel it also offers the
// bootstrap and historical-cohort models and live per-holding re-weighting;
// a panel shorter than two years is treated as absent (Fit.Valid,
// minPanelMonths) so a degenerate fit can never seed a doom model.
//
// Three options tune the chrome, all surfaced to the front end through
// /api/meta: WithNav (cross-navigation to sibling surfaces),
// WithSourceLabel (which market this mount runs on, shown as a provenance
// pill in the top bar) and WithPicker (the in-drawer portfolio loader, whose
// example list, catalog URL and comparison-report URL the caller supplies,
// since only the embedding server knows those mounts and endpoints). The
// loader fills the drawer where no portfolio is bound, and folds away under
// the allocation bar where one is, so a loaded mount is not a dead end.
// Without a picker the page states in one line how to bind a portfolio from
// the command line.
//
// Two further options describe the DEPLOYMENT rather than the chrome, and are
// off unless a mount is told it is a public host: WithIndexNowKey publishes the
// ownership key file at the root, and WithBeaconToken puts the cookieless
// Cloudflare Web Analytics tag (webui.Beacon) on every HTML page this handler
// serves, the book editions mounted under it included. Empty, neither leaves a
// trace: the pages are byte for byte what they are without them.
//
// Spending policies are exclusive in the rail. The kernel resolves clashes
// by a fixed precedence (see the package comment of pkg/decumul), which is
// invisible to a reader ticking two boxes, so claiming one policy clears the
// controls of the others; a shared URL is left as it arrived, only dimmed,
// so old links keep reproducing the run their sender saw.
//
// The tax book is three controls in the Taxes group, over the kernel's
// per-envelope model (decumul.Envelope). GainFrac is the share of today's
// capital that is unrealised GAIN rather than cost basis, which is what the
// rate is charged on: it defaults to 50 %, because a cost basis equal to the
// whole capital is an assumption no long accumulation matches and it flatters
// the sustainable withdrawal rate by 0.30 point. PEACapital and AVCapital, in
// the rail's "envelopes" disclosure, name how much of the invested capital
// sits in each French wrapper, the remainder being the taxable account;
// Params.envelopes then builds the ordered pockets the kernel drains CTO
// first, PEA next, assurance-vie last (that one carrying the couple's annual
// allowance). Both amounts at zero leaves Plan.Envelopes nil, i.e. the
// historical single sleeve at Params.TaxRate, so a plan that names no wrapper
// is computed exactly as it was before the controls existed. The structure is
// worth 0.015 point of withdrawal rate and the drain order 0.03, against 0.12
// for the rate's calibration and 0.30 for the gain fraction: see
// docs/fire-envelopes-tax-model-design.md, which also carries the calibration
// recipe (a gain-weighted rate, a capital-weighted gain fraction) that the
// help texts point at.
//
// Beyond the model strip and the sweeps, the analysis endpoints serve the
// sequence-risk decomposition (/api/decade), the deterministic replay of
// infamous historical vintages through the user's plan (/api/vintages), the
// median funding-mix layers (/api/income), the lived-spending fan, the
// mortality lifecycle and the planning curves. /api/lifecycle runs the
// stochastic-lifetime kernel (decumul.Lifetime, a couple of the user's age):
// the death is drawn inside every path, so ruin there means broke while
// alive, counted, and the terminal-wealth histogram is the estate at the
// household's own end. The same draws are replayed with mortality off to
// carry the headline "ignoring mortality" figure beside it.
//
// The annuity block (AnnuityShare/AnnuityYear/AnnuityLoad, see annuity.go)
// belongs to that same endpoint and to no other. An annuity is longevity
// insurance: under the fixed horizon every other view runs, the household is
// certain to reach the end, so a lifelong income is simply paid for longer
// than it was priced for. /api/lifecycle therefore attaches decumul.Annuity
// beside the Lifetime and replays a THIRD twin on the same draws with the
// purchase removed, which is the before-and-after readout its last three cards
// carry: the risk of outliving the money, the estate that pays for removing
// it, and the payout the quote actually offers against the plan's own
// withdrawal rate. Everything else on the page ignores the block.
//
// Every simulation endpoint is bounded (bounds.go), because the page may be
// served to anonymous visitors from a small machine: the posted body is
// capped at maxBodyBytes before it is decoded, nPaths and every year-like
// field are clamped (maxPaths is the top of the page's own slider), and the
// computations queue behind simParallel slots, so no request can inflate a
// simulation past what the page itself can ask for, and a burst cannot pile
// them up. A request whose client gave up while waiting is refused with 503
// rather than computed for nobody. Coherence is bounded there too: the
// envelope amounts are clamped to the invested capital and a book whose
// pockets add up to more than the sleeve they are carved from is refused with
// a message naming the sum, since clamping it would silently simulate another
// household. The page cannot post one, its two amount sliders stopping at the
// room the other leaves.
package web
