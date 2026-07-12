// FIRE explorer front-end: a grouped control rail on the left, a long
// analysis column on the right, and a sticky plan bar echoing the live
// verdict. All charts are server-rendered SVG; this file only wires state,
// fetches and DOM.

// ---------------------------------------------------------------------------
// Control definitions. r() is a slider, c() a checkbox, chips() a preset row.
// Every control carries a plain-language data-help hover.
// ---------------------------------------------------------------------------
const r = (key, label, min, max, step, def, unit, help) =>
  ({kind: "range", key, label, min, max, step, def, unit, help});
const c = (key, label, help) => ({kind: "check", key, label, help});
const chips = (label, help, items) => ({kind: "chips", label, help, items});

const GROUPS = [
  {title: "Your situation", items: [
    r("capital", "Deployed capital", 800000, 4000000, 10000, 1800000, "eur",
      "Liquid capital deployed for the retirement, excluding your home and the emergency fund."),
    r("age", "Age at retirement", 40, 70, 1, 52, "int",
      "Age in year 0. Drives the mortality view (section 05): being broke at 61 and at 92 are different life events."),
    r("years", "Horizon (years)", 20, 60, 1, 45, "int",
      "Plan past your life expectancy: ruin rises steeply with the horizon (40→50y nearly doubles it)."),
    r("needAnnual", "Net spending /yr", 24000, 84000, 1000, 60000, "eur",
      "Real (inflation-indexed) net-of-tax household spending. 60 k€/yr = 5 000 €/month."),
  ]},
  {title: "Pension & side income", items: [
    chips("Pension scenario",
      "Three pension levels: a politically-stressed 1 000 €/m, the acquired-rights central ~1 700 €/m, and the official-simulator ~2 250 €/m (net real).",
      [["Stress 12k", {pensionAnnual: 12000}],
       ["Central 20.4k", {pensionAnnual: 20400}],
       ["Official 27k", {pensionAnnual: 27000}]]),
    r("pensionAnnual", "Pension /yr (net real)", 0, 36000, 600, 12000, "eur",
      "Net real pension once it starts. Simulations show this is the plan's second-biggest sensitivity."),
    r("pensionYear", "Pension starts in year", 5, 25, 1, 15, "int",
      "Years from retirement to the pension. Retiring at 52 with a pension at 67 = year 15."),
    r("sideAnnual", "Side income /yr", 0, 40000, 1000, 0, "eur",
      "Temporary net real income (rental, activity…) subtracted from the need while it lasts. Income covering the early years is the best sequence-risk insurance there is."),
    r("sideUntilYear", "Side income until year", 0, 20, 1, 0, "int",
      "The side income runs from year 0 up to (excluding) this year."),
  ]},
  {title: "Spending policy", items: [
    r("flexCut", "Cut in downturns (0 = fixed rule)", 0, 0.40, 0.05, 0, "pct",
      "Reversible spending cut while the portfolio drawdown exceeds 20%. The single most powerful lever: 15% roughly halves ruin. Section 04 shows its lived cost."),
    r("wrTrigger", "Also cut above this WR (0 = off)", 0, 0.06, 0.002, 0, "pct",
      "Second trigger from the written rules: cut whenever the current withdrawal rate (spend / portfolio) exceeds this, e.g. 3.6%."),
    c("guardrails", "Guyton-Klinger guardrails (replaces the cut)",
      "Adjust spending ±10% whenever the current withdrawal rate leaves a ±20% band around the initial rate. It trades ruin risk for LIFESTYLE risk: cuts repeat while the portfolio keeps underperforming, so in the bad tail income decays step after step (section 04 shows the lived cost; the frontier in section 06 shows the trade). The guardrails ruin figure is therefore not comparable to the fixed rule's: it means 'even cutting without limit was not enough'. Set a floor below to bound the descent."),
    r("gkFloor", "Guardrails floor (% of initial, 0 = none)", 0, 0.9, 0.05, 0, "pct",
      "The incompressible standard: guardrails cuts never push spending below this share of the initial level. Bounding the descent re-creates some ruin (the floor itself can prove unsustainable), which is the honest trade; 70-80% is a common planning value. Only applies while guardrails are on."),
    c("ratchet", "Ratchet lifestyle up when rich",
      "Only-up rule from the written rules: +10% of the base spend when real wealth exceeds 120% of the start, at most every 2 years, capped at 120% of the base, only while the current rate is below 2.2%."),
    r("spendDrift", "Real spending drift /yr", -0.01, 0.02, 0.001, 0, "pct",
      "Structural real drift of the need: health insurance and care costs drift upward faster than inflation (+0.3-0.5%/yr is a common planning value)."),
    c("smile", "Retirement smile (down, plateau, up late)",
      "Blanchett's observed shape: real spending drifts down through the go-go years, plateaus, then climbs back with late-life health costs."),
    r("percent", "Spend % of portfolio (VPW, 0 = off)", 0, 0.08, 0.005, 0, "pct",
      "Percentage-of-portfolio (VPW) rule: each year spend this share of the current portfolio instead of a fixed amount. It never runs out, but the standard of living swings with the market. The other end of the decumulation frontier; overrides the fixed need and the flex/guardrails/ratchet rules."),
    r("annuityShare", "Annuitise % of capital", 0, 0.5, 0.05, 0, "pct",
      "Spend this share of capital on a joint-life, inflation-linked annuity (1% real rate, 10% insurer load): a guaranteed lifelong income floor that hedges longevity. It converts growth assets into lower guaranteed income, so headline ruin (failing the FULL need) can rise even as the worst late-life outcomes improve; its value is the floor, not the average."),
  ]},
  {title: "Market model", items: [
    r("mu", "Real growth return", 0.01, 0.12, 0.005, 0.05, "pct",
      "Arithmetic mean annual real return of the growth sleeve. Geometric ≈ μ − σ²/2, so 5%/11% ≈ 4.4% real compounding."),
    r("sigma", "Volatility (long-horizon)", 0.06, 0.20, 0.005, 0.11, "pct",
      "Long-horizon annual volatility (variance-ratio-consistent), lower than the 1-year headline vol. Vol matters almost as much as return: 10→12% nearly doubles ruin."),
    r("df", "Tail df (low = fat)", 3, 30, 1, 5, "int",
      "How much more often EXTREME years happen than a bell curve allows, at the same volatility. At df 5 a catastrophic year (say -30% real, a 3-sigma event) is roughly ten times more likely than at df 30 (≈ normal law); ordinary years barely change. Low df = crash-prone world. In portfolio mode it is seeded from the holdings' monthly kurtosis."),
    c("regime", "Sequence-risk stress (cluster bad years)",
      "NOT a replay of a bad historical sequence: a synthetic model where bad years cluster into multi-year bears (about 19% of years, runs of ~3) at the SAME long-run mean, so only the ORDERING is stressed. Roughly one simulated future in five starts inside a bear. The strip's Sequence-stress column always shows it; this checkbox makes it the active model for the detail charts. The real named worst cases are section 02."),
    c("conservative", "Broad-sample prior (override the fit)",
      "Rewrites the three μ/σ/df sliders with cautious world-equity assumptions (~3.5% real geometric, fat tails): what broad century-long samples suggest, not this fund's history. Sliders only: the data-driven columns (Historical, Block bootstrap, Broad-sample) never read the sliders and are untouched."),
    c("capeAdjust", "Anchor return to today's valuation (CAPE)",
      "Replaces ONLY the central mean with the CAPE-implied estimate (1/CAPE): today's rich valuation compresses the next decade, the years that make or break a retirement. Volatility and tails keep their fitted values; the data-driven columns are untouched. Overrides the return slider and the broad-sample blend."),
    c("glidepath", "Rising-equity glidepath (bond tent)",
      "Hold less equity at retirement (30%) gliding to 75% later, blending in bonds (~1.5% real). It cuts sequence risk (the danger years are the least equity-heavy) but gives up return; with a wide equity-bond gap the drag can outweigh the protection, so compare the ruin both ways. Applies to the central case."),
    c("monthly", "Monthly withdrawals (salary-like)",
      "Step the kernel monthly instead of annually: withdrawals, drawdown checks and the bucket rule run every month. Only the plan-detail and buffer sections (§07-§08) honour it; the model strip and the analysis sections always compare annual kernels, for speed and comparability. The effect is small by design (a refinement, not a lever)."),
  ]},
  {title: "Cash buffer", items: [
    r("bufferYears", "Buffer (years of spending)", 0, 10, 1, 3, "int",
      "Low-volatility pocket (cash + short euro linkers) drained when the portfolio is down >10%. Statistically the arbitrage is flat past 2-3 years; its value is behavioural and inflation matching."),
    r("bufferReturn", "Buffer real return", -0.01, 0.05, 0.005, 0.005, "pct",
      "Real return of the buffer: ~0% for inflation-linked, negative for pure cash."),
    r("bufferStopYear", "Buffer refill stops in year (0 = never)", 0, 20, 1, 0, "int",
      "The melting buffer: stop refilling after the sequence-risk window (e.g. year 8) and let it run down — a bond-tent glidepath."),
  ]},
  {title: "Taxes & envelopes", items: [
    r("taxRate", "CTO flat tax on gains", 0, 0.35, 0.01, 0.314, "pct",
      "French flat tax (PFU + CEHR) on the gain share of every CTO sale. The effective rate starts low and drifts up as unrealised gains compound."),
    r("peaCapital", "PEA envelope", 0, 400000, 5000, 0, "eur",
      "Capital held in a PEA (>5y): withdrawals only pay 17.2% social levies on gains. Drained after the CTO."),
    r("avCapital", "Assurance-vie envelope", 0, 500000, 5000, 0, "eur",
      "Capital held in assurance-vie (>8y): 9 200 €/yr of realised gains tax-free for a couple, then 24.7%. Drained last."),
    r("gainFrac", "Embedded gain at start", 0, 0.9, 0.05, 0, "pct",
      "Unrealised gain share of the portfolio on day one. High embedded gains make every early sale taxable — set >0 for a portfolio carrying years of appreciation."),
  ]},
  {title: "Simulation", items: [
    r("nPaths", "Simulated paths", 500, 5000, 500, 2000, "int",
      "Monte-Carlo paths per model. More paths = smoother figures, slower updates."),
  ]},
];

