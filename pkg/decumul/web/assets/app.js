// Slider definitions: [key, label, min, max, step, default, unit].
// unit drives how the live value is shown: pct (×100, "%"), eur, or int.
const SLIDERS = [
  ["capital","Capital",800000,4000000,10000,1800000,"eur"],
  ["needAnnual","Net spending /yr",24000,84000,1000,48000,"eur"],
  ["bufferYears","Buffer (years)",0,10,1,3,"int"],
  ["mu","Real growth return",0.01,0.12,0.005,0.045,"pct"],
  ["sigma","Volatility",0.06,0.20,0.005,0.12,"pct"],
  ["df","Tail df (low=fat)",3,30,1,6,"int"],
  ["bufferReturn","Buffer real return",-0.01,0.05,0.005,0.005,"pct"],
  ["years","Horizon (years)",20,45,1,40,"int"],
  ["pensionYear","Pension from year",5,20,1,12,"int"],
  ["pensionAnnual","Pension /yr",0,36000,1000,12000,"eur"],
  ["flexCut","Possible spending cut",0,0.40,0.05,0.25,"pct"],
  ["taxRate","Flat tax on gains",0,0.35,0.01,0.314,"pct"],
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
let weights = null, labels = [], hasPanel = false, lastFitW = null, baseline = null;

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

// --- shareable scenarios: the whole slider/model/allocation state round-trips
// through the URL hash, so a configuration can be bookmarked or shared. ---
const shared = new URLSearchParams(location.hash.slice(1));
const sharedModel = shared.get("model");
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
function syncURL() {
  const p = new URLSearchParams();
  for (const [k] of SLIDERS) p.set(k, state[k]);
  if (state.model) p.set("model", state.model);
  if (state.monthly) p.set("monthly", "1");
  if (weights) p.set("w", weights.map(x => x.toFixed(4)).join(","));
  history.replaceState(null, "", "#" + p.toString());
}

let timer = null;
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
    if (hasPanel && weightsChanged()) {
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
    pensionYear: Math.round(state.pensionYear), nPaths: Math.round(state.nPaths)};
  // A/B: with a pinned baseline allocation, compare it against the current one.
  if (hasPanel && baseline) {
    const r = await (await fetch("/api/compare", {method:"POST",
      headers:{"Content-Type":"application/json"},
      body: JSON.stringify({...body, baselineWeights: baseline})})).json();
    document.getElementById("note").textContent = r.variant.note || r.baseline.note || "";
    document.getElementById("cards").innerHTML =
      `<div class="abcol"><h3>Baseline</h3><div class="cardrow">${cardsHTML(r.baseline.cards)}</div></div>` +
      `<div class="abcol"><h3>Variant (current)</h3><div class="cardrow">${cardsHTML(r.variant.cards)}</div></div>`;
    for (const id of ["arbitrageSvg","recoverySvg"])
      document.getElementById(id).innerHTML = r.variant[id] || "";
    syncURL();
    return;
  }
  const res = await fetch("/api/sim",{method:"POST",headers:{"Content-Type":"application/json"},
    body: JSON.stringify(body)});
  const r = await res.json();
  document.getElementById("note").textContent = r.note || "";
  for (const id of ["arbitrageSvg","recoverySvg"])
    document.getElementById(id).innerHTML = r[id] || "";
  document.getElementById("cards").innerHTML = cardsHTML(r.cards);
  syncURL();
};

// --- solver: required capital for a target ruin, and the ruin-minimising
// buffer at the current capital. ---
const eur = v => Math.round(v).toLocaleString("fr-FR") + " €";
document.getElementById("solveBtn").addEventListener("click", async () => {
  const out = document.getElementById("solveOut");
  out.textContent = "solving…";
  const target = (parseFloat(document.getElementById("targetRuin").value) || 5) / 100;
  const body = {...state, years: Math.round(state.years),
    pensionYear: Math.round(state.pensionYear), nPaths: Math.round(state.nPaths),
    targetRuin: target, weights};
  try {
    const r = await (await fetch("/api/solve", {method:"POST",
      headers:{"Content-Type":"application/json"}, body: JSON.stringify(body)})).json();
    if (r.note) { out.textContent = r.note; return; }
    out.innerHTML = `Capital for ${(r.targetRuin*100).toFixed(1)}% ruin: <b>${eur(r.requiredCapital)}</b>` +
      ` · ruin-minimising buffer: <b>${r.bestBufferYears.toFixed(0)} y</b> (${(r.bestBufferRuin*100).toFixed(1)}% ruin)`;
  } catch (e) { out.textContent = "solve failed"; }
});

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

  const sel = document.createElement("label"); sel.className="ctl span";
  sel.innerHTML = `<span class="lab"><span>Return model</span></span>
    <select id="model"><option value="parametric">parametric</option>
    <option value="bootstrap">historical bootstrap</option>
    <option value="cohorts">historical cohorts</option></select>`;
  form.prepend(sel);
  const MODEL_HELP = {
    parametric: "Draws i.i.d. annual real returns from the mu/sigma sliders above (fat-tailed Student-t). Sliders are seeded from this portfolio's historical ANNUAL real-return dispersion, which is usually below the report's daily-annualised volatility (vol drag / trending); raise sigma toward that headline figure for a more conservative test.",
    bootstrap: "Resamples 2-year blocks of this portfolio's actual monthly real returns (2006→), preserving regimes and cross-asset correlations. Optimistic by construction: anchored to that one favourable historical window.",
    cohorts: "Replays every actual historical start month, no resampling. The most faithful but limited to the available history length, so long horizons may be unavailable.",
  };
  const help = document.getElementById("modelhelp");
  const setHelp = mdl => { help.textContent = MODEL_HELP[mdl] || ""; };
  sel.querySelector("select").addEventListener("change", e=>{state.model=e.target.value;setHelp(state.model);schedule();});
  state.model = sharedModel || "parametric";
  sel.querySelector("select").value = state.model;
  setHelp(state.model);

  const alloc = document.createElement("div"); alloc.className = "ctl span";
  alloc.innerHTML = `<span class="lab"><span>Allocation — drag a divider to shift weight</span></span>
    <div class="allocbar" id="allocbar"></div><div class="alloclegend" id="alloclegend"></div>
    <button type="button" id="pinBtn" class="pinbtn">Pin allocation as baseline (A/B)</button>`;
  form.prepend(alloc);
  alloc.querySelector("#pinBtn").addEventListener("click", e => {
    if (baseline) {
      baseline = null;
      e.target.textContent = "Pin allocation as baseline (A/B)";
    } else {
      baseline = weights.slice();
      e.target.textContent = "Clear baseline (A/B on)";
    }
    schedule();
  });
  renderAlloc();
  run();
});
