// Slider definitions: [key, label, min, max, step, default, unit].
// unit drives how the live value is shown: pct (×100, "%"), eur, or int.
const SLIDERS = [
  ["capital","Capital",800000,4000000,10000,1800000,"eur"],
  ["needAnnual","Net spending /yr",24000,84000,1000,60000,"eur"],
  ["bufferYears","Buffer (years)",0,10,1,3,"int"],
  ["mu","Real growth return",0.01,0.12,0.005,0.05,"pct"],
  ["sigma","Volatility (long-horizon)",0.06,0.20,0.005,0.11,"pct"],
  ["df","Tail df (low=fat)",3,30,1,5,"int"],
  ["bufferReturn","Buffer real return",-0.01,0.05,0.005,0.005,"pct"],
  ["years","Horizon (years from today)",20,60,1,45,"int"],
  ["pensionYear","Pension from year",5,20,1,12,"int"],
  ["pensionAnnual","Pension /yr",0,36000,1000,12000,"eur"],
  ["flexCut","Spending cut in downturns (0 = fixed rule)",0,0.40,0.05,0,"pct"],
  ["taxRate","Flat tax on gains",0,0.35,0.01,0.314,"pct"],
  ["sideAnnual","Side income /yr",0,30000,1000,0,"eur"],
  ["sideUntilYear","Side income until year",0,20,1,0,"int"],
  ["bufferStopYear","Buffer refill stop (yr, 0=never)",0,20,1,0,"int"],
  ["nPaths","Simulations",500,5000,500,2000,"int"],
];

const FMT = {
  pct: v => (v * 100).toFixed(1).replace(/\.0$/, "") + "%",
  eur: v => Math.round(v).toLocaleString("fr-FR") + " €",
  int: v => String(Math.round(v)),
};
const UNIT = {};
for (const [k, , , , , , unit] of SLIDERS) UNIT[k] = unit;
const fmtVal = (k, v) => (FMT[UNIT[k] || "int"])(v);
const PAL = ["#1f77b4","#ff7f0e","#2ca02c","#d62728","#9467bd","#8c564b","#e377c2","#17becf"];

// portfolio-mode state, set once /api/meta resolves.
let weights = null, labels = [], hasPanel = false, lastFitW = null;

const cardsHTML = cards => (cards || [])
  .map(c => `<div class="card"><div class="k">${c.label}</div><div class="v">${c.value}</div></div>`).join("");

const form = document.getElementById("controls");
const state = {};
for (const [k, label, min, max, step, def] of SLIDERS) {
  state[k] = def;
  const d = document.createElement("label"); d.className = "ctl";
  d.innerHTML = `<span class="lab"><span>${label}</span><span class="val" id="v_${k}">${fmtVal(k, def)}</span></span>
    <input type="range" min="${min}" max="${max}" step="${step}" value="${def}" id="s_${k}">`;
  form.appendChild(d);
  d.querySelector("input").addEventListener("input", e => {
    state[k] = parseFloat(e.target.value);
    document.getElementById("v_" + k).textContent = fmtVal(k, state[k]);
    schedule();
  });
}
function setSliderVal(k, v) {
  state[k] = v;
  const s = document.getElementById("s_" + k);
  if (s) { s.value = v; document.getElementById("v_" + k).textContent = fmtVal(k, v); }
}

// Monthly-withdrawal toggle: step the kernel monthly (salary-like) instead of
// once a year.
state.monthly = false;
const monthlyCtl = document.createElement("label"); monthlyCtl.className = "ctl span chk";
monthlyCtl.innerHTML = `<input type="checkbox" id="monthly"> <span>Monthly withdrawals (salary-like)</span>`;
form.appendChild(monthlyCtl);
monthlyCtl.querySelector("input").addEventListener("change", e => { state.monthly = e.target.checked; schedule(); });

// Stress regimes: a two-state Markov source where bad years cluster, adding the
// sequence risk and fatter left tail that i.i.d. draws miss (annual steps).
state.regime = false;
const regimeCtl = document.createElement("label"); regimeCtl.className = "ctl span chk";
regimeCtl.innerHTML = `<input type="checkbox" id="regime"> <span>Sequence-risk stress: cluster bad years into drawdowns</span>`;
form.appendChild(regimeCtl);
regimeCtl.querySelector("input").addEventListener("change", e => { state.regime = e.target.checked; schedule(); });

