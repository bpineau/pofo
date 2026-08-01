// FIRE explorer front-end: a grouped control rail on the left, a long
// analysis column on the right, and a sticky plan bar echoing the live
// verdict. All charts are server-rendered SVG; this file only wires state,
// fetches and DOM.

// ---------------------------------------------------------------------------
// Control definitions. r() is a slider, c() a checkbox.
// Every control carries a plain-language data-help hover.
// ---------------------------------------------------------------------------
const r = (key, label, min, max, step, def, unit, help) =>
  ({kind: "range", key, label, min, max, step, def, unit, help});
const c = (key, label, help) => ({kind: "check", key, label, help});

const GROUPS = [
  {title: "Your situation", col: 0, items: [
    r("capital", "Deployed capital", 600000, 6000000, 10000, 600000, "eur",
      "Liquid capital deployed for the retirement, excluding your home and the emergency fund."),
    r("age", "Age at retirement", 20, 60, 1, 40, "int",
      "Age in year 0. Drives the mortality view (section 05): being broke at 61 and at 92 are different life events."),
    r("years", "Horizon (years)", 20, 60, 1, 42, "int",
      "Plan past your life expectancy: ruin rises steeply with the horizon (40→50y nearly doubles it)."),
    r("needAnnual", "Net spending /yr", 12000, 240000, 1000, 24000, "eur",
      "Real (inflation-indexed) net-of-tax household spending. 24 k€/yr = 2 000 €/month."),
  ]},
  {title: "Pension & side income", col: 1, items: [
    r("pensionAnnual", "Pension /yr (net real)", 0, 60000, 600, 12000, "eur",
      "Net real pension once it starts. Simulations show this is the plan's second-biggest sensitivity."),
    r("pensionYear", "Pension starts in year", 0, 40, 1, 25, "int",
      "Years from retirement to the pension. Retiring at 42 with a pension at 67 = year 25."),
    r("sideAnnual", "Side income /yr", 0, 100000, 1000, 0, "eur",
      "Temporary net real income (rental, activity…) subtracted from the need while it lasts. Income covering the early years is the best sequence-risk insurance there is."),
    r("sideUntilYear", "Side income until year", 0, 40, 1, 0, "int",
      "The side income runs from year 0 up to (excluding) this year."),
  ]},
  {title: "Spending policy", col: 2, items: [
    r("flexCut", "Cut in downturns (0 = fixed rule)", 0, 0.40, 0.05, 0, "pct",
      "Reversible spending cut while the portfolio drawdown exceeds 20%. The single most powerful lever: 15% roughly halves ruin. Section 04 shows its lived cost."),
    r("wrTrigger", "Also cut above this WR (0 = off)", 0, 0.06, 0.002, 0, "pct",
      "Second trigger from the written rules: cut whenever the current withdrawal rate (spend / portfolio) exceeds this, e.g. 3.6%."),
    c("guardrails", "Guyton-Klinger guardrails",
      "Adjust spending ±10% whenever the current withdrawal rate leaves a ±20% band around the initial rate. It trades ruin risk for LIFESTYLE risk: cuts repeat while the portfolio keeps underperforming, so in the bad tail income decays step after step (section 04 shows the lived cost; the frontier in section 06 shows the trade). The guardrails ruin figure is therefore not comparable to the fixed rule's: it means 'even cutting without limit was not enough'. Set a floor below to bound the descent."),
    c("riskGuard", "Risk-based guardrails (Kitces / Morningstar)",
      "The state of the art of the guardrails family, and the same machinery as the line above with a better sensor. Guyton-Klinger compares your withdrawal rate to a band around the rate you STARTED at, a sensor blind to everything since: it cuts a 78-year-old whose remaining horizon makes 6% perfectly sustainable, and stays quiet on a 52-year-old drifting into trouble. Here the band follows the rate that is still safe for the horizon you have LEFT, and the rate is read on total wealth, so the pension you are owed counts before it starts, discounted. Same +-20% band, same +-10% moves, same floor slider; raises stop at 150% of your planned spending, so the rule cannot ratchet a lifestyle you never asked for. TWO THINGS TO KNOW. The safe-rate table is solved under the model SELECTED in the strip above and then lived with, so planning on a rosy fitted history buys a permissive table that a harsher column will then catch out: that gap is the model risk, shown rather than hidden. And late in the horizon the band opens wide on purpose, because a plan that ends at the horizon is meant to be spent, which is the ABW behaviour arriving through the guardrails door."),
    r("gkRaiseCap", "Risk-guardrail raise ceiling (x planned spend)", 1, 2, 0.05, 1, "mult",
      "How far the risk-based rule may ratchet your spending UP when the sensor says you have room. The default is 1.00, protective only: it cuts when the plan drifts into trouble and never raises, which is the setting that makes its ruin figure comparable to the fixed rule's, since both then deliver the same lifestyle until something breaks. Above 1.00 the rule also spends the good news, and its ruin figure then measures a RICHER life: 1.50 means it may work its way up to half again your planned spending, so comparing that ruin to the fixed rule's is the same category error as comparing Guyton-Klinger's to Bengen's. Section 04 shows which lifestyle each number is actually buying. Only applies while the risk-based guardrails are on."),
    r("gkFloor", "Guardrails floor (% of initial, 0 = none)", 0, 0.9, 0.05, 0, "pct",
      "The incompressible standard: guardrails cuts never push spending below this share of the initial level. Bounding the descent re-creates some ruin (the floor itself can prove unsustainable), which is the honest trade; 70-80% is a common planning value. Applies to both guardrails flavours."),
    c("ratchet", "Ratchet lifestyle up when rich",
      "Only-up rule from the written rules: +10% of the base spend when real wealth exceeds 120% of the start, at most every 2 years, capped at 120% of the base, only while the current rate is below 2.2%."),
    r("spendDrift", "Real spending drift /yr", -0.01, 0.02, 0.001, 0, "pct",
      "Structural real drift of the need: health insurance and care costs drift upward faster than inflation (+0.3-0.5%/yr is a common planning value)."),
    c("smile", "Retirement smile (down, plateau, up late)",
      "Blanchett's observed shape: real spending drifts down through the go-go years, plateaus, then climbs back with late-life health costs."),
    r("percent", "Spend % of portfolio (VPW, 0 = off)", 0, 0.08, 0.005, 0, "pct",
      "Percentage-of-portfolio (VPW) rule: each year spend this share of the current portfolio instead of a fixed amount. It never runs out, but the standard of living swings with the market. The other end of the decumulation frontier. Only one spending rule runs at a time, so raising it clears the fixed need's flex cut, guardrails and ratchet."),
    c("abw", "Amortize over the horizon (ABW / TPAW)",
      "ABW = Amortization-Based Withdrawal; TPAW = Total Portfolio Allocation and Withdrawal, the popular planner built on it. The rule much of the recent literature prefers: each year, spend the payment that would exhaust your CURRENT wealth (after tax, plus the present value of future pensions) exactly over the REMAINING years, assuming the central expected return, a mortgage run in reverse, re-quoted every year. Why it is attractive: it can never run out early (spending is always the sustainable share of what remains), it never dies on a mountain of unspent money (the payout rate rises as the horizon shortens), and it self-corrects continuously in small steps instead of Guyton-Klinger's -10% jolts. The price: income follows the market; after a bad decade the payment is genuinely lower. Assumes the geometric central return (CAPE-implied when the valuation anchor is on). Only one spending rule runs at a time, so ticking it clears the others."),
    c("bounded", "Bounded % of portfolio (Vanguard-style)",
      "Vanguard's 'dynamic spending', the industry's smoothed VPW: each year target the initial percentage of CURRENT wealth, but never move real spending more than +5% up or -2.5% down from last year. In good markets your lifestyle drifts up slowly; in crashes it glides down 2.5% a year instead of jumping. Because the descent is capped, spending can outrun a collapsing portfolio, so unlike VPW/ABW this rule CAN still run out: it sits between the fixed rule and VPW on the frontier. Only one spending rule runs at a time, so ticking it clears the others."),
    r("annuityShare", "Annuitise % of capital", 0, 0.5, 0.05, 0, "pct",
      "Spend this share of capital on a joint-life, inflation-linked annuity (1% real rate, 10% insurer load): a guaranteed lifelong income floor that hedges longevity. It converts growth assets into lower guaranteed income, so headline ruin (failing the FULL need) can rise even as the worst late-life outcomes improve; its value is the floor, not the average."),
  ]},
  {title: "Market model", col: 3, items: [
    r("mu", "Real growth return", 0.01, 0.12, 0.005, 0.05, "pct",
      "Arithmetic mean annual real return of the growth sleeve. Geometric ≈ μ − σ²/2, so 5%/11% ≈ 4.4% real compounding."),
    r("sigma", "Volatility (long-horizon)", 0.06, 0.20, 0.005, 0.11, "pct",
      "Long-horizon annual volatility (variance-ratio-consistent), lower than the 1-year headline vol. Vol matters almost as much as return: 10→12% nearly doubles ruin."),
    r("df", "Tail df (low = fat)", 3, 30, 1, 5, "int",
      "How much more often EXTREME years happen than a bell curve allows, at the same volatility. At df 5 a catastrophic year (say -30% real, a 3-sigma event) is roughly ten times more likely than at df 30 (≈ normal law); ordinary years barely change. Low df = crash-prone world. In portfolio mode it is seeded from the holdings' monthly kurtosis."),
    c("conservative", "Broad-sample prior (override the fit)",
      "Rewrites the three μ/σ/df sliders with cautious world-equity assumptions (~3.5% real geometric, fat tails): what broad century-long samples suggest, not this fund's history. Sliders only: the data-driven columns (Historical, Block bootstrap, Broad-sample) never read the sliders and are untouched."),
    c("capeAdjust", "Anchor return to today's valuation (CAPE)",
      "Replaces ONLY the central mean with the CAPE-implied estimate (1/CAPE): today's rich valuation compresses the next decade, the years that make or break a retirement. Volatility and tails keep their fitted values; the data-driven columns are untouched. Overrides the return slider and the broad-sample blend."),
    c("glidepath", "Rising-equity glidepath (bond tent)",
      "Hold less equity at retirement (30%) gliding to 75% later, blending in bonds (~1.5% real). It cuts sequence risk (the danger years are the least equity-heavy) but gives up return; with a wide equity-bond gap the drag can outweigh the protection, so compare the ruin both ways. Applies to the central case."),
    c("monthly", "Monthly withdrawals (salary-like)",
      "Step the kernel monthly instead of annually: withdrawals, drawdown checks and the bucket rule run every month. Only the plan-detail and buffer sections (§07-§08) honour it; the model strip and the analysis sections always compare annual kernels, for speed and comparability. The effect is small by design (a refinement, not a lever)."),
  ]},
  {title: "Cash buffer", col: 1, items: [
    r("bufferYears", "Buffer (years of spending)", 0, 10, 1, 3, "int",
      "Low-volatility pocket (cash + short euro linkers) drained when the portfolio is down >10%. Statistically the arbitrage is flat past 2-3 years; its value is behavioural and inflation matching."),
    r("bufferReturn", "Buffer real return", -0.01, 0.05, 0.005, 0.005, "pct",
      "Real return of the buffer: ~0% for inflation-linked, negative for pure cash."),
    r("bufferStopYear", "Buffer refill stops in year (0 = never)", 0, 20, 1, 0, "int",
      "The melting buffer: stop refilling after the sequence-risk window (e.g. year 8) and let it run down: a bond-tent glidepath."),
  ]},
  {title: "Taxes", col: 1, items: [
    r("taxRate", "Tax on gains", 0, 0.40, 0.01, 0.328, "pct",
      "Your country's rate on realised investment gains, charged on the GAIN share of every sale (withdrawing 60k net sells more than 60k of assets). Use your blended effective rate across accounts, e.g. ~31-35% for a plain French taxable account (the 2026 flat tax is 31.4%, plus the PUMa levy), less if part of the capital sits in sheltered wrappers. The effective burden starts low and drifts up as unrealised gains compound."),
  ]},
  {title: "Simulation", col: 0, id: "group-simulation", items: [
    r("nPaths", "Simulated paths", 1000, 10000, 500, 2000, "int",
      "Monte-Carlo paths per model. At 2000 the ruin figure wobbles by roughly ±0.7 point run to run (pure sampling noise); 4000 halves the wobble, 8000 halves it again. Raise it if the figures dance when nothing changed; the engine is fast."),
  ]},
];