const FMT = {
  pct: v => (v * 100).toFixed(1).replace(/\.0$/, "") + "%",
  eur: v => Math.round(v).toLocaleString("fr-FR") + " €",
  int: v => String(Math.round(v)),
};
const UNIT = {}, INTKEYS = [];
for (const g of GROUPS) for (const it of g.items) if (it.kind === "range") {
  UNIT[it.key] = it.unit;
  if (it.unit === "int") INTKEYS.push(it.key);
}
const fmtVal = (k, v) => (FMT[UNIT[k] || "int"])(v);
// Series palette, mirrors chart.PaletteColor so page and SVG stay consistent.
const PAL = ["#1B95B4","#BE4E6C","#7A5FC0","#C4820F","#9A6FD0","#2F9463","#C24E86","#2FA0B8"];
const esc = s => (s || "").replace(/&/g, "&amp;").replace(/"/g, "&quot;").replace(/</g, "&lt;").replace(/>/g, "&gt;");

// portfolio-mode state, set once /api/meta resolves.
let weights = null, labels = [], hasPanel = false, lastFitW = null;

// ---------------------------------------------------------------------------
// Build the rail.
// ---------------------------------------------------------------------------
const form = document.getElementById("controls");
const state = {};
const checkEls = {};

function renderRail() {
  GROUPS.forEach((g, gi) => {
    const box = document.createElement("div");
    box.className = "group";
    box.innerHTML = `<div class="group-h">${g.title}</div>`;
    // The plan-defining sliders (first group) get a ruler of ticks.
    for (const it of g.items) box.appendChild(buildControl(it, gi === 0));
    form.appendChild(box);
  });
}

function buildControl(it, ruler) {
  if (it.kind === "check") {
    state[it.key] = false;
    const d = document.createElement("label");
    d.className = "ctl chk";
    if (it.help) d.setAttribute("data-help", it.help);
    d.innerHTML = `<input type="checkbox" id="c_${it.key}"> <span>${it.label}</span>`;
    const input = d.querySelector("input");
    checkEls[it.key] = input;
    input.addEventListener("change", e => {
      state[it.key] = e.target.checked;
      if (it.key === "conservative") applyConservative();
      schedule();
    });
    return d;
  }
  if (it.kind === "chips") {
    const d = document.createElement("div");
    d.className = "ctl chips";
    if (it.help) d.setAttribute("data-help", it.help);
    d.innerHTML = `<span class="lab"><span>${it.label}</span></span>`;
    const row = document.createElement("div");
    row.className = "chiprow";
    for (const [text, sets] of it.items) {
      const b = document.createElement("button");
      b.type = "button";
      b.textContent = text;
      b.addEventListener("click", () => {
        for (const [k, v] of Object.entries(sets)) setSliderVal(k, v);
        schedule();
      });
      row.appendChild(b);
    }
    d.appendChild(row);
    return d;
  }
  // range slider
  state[it.key] = it.def;
  const d = document.createElement("label");
  d.className = "ctl";
  if (it.help) d.setAttribute("data-help", it.help);
  d.innerHTML = `<span class="lab"><span>${it.label}</span><span class="val" id="v_${it.key}">${fmtVal(it.key, it.def)}</span></span>
    <input type="range" min="${it.min}" max="${it.max}" step="${it.step}" value="${it.def}" id="s_${it.key}">` +
    (ruler ? `<div class="ticks"></div>` : ``);
  const input = d.querySelector("input");
  paintFill(input);
  input.addEventListener("input", e => {
    state[it.key] = parseFloat(e.target.value);
    paintFill(e.target);
    refreshVal(it.key);
    schedule();
  });
  return d;
}

// paintFill keeps the slider track's accent fill in sync with its value
// (the theme's track is a two-stop gradient split at --fill).
function paintFill(s) {
  const min = parseFloat(s.min), max = parseFloat(s.max);
  const p = max > min ? (100 * (parseFloat(s.value) - min)) / (max - min) : 0;
  s.style.setProperty("--fill", p.toFixed(2) + "%");
}

// refreshVal renders a slider's live value, with contextual extras (ages).
function refreshVal(k) {
  const el = document.getElementById("v_" + k);
  if (!el) return;
  let text = fmtVal(k, state[k]);
  if (k === "pensionYear" || k === "years") text += ` (age ${Math.round(state.age + state[k])})`;
  el.textContent = text;
}
function refreshAges() { refreshVal("pensionYear"); refreshVal("years"); }

function setSliderVal(k, v) {
  state[k] = v;
  const s = document.getElementById("s_" + k);
  if (s) { s.value = v; paintFill(s); refreshVal(k); }
  if (k === "age") refreshAges();
}

renderRail();
refreshAges();
document.getElementById("s_age").addEventListener("input", refreshAges);

// Broad-sample prior: override the fitted mu/sigma/df with cautious values
// matching the server's Broad-sample column, or restore the fit/defaults.
const PRIOR = {mu: 0.045, sigma: 0.13, df: 4};
const DEFAULTS = {mu: 0.05, sigma: 0.11, df: 5};
function applyReturns(src) { for (const k of ["mu", "sigma", "df"]) setSliderVal(k, src[k]); }
function applyConservative() {
  if (state.conservative) applyReturns(PRIOR);
  else if (hasPanel) lastFitW = null; // force a refit from the panel
  else applyReturns(DEFAULTS);
}

// ---------------------------------------------------------------------------
// Shareable scenarios: the whole state round-trips through the URL hash.
// ---------------------------------------------------------------------------
const CHECKKEYS = Object.keys(checkEls);
const shared = new URLSearchParams(location.hash.slice(1));
const sharedWeights = shared.has("w")
  ? shared.get("w").split(",").map(Number).filter(x => !isNaN(x)) : null;
function applySharedSliders() {
  for (const k of Object.keys(UNIT))
    if (shared.has(k)) { const v = parseFloat(shared.get(k)); if (!isNaN(v)) setSliderVal(k, v); }
}
applySharedSliders();
for (const k of CHECKKEYS) {
  if (shared.get(k) === "1") { state[k] = true; checkEls[k].checked = true; }
}
if (state.conservative) applyReturns(PRIOR);

function syncURL() {
  const p = new URLSearchParams();
  for (const k of Object.keys(UNIT)) p.set(k, state[k]);
  if (state.model) p.set("model", state.model);
  for (const k of CHECKKEYS) if (state[k]) p.set(k, "1");
  if (weights) p.set("w", weights.map(x => x.toFixed(4)).join(","));
  history.replaceState(null, "", "#" + p.toString());
}

// ---------------------------------------------------------------------------
// Scheduling: a fast lane for the live views, a slow lane for the two
// solver-heavy planning curves. A run id drops stale responses.
// ---------------------------------------------------------------------------
let timer = null, slowTimer = null, runId = 0;
function schedule() {
  clearTimeout(timer); timer = setTimeout(run, 200);
  clearTimeout(slowTimer); slowTimer = setTimeout(runSlow, 600);
}
const post = (url, body) =>
  fetch(url, {method: "POST", headers: {"Content-Type": "application/json"}, body: JSON.stringify(body)})
    .then(r => r.json());
// fresh(id) is true while no newer run started: renderers check it before
// touching the DOM so a slow older response never overwrites a newer one.
const fresh = id => id === runId;

function body() {
  const b = {...state};
  for (const k of INTKEYS) b[k] = Math.round(b[k]);
  b.targetRuin = (parseFloat(document.getElementById("targetRuin").value) || 5) / 100;
  if (weights) b.weights = weights;
  return b;
}

let run = async function() {
  runId++;
  const id = runId;
  // In portfolio mode the parametric models read mu/sigma, not the weights,
  // so a weight change re-fits mu/sigma from the panel before computing.
  if (weights && hasPanel && weightsChanged() && !state.conservative) {
    try {
      const f = await post("/api/fit", {weights});
      if (typeof f.mu === "number") setSliderVal("mu", f.mu);
      if (typeof f.sigma === "number") setSliderVal("sigma", f.sigma);
      if (typeof f.df === "number") setSliderVal("df", f.df);
    } catch (e) { /* keep the current mu/sigma on failure */ }
    lastFitW = weights.slice();
  }
  const b = body();
  renderModels(b, id);
  renderPaths(b, id);
  renderMarket(b, id);
  renderSolver(b, id);
  renderFrontier(b, id);
  renderPolicyFrontier(b, id);
  renderSensitivity(b, id);
  renderSpending(b, id);
  renderVintages(b, id);
  renderDecade(b, id);
  renderIncome(b, id);
  renderLifecycle(b, id);
  renderSim(b, id);
  updateCmd();
  syncURL();
};

// The top-bar command echo mirrors the live plan, terminal-style.
function updateCmd() {
  const el = document.getElementById("cmdEcho");
  if (!el) return;
  const rule = state.percent > 0 ? "vpw" : state.guardrails ? "guardrails" : state.flexCut > 0 ? "flex" : "fixed·real";
  const model = state.conservative ? "broad-sample" : state.regime ? "seq-stress" : state.capeAdjust ? "cape" : "student-t";
  const f = (k) => fmtVal(k, state[k]).replace(/\s?€/, "").replace(/\s/g, "");
  el.innerHTML =
    `<span class="flag">plan</span> <span class="flag">--capital</span> <span class="val">${f("capital")}€</span>` +
    ` <span class="flag">--spend</span> <span class="val">${f("needAnnual")}€</span>` +
    ` <span class="flag">--horizon</span> <span class="val">${state.years}y</span>` +
    ` <span class="flag">--rule</span> <span class="val">${rule}</span>` +
    ` <span class="flag">--model</span> <span class="val">${model}</span>`;
}

// The valuation strip: a cheap→rich scale with the median tick and today's
// marker, plus the implied real return, built from the CAPE snapshot.
function renderCape(cape) {
  const el = document.getElementById("capeStrip");
  if (!el || !cape || !cape.value) return;
  const pct = Math.max(2, Math.min(98, cape.percentile));
  const asOf = (cape.asOf || "").slice(0, 7);
  // A reading older than a year must never present itself as "now".
  const freshness = cape.stale
    ? `<span class="stale" data-help="The bundled Shiller series ends in ${asOf}: this is NOT today's valuation. Refresh it with 'make cape'.">as of ${asOf} · stale</span>`
    : `<span class="lbl">as of ${asOf}</span>`;
  el.innerHTML =
    `<div class="vbig"><span class="n">CAPE ${cape.value.toFixed(1)}</span>` +
    `<span class="lbl">${Math.round(cape.percentile)}th percentile since 1881</span>${freshness}</div>` +
    `<div class="vstrip" title="cheap on the left, rich on the right">` +
    `<span class="tick" style="left:50%"></span><span class="now" style="left:${pct}%"></span></div>` +
    `<div class="vstrip-labels"><span style="left:4%">cheap</span>` +
    (pct < 74 || pct > 82 ? `<span style="left:50%">median ${cape.median.toFixed(1)}</span>` : ``) +
    (pct < 90 ? `<span class="now" style="left:${pct}%">today</span>` : `<span class="now" style="right:2%;left:auto;transform:none">today</span>`) +
    (pct < 88 ? `<span style="left:96%">rich</span>` : ``) +
    `</div>` +
    `<div class="vbig"><span class="n">${(cape.impliedReal * 100).toFixed(1)}%</span>` +
    `<span class="lbl">implied 10y real return · 1/CAPE</span></div>` +
    `<div class="vnote">Rich valuations compress the first decade, the years that decide a retirement. Enable <b>anchor to CAPE</b> to plan on the implied return instead of the fitted average.</div>`;
}

// Settings drawer: fold the controls away to run the analysis full-width.
(function () {
  const t = document.getElementById("drawerToggle"), d = document.getElementById("drawer");
  if (!t || !d) return;
  t.addEventListener("click", () => {
    const open = d.hasAttribute("hidden");
    if (open) { d.removeAttribute("hidden"); } else { d.setAttribute("hidden", ""); }
    t.setAttribute("aria-expanded", String(open));
    t.textContent = open ? "parameters ▴" : "parameters ▾";
  });
})();

async function runSlow() {
  const id = runId;
  try {
    const r = await post("/api/curves", body());
    if (!fresh(id)) return;
    setSVG("horizonSvg", r.horizonSvg);
    setSVG("capitalSvg", r.capitalSvg);
  } catch (e) { /* keep the previous curves */ }
}

function weightsChanged() {
  if (!weights || !lastFitW || lastFitW.length !== weights.length) return true;
  return weights.some((w, i) => Math.abs(w - lastFitW[i]) > 1e-9);
}

function setSVG(elId, svg) {
  const el = document.getElementById(elId);
  if (el && svg) el.innerHTML = svg;
}
// cardsHTML renders summary cards; a help field becomes the hover, and a
// verdict-shaped value (vintage replays) is graded good/bad by its outcome.
const cardGrade = v =>
  v.startsWith("ruined") ? " bad" : v.startsWith("survived") ? " good" : "";
const cardsHTML = cards => (cards || [])
  .map(c => `<div class="card"${c.help ? ` data-help="${esc(c.help)}"` : ""}>` +
    `<div class="k">${esc(c.label)}</div><div class="v${cardGrade(c.value)}">${esc(c.value)}</div></div>`).join("");

// fmtW renders a wealth figure with a readable unit (M€ past the million).
const fmtW = v => Math.abs(v) >= 1e6 ? (v / 1e6).toFixed(2) + "M€" : Math.round(v / 1000) + "k€";

// ---------------------------------------------------------------------------
// Instant tooltip for any [data-help] element.
// ---------------------------------------------------------------------------
const tip = document.createElement("div");
tip.id = "tip";
document.body.appendChild(tip);
document.addEventListener("mouseover", e => {
  const el = e.target.closest("[data-help]");
  if (!el) return;
  // Over a crosshair-instrumented chart the data tooltip wins; the help
  // text stays reachable from the chart's frame border and title.
  const svg = e.target.closest("svg");
  if (svg && hoverData(svg)) return;
  tip.textContent = el.getAttribute("data-help");
  tip.style.display = "block";
});
document.addEventListener("mousemove", e => {
  if (tip.style.display !== "block") return;
  const pad = 14, w = tip.offsetWidth, h = tip.offsetHeight;
  let x = e.clientX + pad, y = e.clientY + pad;
  if (x + w > innerWidth) x = e.clientX - pad - w;
  if (y + h > innerHeight) y = e.clientY - pad - h;
  tip.style.left = x + "px";
  tip.style.top = y + "px";
});
document.addEventListener("mouseout", e => {
  if (e.target.closest("[data-help]")) tip.style.display = "none";
});

// ---------------------------------------------------------------------------
// Crosshair + tooltip for every chart carrying hover metadata (fans, lines,
// stacked areas). The server embeds each chart's data as an SVG <metadata>
// element; this layer maps the pointer back to the data, snaps a hairline to
// the nearest x, and lists every series' value at that x. Values are built
// with textContent (labels are untrusted data).
// ---------------------------------------------------------------------------
const xtip = document.createElement("div");
xtip.id = "xtip";
document.body.appendChild(xtip);
let xline = null; // the crosshair line element, moved between charts

function hoverData(svg) {
  if (svg.__hd !== undefined) return svg.__hd;
  const meta = svg.querySelector("metadata.hover");
  try { svg.__hd = meta ? JSON.parse(meta.textContent) : null; }
  catch (e) { svg.__hd = null; }
  return svg.__hd;
}

// fmtHV renders a tooltip value with unit-free smart precision (the chart's
// axis label carries the unit).
function fmtHV(v) {
  const a = Math.abs(v);
  if (a >= 1e6) return (v / 1e6).toFixed(2) + "M";
  if (a >= 1e4) return Math.round(v / 1e3) + "k";
  if (a >= 100) return v.toFixed(0);
  if (a >= 10) return v.toFixed(1);
  return v.toFixed(2);
}

function hideCrosshair() {
  xtip.style.display = "none";
  if (xline) { xline.remove(); xline = null; }
}

document.addEventListener("pointermove", e => {
  const svg = e.target.closest ? e.target.closest("svg") : null;
  const hd = svg ? hoverData(svg) : null;
  if (!hd || !hd.series || !hd.series.length) { hideCrosshair(); return; }

  // Pointer position in viewBox coordinates.
  const rect = svg.getBoundingClientRect();
  const vb = svg.viewBox.baseVal;
  const px = (e.clientX - rect.left) * vb.width / rect.width;
  if (px < hd.x0 - 10 || px > hd.x1 + 10) { hideCrosshair(); return; }
  const xdom = hd.xmin + (Math.min(Math.max(px, hd.x0), hd.x1) - hd.x0) / (hd.x1 - hd.x0) * (hd.xmax - hd.xmin);

  // Snap per series to its nearest x; indexed charts snap to the index.
  const rows = [];
  let snapX = null;
  for (const s of hd.series) {
    if (!s.ys || !s.ys.length) continue;
    let i, sx;
    if (s.xs && s.xs.length) {
      i = 0;
      for (let k = 1; k < s.xs.length; k++)
        if (Math.abs(s.xs[k] - xdom) < Math.abs(s.xs[i] - xdom)) i = k;
      sx = s.xs[i];
      // A series that ends before the pointer (truncated vintage) drops out
      // rather than repeating its last value.
      if (Math.abs(sx - xdom) > (hd.xmax - hd.xmin) / 8) continue;
    } else {
      i = Math.round(Math.min(Math.max(xdom, 0), s.ys.length - 1));
      if (i >= s.ys.length) continue;
      sx = i;
    }
    if (snapX === null) snapX = sx;
    rows.push({name: s.name, color: s.color, v: s.ys[i]});
  }
  if (!rows.length || snapX === null) { hideCrosshair(); return; }

  // Crosshair line at the snapped x, spanning the plot box.
  const cx = hd.x0 + (snapX - hd.xmin) / (hd.xmax - hd.xmin) * (hd.x1 - hd.x0);
  if (!xline || xline.ownerSVGElement !== svg) {
    if (xline) xline.remove();
    xline = document.createElementNS("http://www.w3.org/2000/svg", "line");
    xline.setAttribute("stroke", "#B4A991");
    xline.setAttribute("stroke-width", "1");
    xline.setAttribute("stroke-dasharray", "2 3");
    xline.setAttribute("pointer-events", "none");
    svg.appendChild(xline);
  }
  xline.setAttribute("x1", cx); xline.setAttribute("x2", cx);
  xline.setAttribute("y1", hd.y0); xline.setAttribute("y2", hd.y1);

  // Tooltip: header = x (+ y-axis unit), then one row per series.
  xtip.textContent = "";
  const head = document.createElement("div");
  head.className = "xh";
  const xv = Math.abs(snapX - Math.round(snapX)) < 1e-9 ? String(Math.round(snapX)) : snapX.toFixed(1);
  head.textContent = (hd.xlabel || "x") + " " + xv + (hd.ylabel ? " · " + hd.ylabel : "");
  xtip.appendChild(head);
  for (const r of rows) {
    const row = document.createElement("div");
    row.className = "xr";
    const key = document.createElement("i");
    if (r.color) key.style.background = r.color;
    const val = document.createElement("b");
    val.textContent = fmtHV(r.v);
    const name = document.createElement("span");
    name.textContent = r.name || "";
    row.appendChild(key); row.appendChild(val); row.appendChild(name);
    xtip.appendChild(row);
  }
  xtip.style.display = "block";
  const pad = 14, tw = xtip.offsetWidth, th = xtip.offsetHeight;
  let tx = e.clientX + pad, ty = e.clientY + pad;
  if (tx + tw > innerWidth) tx = e.clientX - pad - tw;
  if (ty + th > innerHeight) ty = e.clientY - pad - th;
  xtip.style.left = tx + "px";
  xtip.style.top = ty + "px";
});
document.addEventListener("pointerleave", hideCrosshair);

// ---------------------------------------------------------------------------
// Chart lightbox: click any chart to view it large over the page, click
// again (or press Escape) to come back.
// ---------------------------------------------------------------------------
const lightbox = document.createElement("div");
lightbox.id = "lightbox";
lightbox.hidden = true;
document.body.appendChild(lightbox);
function closeLightbox() {
  lightbox.hidden = true;
  lightbox.innerHTML = "";
  document.body.classList.remove("noscroll");
}
document.addEventListener("click", e => {
  if (!lightbox.hidden) { closeLightbox(); return; }
  const frame = e.target.closest(".chart-frame, .fan");
  if (!frame || !frame.querySelector("svg")) return;
  lightbox.innerHTML = frame.innerHTML;
  lightbox.hidden = false;
  document.body.classList.add("noscroll");
});
document.addEventListener("keydown", e => {
  if (e.key === "Escape" && !lightbox.hidden) closeLightbox();
});

// Ruin colour: the theme's risk ramp, green (safe) through amber to red,
// saturating at 30%. Rendered as small dots/beads next to mono figures, so
// the numbers themselves stay in ink.
function beadColor(r) {
  const x = Math.max(0, Math.min(r, 0.30)) / 0.30;
  const hue = x < 0.5 ? 152 - 236 * x : 34 - 60 * (x - 0.5);
  return `hsl(${hue.toFixed(0)},78%,44%)`;
}

// ---------------------------------------------------------------------------
// Hero strip + plan bar.
// ---------------------------------------------------------------------------
async function renderModels(b, id) {
  let r;
  try { r = await post("/api/models", b); } catch (e) { return; }
  if (!fresh(id)) return;
  document.getElementById("verdict").textContent = r.verdict || "";
  const conf = document.getElementById("confidence");
  const cap = s => s ? s[0] + s.slice(1).toLowerCase() : "";
  conf.textContent = r.confidence ? `Confidence: ${cap(r.confidence)}` : "";
  if (r.confNote) conf.setAttribute("data-help", r.confNote);
  else conf.removeAttribute("data-help");
  const ms = r.models || [];
  const central = ms.findIndex(m => m.name === "Student-t");
  renderReadout(central >= 0 ? ms[central].ruin : NaN, r.targetRuin || 0.05);
  const sel = i => (i === central ? ` class="sel"` : "");
  const cells = fn => ms.map((m, i) => `<td${sel(i)}>${fn(m)}</td>`).join("");
  const head = `<tr><th></th>${ms.map((m, i) =>
    `<th${sel(i)} data-help="${esc(m.help)}">${m.name}</th>`).join("")}</tr>`;
  // Risk is carried by a coloured dot per cell; the figures stay in ink.
  const ruinRow = `<tr><th data-help="Share of simulated retirements that run out of money, at your planned spend.">Ruin</th>` +
    cells(m => `<i class="dot" style="background:${beadColor(m.ruin)}"></i>${(m.ruin * 100).toFixed(1)}%`) + `</tr>`;
  const spendRow = `<tr><th data-help="The most you could spend per year and still keep ruin at your acceptable level, under this model.">Safe spend</th>` +
    cells(m => `${(m.safeSpend / 1000).toFixed(0)}k€<span class="sub">${(m.safeWR * 100).toFixed(m.safeWR < 0.01 ? 2 : 1)}%</span>`) + `</tr>`;
  const wealthRow = `<tr><th data-help="Median real wealth left at the end of the horizon, at your planned spend.">Median wealth</th>` +
    cells(m => fmtW(m.medianWealth)) + `</tr>`;
  document.getElementById("modelstrip").innerHTML =
    `<table class="modeltab"><thead>${head}</thead><tbody>${ruinRow}${spendRow}${wealthRow}</tbody></table>`;

  // Plan bar echo: verdict condensed + one ruin bead per model.
  document.getElementById("planbar-verdict").textContent = r.verdict || "";
  document.getElementById("planbar-beads").innerHTML = ms.map(m =>
    `<i title="${esc(m.name)} ${(m.ruin * 100).toFixed(1)}%" style="background:${beadColor(m.ruin)}"></i>`).join("");
}
document.getElementById("targetRuin").addEventListener("input", schedule);

// renderReadout paints the hero instrument: the big central-case ruin figure,
// the tolerance chip, and the gauge (value fill + tolerance tick) on a scale
// stretching to max(10%, 2.5x tolerance) so the tick sits inside the dial.
function renderReadout(ruin, target) {
  const big = document.getElementById("ruinBig");
  const chip = document.getElementById("ruinChip");
  if (!isFinite(ruin)) { big.textContent = "·"; chip.hidden = true; return; }
  big.innerHTML = `${(ruin * 100).toFixed(1)}<span class="pct">%</span>`;
  const grade = ruin <= target ? "good" : ruin <= 1.5 * target ? "warn" : "bad";
  chip.hidden = false;
  chip.className = "chip " + grade;
  chip.textContent = grade === "good"
    ? `inside your ${(target * 100).toFixed(1).replace(/\.0$/, "")}% tolerance`
    : `above your ${(target * 100).toFixed(1).replace(/\.0$/, "")}% tolerance`;
  const scale = Math.max(0.10, 2.5 * target, 1.25 * ruin);
  document.getElementById("gaugeFill").style.width = (100 * Math.min(1, ruin / scale)).toFixed(1) + "%";
  document.getElementById("gaugeFill").className = "fill " + grade;
  document.getElementById("gaugeLim").style.left = (100 * target / scale).toFixed(1) + "%";
  document.getElementById("gaugeTol").textContent = `tolerance ${(target * 100).toFixed(1).replace(/\.0$/, "")}%`;
  document.getElementById("gaugeTol").style.left = (100 * target / scale).toFixed(1) + "%";
  document.getElementById("gaugeMax").textContent = (scale * 100).toFixed(0) + "%";
}

// Show the plan bar only while the hero is out of view.
const planbar = document.getElementById("planbar");
new IntersectionObserver(entries => {
  planbar.hidden = entries[0].isIntersecting;
}, {rootMargin: "-60px 0px 0px 0px"}).observe(document.getElementById("hero"));

// ---------------------------------------------------------------------------
// Section renderers.
// ---------------------------------------------------------------------------
async function renderPaths(b, id) {
  try {
    const r = await post("/api/paths", b);
    if (!fresh(id)) return;
    document.getElementById("fansGrid").innerHTML =
      (r.fans || []).map(f => `<div class="fan">${f.svg || ""}</div>`).join("");
  } catch (e) { /* keep the previous charts */ }
}

// §01 first row: the market each model imagines, before any withdrawal.
async function renderMarket(b, id) {
  try {
    const r = await post("/api/market", b);
    if (!fresh(id)) return;
    document.getElementById("marketGrid").innerHTML =
      (r.fans || []).map(f => `<div class="fan">${f.svg || ""}</div>`).join("");
  } catch (e) { /* keep the previous charts */ }
}

async function renderFrontier(b, id) {
  try {
    const r = await post("/api/frontier", b);
    if (fresh(id)) setSVG("frontierSvg", r.frontierSvg);
  } catch (e) { /* keep the previous chart */ }
}

async function renderPolicyFrontier(b, id) {
  try {
    const r = await post("/api/policyfrontier", b);
    if (fresh(id)) setSVG("policyFrontierSvg", r.policyFrontierSvg);
  } catch (e) { /* keep the previous chart */ }
}

async function renderSensitivity(b, id) {
  try {
    const r = await post("/api/sensitivity", b);
    if (fresh(id)) setSVG("sensitivitySvg", r.sensitivitySvg);
  } catch (e) { /* keep the previous chart */ }
}

async function renderSpending(b, id) {
  try {
    const r = await post("/api/spending", b);
    if (!fresh(id)) return;
    setSVG("spendingSvg", r.spendingSvg);
    document.getElementById("spendingCards").innerHTML = cardsHTML(r.cards);
  } catch (e) { /* keep the previous chart */ }
}

// §02: the plan replayed through the infamous historical start dates.
async function renderVintages(b, id) {
  try {
    const r = await post("/api/vintages", b);
    if (!fresh(id)) return;
    setSVG("vintagesSvg", r.vintagesSvg);
    document.getElementById("vintagesCards").innerHTML =
      cardsHTML(r.cards) + (r.note ? `<p class="cardnote">${esc(r.note)}</p>` : "");
  } catch (e) { /* keep the previous chart */ }
}

// §03: ruin decomposed by the first decade's market return.
async function renderDecade(b, id) {
  try {
    const r = await post("/api/decade", b);
    if (!fresh(id)) return;
    setSVG("decadeSvg", r.decadeSvg);
    document.getElementById("decadeCards").innerHTML = cardsHTML(r.cards);
  } catch (e) { /* keep the previous chart */ }
}

// §04 second row: the median path's funding mix.
async function renderIncome(b, id) {
  try {
    const r = await post("/api/income", b);
    if (!fresh(id)) return;
    setSVG("incomeSvg", r.incomeSvg);
    document.getElementById("incomeCards").innerHTML = cardsHTML(r.cards);
  } catch (e) { /* keep the previous chart */ }
}

async function renderLifecycle(b, id) {
  try {
    const r = await post("/api/lifecycle", b);
    if (!fresh(id)) return;
    setSVG("lifeSvg", r.lifeSvg);
    setSVG("ruinYearSvg", r.ruinYearSvg);
    setSVG("causesSvg", r.causesSvg);
    setSVG("bequestSvg", r.bequestSvg);
    document.getElementById("lifecycleCards").innerHTML = cardsHTML(r.cards);
  } catch (e) { /* keep the previous chart */ }
}

async function renderSim(b, id) {
  try {
    const r = await post("/api/sim", b);
    if (!fresh(id)) return;
    document.getElementById("note").textContent = r.note || "";
    setSVG("arbitrageSvg", r.arbitrageSvg);
    setSVG("arbitrage2Svg", r.arbitrage2Svg);
    setSVG("recoverySvg", r.recoverySvg);
    document.getElementById("cards").innerHTML = cardsHTML(r.cards);
  } catch (e) { /* keep the previous cards */ }
}

// Solver menu: the equivalent ways to reach the acceptable ruin.
async function renderSolver(b, id) {
  let m;
  try { m = await post("/api/solvemenu", b); } catch (e) { return; }
  if (!fresh(id)) return;
  const head = m.met
    ? `<b>Your plan meets the target</b> (ruin ${(m.currentRuin * 100).toFixed(1)}% ≤ ${(m.targetRuin * 100).toFixed(1)}%):`
    : `<b>To get ruin down to ${(m.targetRuin * 100).toFixed(1)}%</b> (now ${(m.currentRuin * 100).toFixed(1)}%), any one of:`;
  const items = (m.options || []).map(o =>
    `<li class="${o.ok ? "" : "no"}"><span class="lev">${esc(o.lever)}</span><span>${o.ok ? "" : "✗ "}${esc(o.text)}</span></li>`).join("");
  // Few, long options (the "plan meets the target" case) read better in one
  // column so the sentence doesn't wrap at the middle of the page.
  const wide = m.met || (m.options || []).length <= 2 ? " wide" : "";
  document.getElementById("solvermenu").innerHTML =
    `<div class="solvehead">${head}</div><ul class="solveopts${wide}">${items}</ul>`;
}

// ---------------------------------------------------------------------------
// Allocation bar (portfolio mode): drag a divider to shift weight.
// ---------------------------------------------------------------------------
function renderAlloc() {
  const bar = document.getElementById("allocbar");
  bar.innerHTML = "";
  let cum = 0; const cums = [0];
  for (const w of weights) { cum += w; cums.push(cum); }
  weights.forEach((w, i) => {
    const seg = document.createElement("div");
    seg.className = "seg";
    seg.style.left = (cums[i] * 100) + "%";
    seg.style.width = (w * 100) + "%";
    seg.style.background = PAL[i % PAL.length];
    seg.innerHTML = `<span>${esc(labels[i])}</span><b>${Math.round(w * 100)}%</b>`;
    bar.appendChild(seg);
  });
  for (let i = 0; i < weights.length - 1; i++) {
    const h = document.createElement("div");
    h.className = "handle";
    h.style.left = (cums[i + 1] * 100) + "%";
    h.addEventListener("pointerdown", ev => startDrag(ev, i));
    bar.appendChild(h);
  }
  const leg = document.getElementById("alloclegend");
  leg.innerHTML = labels.map((n, i) =>
    `<span><i style="background:${PAL[i % PAL.length]}"></i>${esc(n)} ${Math.round(weights[i] * 100)}%</span>`).join("");
}

function startDrag(ev, i) {
  ev.preventDefault();
  const bar = document.getElementById("allocbar");
  const rect = bar.getBoundingClientRect();
  const left = weights.slice(0, i).reduce((a, b) => a + b, 0);
  const pair = weights[i] + weights[i + 1];
  function move(e) {
    let x = (e.clientX - rect.left) / rect.width;
    x = Math.max(left, Math.min(left + pair, x));
    weights[i] = x - left;
    weights[i + 1] = pair - (x - left);
    renderAlloc();
    schedule();
  }
  function up() {
    window.removeEventListener("pointermove", move);
    window.removeEventListener("pointerup", up);
  }
  window.addEventListener("pointermove", move);
  window.addEventListener("pointerup", up);
}

// ---------------------------------------------------------------------------
// Portfolio mode bootstrap: fetch holdings, seed the fit, add the bar.
// ---------------------------------------------------------------------------
fetch("/api/meta").then(r => r.json()).then(m => {
  renderCape(m.cape);
  setSVG("capeHistory", m.capeHistory);
  if (!m.hasPanel) { run(); runSlow(); return; }
  hasPanel = true;
  labels = m.labels;
  weights = (sharedWeights && sharedWeights.length === labels.length) ? sharedWeights.slice()
    : (m.weights && m.weights.length === labels.length) ? m.weights.slice()
    : labels.map(() => 1 / labels.length);
  lastFitW = weights.slice(); // mu/sigma seeded below; avoid a redundant refit
  for (const [k, v] of [["mu", m.mu], ["sigma", m.sigma], ["df", m.df]])
    if (typeof v === "number") setSliderVal(k, v);
  applySharedSliders(); // a shared mu/sigma/df overrides the historical seed
  if (state.conservative) applyReturns(PRIOR); // the prior wins over the fit

  // The hero strip evaluates every return model side by side; the detail
  // charts below use the central parametric one.
  state.model = "parametric";

  const box = document.createElement("div");
  box.className = "group";
  box.innerHTML = `<div class="group-h">Allocation</div>
    <div class="ctl span" data-help="Drag a divider to shift weight between adjacent holdings. Every model re-fits (μ/σ/df and the historical panel) from the live weights.">
      <span class="lab"><span>Drag a divider to shift weight</span></span>
      <div class="allocbar" id="allocbar"></div><div class="alloclegend" id="alloclegend"></div>
    </div>`;
  form.insertBefore(box, form.children[1] || null);
  renderAlloc();
  run();
  runSlow();
});