// Guyton-Klinger guardrails: adjust spending to a band around the initial
// withdrawal rate, instead of the single drawdown-triggered flex cut.
state.guardrails = false;
const guardCtl = document.createElement("label"); guardCtl.className = "ctl span chk";
guardCtl.innerHTML = `<input type="checkbox" id="guardrails"> <span>Guyton-Klinger guardrails (replaces flex cut)</span>`;
form.appendChild(guardCtl);
guardCtl.querySelector("input").addEventListener("change", e => { state.guardrails = e.target.checked; schedule(); });

// Conservative broad-sample prior: override the (often rosy) fitted/default
// mu/sigma/df with cautious, forward-looking world-equity real assumptions.
// Lower real return, higher volatility, fatter tails than a favourable window.
// Matches the server's Conservative column (web.consMu/consSigma/consDf): a
// ~3.5% real geometric mean with fat tails, not an "equities barely grow" 1.4%.
const PRIOR = {mu: 0.045, sigma: 0.13, df: 4};
const DEFAULT = Object.fromEntries(SLIDERS.map(([k, , , , , def]) => [k, def]));
function applyReturns(src) { for (const k of ["mu", "sigma", "df"]) setSliderVal(k, src[k]); }
state.conservative = false;
const consCtl = document.createElement("label"); consCtl.className = "ctl span chk";
consCtl.innerHTML = `<input type="checkbox" id="conservative"> <span>Broad-sample prior (override the fit)</span>`;
form.appendChild(consCtl);
consCtl.querySelector("input").addEventListener("change", e => {
  state.conservative = e.target.checked;
  if (state.conservative) applyReturns(PRIOR);
  else if (hasPanel) lastFitW = null; // force a refit from the panel
  else applyReturns(DEFAULT);
  schedule();
});

// --- shareable scenarios: the whole slider/model/allocation state round-trips
// through the URL hash, so a configuration can be bookmarked or shared. ---
const shared = new URLSearchParams(location.hash.slice(1));
const sharedWeights = shared.has("w")
  ? shared.get("w").split(",").map(Number).filter(x => !isNaN(x)) : null;
// Apply any shared slider values up front (portfolio-mode seeding re-applies
// them after the fit so the shared values still win).
function applySharedSliders() {
  for (const [k] of SLIDERS)
    if (shared.has(k)) { const v = parseFloat(shared.get(k)); if (!isNaN(v)) setSliderVal(k, v); }
}
applySharedSliders();
if (shared.get("monthly") === "1") {
  state.monthly = true;
  monthlyCtl.querySelector("input").checked = true;
}
if (shared.get("guardrails") === "1") {
  state.guardrails = true;
  guardCtl.querySelector("input").checked = true;
}
if (shared.get("conservative") === "1") {
  state.conservative = true;
  consCtl.querySelector("input").checked = true;
  applyReturns(PRIOR);
}
if (shared.get("regime") === "1") {
  state.regime = true;
  regimeCtl.querySelector("input").checked = true;
}
function syncURL() {
  const p = new URLSearchParams();
  for (const [k] of SLIDERS) p.set(k, state[k]);
  if (state.model) p.set("model", state.model);
  if (state.monthly) p.set("monthly", "1");
  if (state.guardrails) p.set("guardrails", "1");
  if (state.conservative) p.set("conservative", "1");
  if (state.regime) p.set("regime", "1");
  if (weights) p.set("w", weights.map(x => x.toFixed(4)).join(","));
  history.replaceState(null, "", "#" + p.toString());
}

let timer = null, lastBody = null;
function schedule(){ clearTimeout(timer); timer = setTimeout(run, 200); }

function weightsChanged() {
  if (!weights || !lastFitW || lastFitW.length !== weights.length) return true;
  return weights.some((w, i) => Math.abs(w - lastFitW[i]) > 1e-9);
}