const FMT = {
  pct: v => (v * 100).toFixed(1).replace(/\.0$/, "") + "%",
  mult: v => "x" + v.toFixed(2),
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

// The strip is the model selector: clicking a column runs every detail
// section (spending, lifecycle, risk, buffer…) under that lens. MODELKEY
// maps column names to the server's Central parameter; "" = the calibrated
// central case (Student-t / glidepath).
const MODELKEY = {
  "Historical windows": "hist", "Block bootstrap": "boot", "Student-t": "",
  "Sequence stress": "stress", "Broad-sample": "broad", "Lost decade": "lost",
};

// ---------------------------------------------------------------------------
// Build the rail.
// ---------------------------------------------------------------------------
const form = document.getElementById("controls");
const state = {};
const checkEls = {};
const ctlEls = {};

function renderRail() {
  // Four explicit columns (deterministic bin-packing): the situation and the
  // simulation size on the left, income and the money pockets next, then the
  // two tall rule groups each get a full column.
  const cols = [];
  for (let i = 0; i < 4; i++) {
    const c = document.createElement("div");
    c.className = "ctlcol";
    form.appendChild(c);
    cols.push(c);
  }
  GROUPS.forEach((g, gi) => {
    const box = document.createElement("div");
    box.className = "group";
    if (g.id) box.id = g.id;
    box.innerHTML = `<div class="group-h">${g.title}</div>`;
    // The plan-defining sliders (first group) get a ruler of ticks.
    for (const it of g.items) box.appendChild(buildControl(it, gi === 0));
    cols[g.col || 0].appendChild(box);
  });
  // A full-width footer to persist the personal-defaults subset to a cookie.
  const foot = document.createElement("div");
  foot.className = "rail-foot";
  foot.setAttribute("data-help",
    "Save your capital, age, horizon, net spending and pension figures as personal defaults in a cookie on this browser, so you land on them next time. Click again to update them. It changes nothing you can share: the page URL always reproduces the exact scenario for anyone, cookie or not.");
  foot.innerHTML = `<button type="button" id="saveDefaults" class="save-defaults">Save as my defaults</button><span class="saved-note" id="savedNote" hidden>saved</span>`;
  form.appendChild(foot);
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
    ctlEls[it.key] = d;
    input.addEventListener("change", e => {
      state[it.key] = e.target.checked;
      if (it.key === "conservative") applyConservative();
      if (e.target.checked) claimPolicy(it.key);
      syncPolicy();
      schedule();
    });
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
  ctlEls[it.key] = d;
  paintFill(input);
  input.addEventListener("input", e => {
    state[it.key] = parseFloat(e.target.value);
    paintFill(e.target);
    refreshVal(it.key);
    if (state[it.key] !== it.def) claimPolicy(it.key);
    syncPolicy();
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
function setCheckVal(k, v) {
  state[k] = v;
  if (checkEls[k]) checkEls[k].checked = v;
}

// ---------------------------------------------------------------------------
// Spending policies are mutually exclusive. The kernel applies exactly one,
// in a fixed precedence (ABW > bounded > VPW > risk guardrails >
// Guyton-Klinger > the fixed rule and its flex cut / ratchet), so leaving
// every box tickable let a reader believe rules combine when the extra ones
// were silently ignored. Touching a control now claims its policy and clears
// the others; a shared URL is left as it arrived, only masked, so old links
// keep reproducing the exact run their sender saw.
// ---------------------------------------------------------------------------
const POLICIES = [
  {id: "abw", trigger: ["abw"], owns: ["abw"]},
  {id: "bounded", trigger: ["bounded"], owns: ["bounded"]},
  {id: "percent", trigger: ["percent"], owns: ["percent"]},
  {id: "riskGuard", trigger: ["riskGuard"], owns: ["riskGuard", "gkRaiseCap", "gkFloor"]},
  {id: "guardrails", trigger: ["guardrails"], owns: ["guardrails", "gkFloor"]},
  {id: "fixed", trigger: ["flexCut", "wrTrigger", "ratchet"],
    owns: ["flexCut", "wrTrigger", "ratchet"]},
];
const POLICYOF = {};
for (const p of POLICIES) for (const k of p.trigger) POLICYOF[k] = p.id;
const DEFOF = {};
for (const g of GROUPS) for (const it of g.items) DEFOF[it.key] = it.kind === "check" ? false : it.def;

// claimPolicy resets every control owned by the policies that k displaces.
// Controls shared with the claiming policy (the guardrails floor) survive.
function claimPolicy(k) {
  const mine = POLICIES.find(p => p.id === POLICYOF[k]);
  if (!mine) return;
  for (const p of POLICIES) {
    if (p.id === mine.id) continue;
    for (const key of p.owns) {
      if (mine.owns.includes(key) || state[key] === DEFOF[key]) continue;
      if (checkEls[key]) setCheckVal(key, false); else setSliderVal(key, DEFOF[key]);
    }
  }
}
// syncPolicy dims the parameters whose owning rule is off, so the rail never
// offers a live-looking control that the kernel is ignoring.
function syncPolicy() {
  const dim = (key, on) => {
    const el = ctlEls[key];
    if (!el) return;
    el.classList.toggle("masked", !on);
    const input = el.querySelector("input");
    if (input) input.disabled = !on;
  };
  dim("gkFloor", state.guardrails || state.riskGuard);
  dim("gkRaiseCap", state.riskGuard);
}

renderRail();
refreshAges();
document.getElementById("s_age").addEventListener("input", refreshAges);
document.getElementById("saveDefaults").addEventListener("click", () => {
  saveDefaults();
  const note = document.getElementById("savedNote");
  note.hidden = false;
  clearTimeout(note._t);
  note._t = setTimeout(() => { note.hidden = true; }, 1800);
});

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
// Personal defaults: a chosen subset of the situation sliders, saved to a
// cookie on demand (the "Save as my defaults" button) so a regular visitor
// lands on their own figures instead of the generic ones. The cookie seeds
// ONLY these keys, and a shared #hash always overrides it (applied after), so
// a link reproduces the sender's exact scenario for anyone, cookie or not.
// ---------------------------------------------------------------------------
const SAVEKEYS = ["capital", "age", "years", "needAnnual", "pensionAnnual", "pensionYear"];
const PREF_COOKIE = "fire_defaults";
function readCookie(name) {
  const m = document.cookie.match(new RegExp("(?:^|; )" + name + "=([^;]*)"));
  return m ? decodeURIComponent(m[1]) : "";
}
function applySavedDefaults() {
  const raw = readCookie(PREF_COOKIE);
  if (!raw) return;
  const p = new URLSearchParams(raw);
  for (const k of SAVEKEYS)
    if (p.has(k)) { const v = parseFloat(p.get(k)); if (!isNaN(v)) setSliderVal(k, v); }
}
function saveDefaults() {
  const p = new URLSearchParams();
  for (const k of SAVEKEYS) p.set(k, state[k]);
  // A year, path-wide, readable by this script (not sensitive; it only seeds
  // the sliders). Re-saving overwrites it.
  document.cookie = PREF_COOKIE + "=" + encodeURIComponent(p.toString()) +
    ";path=/;max-age=31536000;samesite=lax";
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
applySavedDefaults();  // cookie over the generic defaults...
applySharedSliders();  // ...then a shared #hash overrides everything it names
for (const k of CHECKKEYS) {
  if (shared.get(k) === "1") { state[k] = true; checkEls[k].checked = true; }
}
if (state.conservative) applyReturns(PRIOR);
syncPolicy();

function syncURL() {
  const p = new URLSearchParams();
  for (const k of Object.keys(UNIT)) p.set(k, state[k]);
  if (state.central) p.set("central", state.central);
  for (const k of CHECKKEYS) if (state[k]) p.set(k, "1");
  if (weights) p.set("w", weights.map(x => x.toFixed(4)).join(","));
  history.replaceState(null, "", "#" + p.toString());
}
state.central = shared.get("central") || "";

// ---------------------------------------------------------------------------
// Scheduling: a fast lane for the live views, a slow lane for the two
// solver-heavy planning curves. A run id drops stale responses.
// ---------------------------------------------------------------------------
let timer = null, slowTimer = null, runId = 0;
function schedule() {
  clearTimeout(timer); timer = setTimeout(run, 200);
  clearTimeout(slowTimer); slowTimer = setTimeout(runSlow, 600);
}
// apiURL maps an absolute "/api/…" route onto the document's base URL so
// the app reaches its own backend whether it is served at the site root
// (pofo -fire) or mounted under a sub-path (pofo -serve serves it at
// /firesimulator/, where document.baseURI ends in "/firesimulator/").
const apiURL = p => new URL(p.replace(/^\//, ""), document.baseURI).href;
const post = (url, body) =>
  fetch(apiURL(url), {method: "POST", headers: {"Content-Type": "application/json"}, body: JSON.stringify(body)})
    .then(r => r.json());
// fresh(id) is true while no newer run started: renderers check it before
// touching the DOM so a slow older response never overwrites a newer one.
const fresh = id => id === runId;

function body() {
  const b = {...state};
  for (const k of INTKEYS) b[k] = Math.round(b[k]);
  b.targetRuin = (parseFloat(document.getElementById("targetRuin").value) || 5) / 100;
  b.central = state.central || "";
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
  // The command echo and the URL reflect the plan, not the results, so update
  // them at once rather than after the simulations.
  updateCmd();
  syncURL();
  // Two waves, so the top of the page answers first. Tier 1 feeds what the eye
  // is on while dragging a slider: the hero verdict, the model strip and the
  // simulated-futures fans. Fire it and wait for it, so on a low-core host it
  // wins the cores before the heavier analysis lower down starts competing.
  await Promise.all([renderModels(b, id), renderPaths(b, id), renderMarket(b, id)]);
  // A newer run started while tier 1 was in flight: skip the analysis half
  // entirely rather than compute it for a plan already superseded. A fast drag
  // thus pays for the heavy sections only once, at the value it lands on.
  if (!fresh(id)) return;
  // Tier 2, dispatched top-to-bottom in page order: the sections you scroll to.
  renderVintages(b, id);
  renderDecade(b, id);
  renderSpending(b, id);
  renderIncome(b, id);
  renderLifecycle(b, id);
  renderFrontier(b, id);
  renderPolicyFrontier(b, id);
  renderSensitivity(b, id);
  renderSim(b, id);
  renderSolver(b, id);
};

// The top-bar command echo mirrors the live plan, terminal-style.
function updateCmd() {
  const el = document.getElementById("cmdEcho");
  if (!el) return;
  const rule = state.abw ? "abw" : state.bounded ? "bounded%" : state.percent > 0 ? "vpw"
    : state.guardrails ? "guardrails" : state.flexCut > 0 ? "flex" : "fixed·real";
  const model = {hist: "historical", boot: "bootstrap", stress: "seq-stress", broad: "broad-sample", lost: "lost-decade"}[state.central]
    || (state.conservative ? "prior" : state.capeAdjust ? "cape" : "student-t");
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
// Opening from deep in the page scrolls back up to the drawer, otherwise the
// click appears to do nothing (the drawer unfolds above the viewport).
const drawerToggle = document.getElementById("drawerToggle");
const drawerEl = document.getElementById("drawer");
function setDrawer(open) {
  if (!drawerToggle || !drawerEl) return;
  if (open) {
    drawerEl.removeAttribute("hidden");
    window.scrollTo({top: 0, behavior: "smooth"});
  } else {
    drawerEl.setAttribute("hidden", "");
  }
  drawerToggle.setAttribute("aria-expanded", String(open));
  drawerToggle.textContent = open ? "parameters ▴" : "parameters ▾";
}
if (drawerToggle && drawerEl)
  drawerToggle.addEventListener("click", () => setDrawer(drawerEl.hasAttribute("hidden")));

// flashGroup outlines one control group for a moment, so a click that opens
// the drawer says where in it to look.
function flashGroup(id) {
  const el = document.getElementById(id);
  if (!el) return;
  el.classList.remove("flash");
  void el.offsetWidth; // restart the animation on a repeated click
  el.classList.add("flash");
  clearTimeout(el._flash);
  el._flash = setTimeout(() => el.classList.remove("flash"), 1600);
}

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
  if (el && svg) {
    el.innerHTML = svg;
    enhanceChart(el);
  }
}

// enhanceChart makes a freshly rendered chart keyboard-usable: the container
// becomes focusable (arrow keys drive the crosshair, T opens the data table)
// and gains a small "data" button. Idempotent per render, since setSVG wipes
// the container's children.
function enhanceChart(el) {
  const svg = el.querySelector("svg");
  if (!svg || !svg.querySelector("metadata.hover")) return;
  el.tabIndex = 0;
  if (!el.__keys) {
    el.__keys = true;
    el.addEventListener("keydown", chartKeydown);
    el.addEventListener("blur", hideCrosshair);
  }
  const help = el.getAttribute("data-help");
  if (help) {
    // Over an instrumented chart the crosshair tooltip wins, so the frame's
    // explanation would only surface on its border: give it its own affordance.
    const hb = document.createElement("button");
    hb.type = "button";
    hb.className = "databtn helpbtn";
    hb.textContent = "? what this says";
    hb.setAttribute("aria-label", "What this chart says");
    hb.setAttribute("data-help", help);
    el.appendChild(hb);
  }
  const btn = document.createElement("button");
  btn.type = "button";
  btn.className = "databtn";
  btn.textContent = "\u2317 data";
  btn.setAttribute("aria-label", "Show this chart's data as a table");
  btn.title = "Show the data as a table (T when the chart is focused)";
  btn.addEventListener("click", ev => {
    ev.stopPropagation();
    showTable(el.querySelector("svg"));
  });
  el.appendChild(btn);
}

// chartKeydown drives the crosshair from the keyboard: left/right step
// through the x positions, T (or Enter) opens the table, Escape clears.
function chartKeydown(e) {
  const el = e.currentTarget;
  const svg = el.querySelector("svg");
  const hd = svg ? hoverData(svg) : null;
  if (!hd) return;
  if (e.key === "t" || e.key === "T" || e.key === "Enter") {
    e.preventDefault();
    showTable(svg);
    return;
  }
  if (e.key === "Escape") { hideCrosshair(); return; }
  if (e.key !== "ArrowLeft" && e.key !== "ArrowRight") return;
  if (!["line", "fan", "stack"].includes(hd.kind)) return; // discrete marks: table only
  e.preventDefault();
  let steps = 2;
  for (const s of hd.series) steps = Math.max(steps, (s.xs || s.ys).length);
  const unit = (hd.xmax - hd.xmin) / (steps - 1);
  if (svg.__kx === undefined) svg.__kx = hd.xmin + (hd.xmax - hd.xmin) / 2;
  svg.__kx += e.key === "ArrowRight" ? unit : -unit;
  svg.__kx = Math.min(Math.max(svg.__kx, hd.xmin), hd.xmax);
  // Anchor the tooltip at the crosshair position inside the chart.
  const rect = svg.getBoundingClientRect();
  const vb = svg.viewBox.baseVal;
  const cx = hd.x0 + (svg.__kx - hd.xmin) / (hd.xmax - hd.xmin) * (hd.x1 - hd.x0);
  showCrosshair(svg, hd, svg.__kx,
    rect.left + cx / vb.width * rect.width,
    rect.top + (hd.y0 + hd.y1) / 2 / vb.height * rect.height);
}

// showTable renders the chart's embedded data as an HTML table in a modal
// overlay: the always-reachable, hover-free reading of every chart.
function showTable(svg) {
  const hd = svg ? hoverData(svg) : null;
  if (!hd) return;
  dtable.textContent = "";
  const box = document.createElement("div");
  box.className = "dtable-box";
  const table = document.createElement("table");
  const discrete = hd.rows && hd.rows.length;
  const thead = document.createElement("tr");
  for (const name of [discrete ? "" : (hd.xlabel || "x"), ...hd.series.map(s => s.name || hd.ylabel || "value")]) {
    const th = document.createElement("th");
    th.textContent = name;
    thead.appendChild(th);
  }
  table.appendChild(thead);
  let n = 0;
  for (const s of hd.series) n = Math.max(n, s.ys.length);
  if (discrete) n = hd.rows.length;
  for (let i = 0; i < n; i++) {
    const tr = document.createElement("tr");
    const td0 = document.createElement("td");
    // Row key: the category label, the series' own x, or the plain index.
    const xs0 = hd.series[0].xs;
    td0.textContent = discrete ? hd.rows[i] : String(xs0 ? (xs0[i] ?? "") : i);
    tr.appendChild(td0);
    for (const s of hd.series) {
      const td = document.createElement("td");
      const v = s.ys[i]; // series shorter than the table leave blanks
      td.textContent = v === undefined ? "" : fmtHV(v);
      tr.appendChild(td);
    }
    table.appendChild(tr);
  }
  box.appendChild(table);
  dtable.appendChild(box);
  dtable.hidden = false;
  document.body.classList.add("noscroll");
}
const dtable = document.createElement("div");
dtable.id = "dtable";
dtable.hidden = true;
document.body.appendChild(dtable);
function closeTable() {
  dtable.hidden = true;
  dtable.textContent = "";
  document.body.classList.remove("noscroll");
}
dtable.addEventListener("click", closeTable);
document.addEventListener("keydown", e => {
  if (e.key === "Escape" && !dtable.hidden) closeTable();
});
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
  if (!hd || !hd.series || !hd.series.length || !["line", "fan", "stack"].includes(hd.kind)) { hideCrosshair(); return; }

  // Pointer position in viewBox coordinates.
  const rect = svg.getBoundingClientRect();
  const vb = svg.viewBox.baseVal;
  const px = (e.clientX - rect.left) * vb.width / rect.width;
  if (px < hd.x0 - 10 || px > hd.x1 + 10) { hideCrosshair(); return; }
  const xdom = hd.xmin + (Math.min(Math.max(px, hd.x0), hd.x1) - hd.x0) / (hd.x1 - hd.x0) * (hd.xmax - hd.xmin);
  showCrosshair(svg, hd, xdom, e.clientX, e.clientY);
});

// showCrosshair snaps the hairline to the nearest x and fills the tooltip:
// the shared engine behind both the pointer and the keyboard paths.
function showCrosshair(svg, hd, xdom, clientX, clientY) {
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
  let tx = clientX + pad, ty = clientY + pad;
  if (tx + tw > innerWidth) tx = clientX - pad - tw;
  if (ty + th > innerHeight) ty = clientY - pad - th;
  xtip.style.left = tx + "px";
  xtip.style.top = ty + "px";
}
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
  // The selected column drives the hero readout and every detail section;
  // Student-t (the calibrated central case) when nothing is selected.
  let central = ms.findIndex(m => MODELKEY[m.name] === (state.central || ""));
  if (central < 0) central = ms.findIndex(m => m.name === "Student-t");
  renderReadout(central >= 0 ? ms[central].ruin : NaN, r.targetRuin || 0.05);
  const eyebrow = document.querySelector(".hero-verdict .eyebrow");
  if (eyebrow && central >= 0) eyebrow.textContent = "Ruin probability · " + ms[central].name;
  const sel = i => (i === central ? ` class="sel"` : "");
  const cells = fn => ms.map((m, i) => `<td${sel(i)}>${fn(m)}</td>`).join("");
  const head = `<tr><th></th>${ms.map((m, i) =>
    `<th${sel(i)} data-model="${esc(m.name)}" data-help="${esc(m.help)}. CLICK to run every detail section of the page (spending, lifecycle, risk, buffer…) under this model.">${m.name}</th>`).join("")}</tr>`;
  // Risk is carried by a coloured dot per cell; the figures stay in ink.
  const ruinRow = `<tr><th data-help="Share of simulated retirements that run out of money, at your planned spend AND under the spending rule you selected (flex cut, guardrails, VPW, ABW, bounded percent). An adaptive rule buys this number down by lowering the income you actually live on, so it is not comparable to the fixed rule's: read section 04 before reading it as safety.">Ruin</th>` +
    cells(m => `<i class="dot" style="background:${beadColor(m.ruin)}"></i>${(m.ruin * 100).toFixed(1)}%`) + `</tr>`;
  const spendRow = `<tr><th data-help="The most you could spend per year and still keep ruin at your acceptable level, under this model. Always solved on the FIXED rule, whatever adaptive rule you selected: a rule that rebases spending on wealth makes ruin non-monotonic in the withdrawal, so the solve would be ill-defined. Everything else (capital, buffer, pension, side income, taxes) is your plan. The gap between this figure and your planned spend is what your flexibility is buying; section 04 shows what it costs in lived income.">Safe spend, no flex</th>` +
    cells(m => `${(m.safeSpend / 1000).toFixed(0)}k€<span class="sub">${(m.safeWR * 100).toFixed(m.safeWR < 0.01 ? 2 : 1)}%</span>`) + `</tr>`;
  const wealthRow = `<tr><th data-help="Median real wealth left at the end of the horizon, at your planned spend.">Median wealth</th>` +
    cells(m => fmtW(m.medianWealth)) + `</tr>`;
  document.getElementById("modelstrip").innerHTML =
    `<table class="modeltab"><thead>${head}</thead><tbody>${ruinRow}${spendRow}${wealthRow}</tbody></table>`;
  for (const th of document.querySelectorAll(".modeltab th[data-model]")) {
    th.addEventListener("click", () => {
      const key = MODELKEY[th.getAttribute("data-model")];
      if (key === undefined) return;
      state.central = key;
      schedule();
    });
  }

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
    for (const el of document.querySelectorAll("#fansGrid .fan")) enhanceChart(el);
  } catch (e) { /* keep the previous charts */ }
}

// §01 first row: the market each model imagines, before any withdrawal.
async function renderMarket(b, id) {
  try {
    const r = await post("/api/market", b);
    if (!fresh(id)) return;
    document.getElementById("marketGrid").innerHTML =
      (r.fans || []).map(f => `<div class="fan">${f.svg || ""}</div>`).join("");
    for (const el of document.querySelectorAll("#marketGrid .fan")) enhanceChart(el);
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
// Market provenance: which market the page runs on (top-bar pill), and, when
// it runs on none, how to load one (the drawer's Portfolio group). Loading is
// pure navigation: the portfolio-bound simulators are their own mounts
// (<base>/e/<name>/ and <base>/p/<spec>/), so the loader only builds a URL and
// carries the current hash along, keeping the visitor's scenario intact.
// ---------------------------------------------------------------------------
const GENERIC_HELP = "This page is running on a generic parametric market: every return is drawn from the mu/sigma/df sliders, not from assets you own. Load a portfolio and the historical models (windows, block bootstrap) and the allocation controls wake up.";

function renderMarketPill(meta) {
  const pill = document.getElementById("marketPill");
  if (!pill) return;
  pill.hidden = false;
  if (meta.hasPanel) {
    pill.className = "pill market loaded";
    pill.textContent = "market: " + (meta.sourceLabel || "your portfolio");
    pill.setAttribute("data-help",
      "The historical models and the allocation bar run on this portfolio's own monthly real returns. Click to open the parameters and re-weight it.");
    pill.addEventListener("click", () => { setDrawer(true); flashGroup("group-allocation"); });
    return;
  }
  pill.className = "pill market generic";
  pill.textContent = "generic market · load portfolio";
  pill.setAttribute("data-help", GENERIC_HELP + " Click to open the loader.");
  pill.addEventListener("click", () => { setDrawer(true); flashGroup("group-portfolio"); });
}

// The /view p= grammar's byte cap and holdings limit, mirrored here so the
// button can refuse early. The server stays authoritative on both.
const PICK_MAXHOLDINGS = 20, PICK_MAXBYTES = 2000, PICK_MAXHITS = 8;

// encP encodes one spec the way the grammar is written by hand (mirrors the
// composer's encP): ":" and "," stay literal, everything else percent-encodes.
const encP = v => encodeURIComponent(v).replace(/%3A/g, ":").replace(/%2C/g, ",");

// go navigates to a portfolio-bound mount, carrying the live scenario hash so
// every slider the visitor already moved survives the move.
const gotoMount = href => location.assign(href + location.hash);

// renderPortfolioGroup builds the drawer's Portfolio group, in the slot the
// Allocation group occupies in portfolio mode. Without a picker (standalone
// pofo -fire, which has neither the mounts nor the catalog endpoint) it is one
// line of command-line instructions.
function renderPortfolioGroup(meta) {
  const sim = document.getElementById("group-simulation");
  if (!sim) return;
  const box = document.createElement("div");
  box.className = "group";
  box.id = "group-portfolio";
  const head = document.createElement("div");
  head.className = "group-h";
  head.textContent = "Portfolio";
  box.appendChild(head);
  const lede = document.createElement("p");
  lede.className = "pick-lede";
  lede.textContent = "No portfolio bound: returns come from the sliders alone. Load one and the historical models and the allocation bar wake up.";
  box.appendChild(lede);
  const picker = meta.picker;
  if (!picker || !picker.base) {
    const hint = document.createElement("p");
    hint.className = "pick-hint";
    hint.innerHTML = `<span class="pick-prompt">$</span> <span class="mono">pofo -fire your-portfolio.txt</span>`;
    box.appendChild(hint);
  } else {
    box.appendChild(buildExampleList(picker));
    box.appendChild(buildMiniComposer(picker));
  }
  sim.parentElement.insertBefore(box, sim);
}

// buildExampleList renders the bundled portfolios as a compact scrollable
// list; a click mounts that example's simulator.
function buildExampleList(picker) {
  const wrap = document.createElement("div");
  wrap.className = "pickblock";
  const cap = document.createElement("div");
  cap.className = "pickcap";
  cap.textContent = "bundled builds";
  wrap.appendChild(cap);
  const list = document.createElement("div");
  list.className = "picklist";
  for (const ex of picker.examples || []) {
    const b = document.createElement("button");
    b.type = "button";
    b.className = "pickex";
    const n = document.createElement("span");
    n.className = "pn";
    n.textContent = ex.title || ex.name;
    b.appendChild(n);
    if (ex.blurb) {
      const d = document.createElement("span");
      d.className = "pb";
      d.textContent = ex.blurb;
      b.appendChild(d);
    }
    b.addEventListener("click", () =>
      gotoMount(picker.base + "/e/" + encodeURIComponent(ex.name) + "/"));
    list.appendChild(b);
  }
  wrap.appendChild(list);
  return wrap;
}

// buildMiniComposer renders the search-and-weigh editor over the same p=
// grammar the report's composer writes: pick assets from the catalog, give
// each an integer weight, load the result as <base>/p/<spec>/.
function buildMiniComposer(picker) {
  const wrap = document.createElement("div");
  wrap.className = "pickblock";
  const cap = document.createElement("div");
  cap.className = "pickcap";
  cap.textContent = "or compose one";
  wrap.appendChild(cap);

  const search = document.createElement("input");
  search.type = "text";
  search.className = "picksearch";
  search.placeholder = "search the catalog";
  search.setAttribute("aria-label", "Search the asset catalog");
  wrap.appendChild(search);
  const hits = document.createElement("div");
  hits.className = "pickhits";
  hits.hidden = true;
  wrap.appendChild(hits);
  const rowsEl = document.createElement("div");
  rowsEl.className = "pickrows";
  wrap.appendChild(rowsEl);
  const sumEl = document.createElement("div");
  sumEl.className = "picksum";
  wrap.appendChild(sumEl);
  const specEl = document.createElement("div");
  specEl.className = "pickspec";
  wrap.appendChild(specEl);
  const load = document.createElement("button");
  load.type = "button";
  load.className = "pickload";
  load.textContent = "Load into simulator";
  load.disabled = true;
  wrap.appendChild(load);
  const why = document.createElement("div");
  why.className = "pickwhy";
  wrap.appendChild(why);

  let catalog = null, catalogFailed = false;
  const rows = []; // [{id, w}]

  // The catalog is fetched on first use, not on page load: the simulator's
  // own charts get the connection first.
  function loadCatalog() {
    if (catalog || catalogFailed) return Promise.resolve();
    return fetch(picker.catalogURL).then(r => r.json()).then(j => { catalog = j || []; })
      .catch(() => {
        catalogFailed = true;
        why.textContent = "the asset catalog did not load; reload the page to search it";
      });
  }
  if (picker.catalogURL) {
    search.addEventListener("focus", loadCatalog, {once: true});
  } else {
    search.disabled = true;
    search.placeholder = "catalog unavailable";
  }

  // match ranks catalog entries by id, alternate identifier (ISIN, ticker,
  // alias) and name, prefix hits first, capped to a readable few.
  function match(q) {
    q = q.trim().toLowerCase();
    if (!q || !catalog) return [];
    const out = [];
    for (const a of catalog) {
      const id = (a.id || "").toLowerCase(), name = (a.name || "").toLowerCase();
      const alt = (a.alt || []).map(x => x.toLowerCase());
      let rank = -1;
      if (id.startsWith(q)) rank = 0;
      else if (alt.some(x => x.startsWith(q))) rank = 1;
      else if (name.startsWith(q)) rank = 2;
      else if (id.includes(q) || name.includes(q) || alt.some(x => x.includes(q))) rank = 3;
      if (rank >= 0) out.push({rank, a});
    }
    out.sort((x, y) => x.rank - y.rank || x.a.id.localeCompare(y.a.id));
    return out.slice(0, PICK_MAXHITS).map(x => x.a);
  }

  function renderHits() {
    const found = match(search.value);
    hits.textContent = "";
    hits.hidden = found.length === 0;
    for (const a of found) {
      const b = document.createElement("button");
      b.type = "button";
      b.className = "pickhit";
      const i = document.createElement("span");
      i.className = "pi";
      i.textContent = a.id;
      const n = document.createElement("span");
      n.className = "pn";
      n.textContent = a.name || a.class || "";
      b.appendChild(i);
      b.appendChild(n);
      b.addEventListener("click", () => addRow(a.id));
      hits.appendChild(b);
    }
  }

  // Adding a line splits the weights evenly; every field stays editable
  // afterwards and nothing is ever re-normalized behind the visitor's back.
  function addRow(id) {
    if (rows.some(r => r.id === id) || rows.length >= PICK_MAXHOLDINGS) return;
    rows.push({id, w: 0});
    const even = Math.floor(100 / rows.length);
    rows.forEach((r, i) => { r.w = i === 0 ? 100 - even * (rows.length - 1) : even; });
    search.value = "";
    hits.hidden = true;
    hits.textContent = "";
    renderRows();
    search.focus();
  }

  function renderRows() {
    rowsEl.textContent = "";
    rows.forEach((r, i) => {
      const row = document.createElement("div");
      row.className = "pickrow";
      const id = document.createElement("span");
      id.className = "pid";
      id.textContent = r.id;
      const w = document.createElement("input");
      w.type = "number";
      w.className = "pw";
      w.min = "0";
      w.max = "100";
      w.step = "1";
      w.value = String(r.w);
      w.setAttribute("aria-label", "Weight of " + r.id + " in percent");
      w.addEventListener("input", () => {
        r.w = Math.round(parseFloat(w.value));
        refresh();
      });
      const pc = document.createElement("span");
      pc.className = "ppc";
      pc.textContent = "%";
      const del = document.createElement("button");
      del.type = "button";
      del.className = "pdel";
      del.textContent = "×";
      del.setAttribute("aria-label", "Remove " + r.id);
      del.addEventListener("click", () => { rows.splice(i, 1); renderRows(); });
      row.appendChild(id);
      row.appendChild(w);
      row.appendChild(pc);
      row.appendChild(del);
      rowsEl.appendChild(row);
    });
    refresh();
  }

  // spec is the p= value: the holdings, then "!sim:on" so the panel splices
  // each holding's reconstructed history. A retirement-length model wants the
  // deepest past available, and real quotes alone are usually far too short.
  const spec = () => rows.map(r => r.id + ":" + r.w).join(",") + "!sim:on";

  // refresh recomputes the weight sum, the spec echo and why the button is
  // (or is not) live. The reasons mirror the server's gates; it re-checks.
  function refresh() {
    const sum = rows.reduce((a, r) => a + (isFinite(r.w) ? r.w : 0), 0);
    sumEl.textContent = rows.length ? "weights sum to " + sum + "%" : "";
    sumEl.className = "picksum" + (rows.length && sum !== 100 ? " off" : "");
    const s = spec();
    specEl.textContent = rows.length ? s : "";
    let reason = "";
    if (!rows.length) reason = "pick at least one asset above";
    else if (rows.some(r => !isFinite(r.w) || r.w <= 0)) reason = "every line needs a weight above zero";
    else if (rows.length > PICK_MAXHOLDINGS) reason = "at most " + PICK_MAXHOLDINGS + " holdings";
    else if (new TextEncoder().encode(s).length > PICK_MAXBYTES) reason = "this composition is too long to fit a link";
    load.disabled = reason !== "";
    why.textContent = reason;
  }

  search.addEventListener("input", renderHits);
  search.addEventListener("keydown", e => {
    if (e.key === "Escape") { search.value = ""; renderHits(); return; }
    if (e.key !== "Enter") return;
    e.preventDefault();
    const found = match(search.value);
    if (found.length) addRow(found[0].id);
  });
  load.addEventListener("click", () => {
    if (load.disabled) return;
    gotoMount(picker.base + "/p/" + encP(spec()) + "/");
  });
  refresh();
  return wrap;
}

// ---------------------------------------------------------------------------
// Portfolio mode bootstrap: fetch holdings, seed the fit, add the bar.
// ---------------------------------------------------------------------------
fetch(apiURL("/api/meta")).then(r => r.json()).then(m => {
  renderCape(m.cape);
  setSVG("capeHistory", m.capeHistory);
  renderMarketPill(m);
  if (!m.hasPanel) {
    // Nothing on the page would otherwise reveal that portfolio mode exists:
    // the drawer's Allocation slot gets the loader instead.
    renderPortfolioGroup(m);
    run();
    runSlow();
    return;
  }
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

  const box = document.createElement("div");
  box.className = "group";
  box.id = "group-allocation";
  box.innerHTML = `<div class="group-h">Allocation</div>
    <div class="ctl span" data-help="Drag a divider to shift weight between adjacent holdings. Every model re-fits (μ/σ/df and the historical panel) from the live weights.">
      <span class="lab"><span>Drag a divider to shift weight</span></span>
      <div class="allocbar" id="allocbar"></div><div class="alloclegend" id="alloclegend"></div>
    </div>`;
  // Allocation slots into the first column, under the situation and above
  // the simulation settings.
  const sim = document.getElementById("group-simulation");
  sim.parentElement.insertBefore(box, sim);
  renderAlloc();
  run();
  runSlow();
});