let run = async function(){
  // In portfolio mode the parametric model reads mu/sigma, not the weights,
  // so a weight change must re-fit mu/sigma from the panel before computing,
  // otherwise dragging the allocation would not move the parametric result.
  if (weights) {
    state.weights = weights;
    if (hasPanel && weightsChanged() && !state.conservative) {
      try {
        const resp = await fetch("/api/fit", {method:"POST",
          headers:{"Content-Type":"application/json"}, body: JSON.stringify({weights})});
        const f = await resp.json();
        if (typeof f.mu === "number") setSliderVal("mu", f.mu);
        if (typeof f.sigma === "number") setSliderVal("sigma", f.sigma);
        if (typeof f.df === "number") setSliderVal("df", f.df);
      } catch (e) { /* keep the current mu/sigma on failure */ }
      lastFitW = weights.slice();
    }
  }
  const body = {...state, years: Math.round(state.years),
    pensionYear: Math.round(state.pensionYear), nPaths: Math.round(state.nPaths),
    sideUntilYear: Math.round(state.sideUntilYear), bufferStopYear: Math.round(state.bufferStopYear),
    targetRuin: (parseFloat(document.getElementById("targetRuin").value) || 5) / 100};
  lastBody = body;
  renderModels(body);   // the multi-model hero strip, in parallel with the detail sim
  renderPaths(body);    // the wealth fan charts, one per planning model
  renderSolver(body);   // the per-lever menu to reach the acceptable ruin
  renderFrontier(body); // ruin vs withdrawal rate, per model
  renderSensitivity(body); // change in ruin per controllable lever
  const res = await fetch("/api/sim",{method:"POST",headers:{"Content-Type":"application/json"},
    body: JSON.stringify(body)});
  const r = await res.json();
  document.getElementById("note").textContent = r.note || "";
  for (const id of ["arbitrageSvg","recoverySvg"])
    document.getElementById(id).innerHTML = r[id] || "";
  document.getElementById("cards").innerHTML = cardsHTML(r.cards);
  syncURL();
};

// --- multi-model hero strip: ruin / safe spend / median wealth per return
// model, the epistemic-uncertainty view that replaces a single ruin figure. ---
const esc = s => (s || "").replace(/&/g, "&amp;").replace(/"/g, "&quot;").replace(/</g, "&lt;").replace(/>/g, "&gt;");

// Instant tooltip for any [data-help] element (native title has a ~1s delay).
const tip = document.createElement("div");
tip.id = "tip";
document.body.appendChild(tip);
document.addEventListener("mouseover", e => {
  const el = e.target.closest("[data-help]");
  if (!el) return;
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
// Ruin colour: green (safe) through amber to red, saturating at 30%.
function ruinColor(r) {
  const x = Math.max(0, Math.min(r, 0.30));
  return `hsl(${(120 * (1 - x / 0.30)).toFixed(0)},65%,88%)`;
}
async function renderModels(body) {
  const target = (parseFloat(document.getElementById("targetRuin").value) || 5) / 100;
  let r;
  try {
    r = await (await fetch("/api/models", {method: "POST",
      headers: {"Content-Type": "application/json"}, body: JSON.stringify({...body, targetRuin: target})})).json();
  } catch (e) { return; }
  document.getElementById("verdict").textContent = r.verdict || "";
  const conf = document.getElementById("confidence");
  conf.textContent = r.confidence ? `Confidence: ${r.confidence} · ${r.confNote}` : "";
  conf.className = r.confidence ? "conf-" + r.confidence.toLowerCase() : "";
  const ms = r.models || [];
  const cells = (fn, attr = "") => ms.map(m => `<td${attr ? " " + attr(m) : ""}>${fn(m)}</td>`).join("");
  const head = `<tr><th></th>${ms.map(m => `<th data-help="${esc(m.help)}">${m.name}</th>`).join("")}</tr>`;
  const ruinRow = `<tr><th data-help="Share of simulated retirements that run out of money, at your planned spend.">Ruin</th>` +
    cells(m => (m.ruin * 100).toFixed(1) + "%", m => `style="background:${ruinColor(m.ruin)}"`) + `</tr>`;
  const spendRow = `<tr><th data-help="The most you could spend per year and still keep ruin at your acceptable level, under this model.">Safe spend</th>` +
    cells(m => `${(m.safeSpend / 1000).toFixed(0)}k€<span class="sub"> ${(m.safeWR * 100).toFixed(1)}%</span>`) + `</tr>`;
  const wealthRow = `<tr><th data-help="Median real wealth left at the end of the horizon, at your planned spend.">Median wealth</th>` +
    cells(m => (m.medianWealth / 1000).toFixed(0) + "k€") + `</tr>`;
  document.getElementById("modelstrip").innerHTML =
    `<table class="modeltab"><thead>${head}</thead><tbody>${ruinRow}${spendRow}${wealthRow}</tbody></table>`;
}
document.getElementById("targetRuin").addEventListener("input", schedule);

// Wealth fan charts: one picture of the simulated market per planning model,
// laid out two per row so the central case and the successive stresses can be
// compared side by side.
async function renderPaths(body) {
  try {
    const r = await (await fetch("/api/paths", {method: "POST",
      headers: {"Content-Type": "application/json"}, body: JSON.stringify(body)})).json();
    document.getElementById("fansGrid").innerHTML =
      (r.fans || []).map(f => `<div class="fan">${f.svg || ""}</div>`).join("");
  } catch (e) { /* leave the previous charts on failure */ }
}

// Ruin vs withdrawal-rate frontier, one curve per model.
async function renderFrontier(body) {
  try {
    const r = await (await fetch("/api/frontier", {method: "POST",
      headers: {"Content-Type": "application/json"}, body: JSON.stringify(body)})).json();
    document.getElementById("frontierSvg").innerHTML = r.frontierSvg || "";
  } catch (e) { /* leave the previous chart on failure */ }
}

// Sensitivity "greeks": the change in ruin from nudging each controllable lever.
async function renderSensitivity(body) {
  try {
    const r = await (await fetch("/api/sensitivity", {method: "POST",
      headers: {"Content-Type": "application/json"}, body: JSON.stringify(body)})).json();
    document.getElementById("sensitivitySvg").innerHTML = r.sensitivitySvg || "";
  } catch (e) { /* leave the previous chart on failure */ }
}

// --- solver menu: the equivalent ways to reach the acceptable ruin, one per
// controllable lever (spend less, cut in downturns, hold a cash buffer). ---
const eur = v => Math.round(v).toLocaleString("fr-FR") + " €";
async function renderSolver(body) {
  const target = (parseFloat(document.getElementById("targetRuin").value) || 5) / 100;
  const box = document.getElementById("solvermenu");
  let m;
  try {
    m = await (await fetch("/api/solvemenu", {method: "POST",
      headers: {"Content-Type": "application/json"}, body: JSON.stringify({...body, targetRuin: target})})).json();
  } catch (e) { return; }
  const head = m.met
    ? `<b>Your plan meets the target</b> (ruin ${(m.currentRuin * 100).toFixed(1)}% ≤ ${(m.targetRuin * 100).toFixed(1)}%):`
    : `<b>To get ruin down to ${(m.targetRuin * 100).toFixed(1)}%</b> (now ${(m.currentRuin * 100).toFixed(1)}%), any one of:`;
  const items = (m.options || []).map(o =>
    `<li class="${o.ok ? "" : "no"}">${o.ok ? "" : "✗ "}<span class="lev">${o.lever}:</span> ${o.text}</li>`).join("");
  box.innerHTML = `<div class="solvehead">${head}</div><ul class="solveopts">${items}</ul>`;
}

// --- allocation bar: drag a divider to move weight between two adjacent
// assets; the total stays at 100 % by construction. ---
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
    seg.innerHTML = `<span>${labels[i]}</span><b>${Math.round(w * 100)}%</b>`;
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
    `<span><i style="background:${PAL[i % PAL.length]}"></i>${n} ${Math.round(weights[i] * 100)}%</span>`).join("");
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

// Portfolio mode: fetch holdings, add a model toggle, help and the bar.
fetch("/api/meta").then(r=>r.json()).then(m=>{
  if(!m.hasPanel) { run(); return; }
  hasPanel = true;
  labels = m.labels;
  weights = (sharedWeights && sharedWeights.length === labels.length) ? sharedWeights.slice()
    : (m.weights && m.weights.length === labels.length) ? m.weights.slice()
    : labels.map(()=>1/labels.length);
  lastFitW = weights.slice(); // mu/sigma already seeded below; avoid a redundant refit
  for (const [k, v] of [["mu", m.mu], ["sigma", m.sigma], ["df", m.df]])
    if (typeof v === "number") setSliderVal(k, v);
  applySharedSliders(); // a shared mu/sigma/df overrides the historical seed
  if (state.conservative) applyReturns(PRIOR); // the prior wins over the fit

  // The hero strip already evaluates every return model side by side, so there
  // is no model to pick: the detail charts below use the central parametric one.
  state.model = "parametric";

  const alloc = document.createElement("div"); alloc.className = "ctl span";
  alloc.innerHTML = `<span class="lab"><span>Allocation — drag a divider to shift weight</span></span>
    <div class="allocbar" id="allocbar"></div><div class="alloclegend" id="alloclegend"></div>`;
  form.prepend(alloc);
  renderAlloc();
  run();
});
