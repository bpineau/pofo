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
let timer = null;
function schedule(){ clearTimeout(timer); timer = setTimeout(run, 200); }

let run = async function(){
  if (weights) { state.weights = weights; }
  const body = {...state, years: Math.round(state.years),
    pensionYear: Math.round(state.pensionYear), nPaths: Math.round(state.nPaths)};
  const res = await fetch("/api/sim",{method:"POST",headers:{"Content-Type":"application/json"},
    body: JSON.stringify(body)});
  const r = await res.json();
  document.getElementById("note").textContent = r.note || "";
  for (const id of ["bufferSvg","ruinCurveSvg","surfaceSvg","recoverySvg"])
    document.getElementById(id).innerHTML = r[id] || "";
  document.getElementById("cards").innerHTML = (r.cards || [])
    .map(c => `<div class="card"><div class="k">${c.label}</div><div class="v">${c.value}</div></div>`).join("");
};

// Portfolio mode: fetch holdings, add a model toggle and allocation sliders.
let weights = null, labels = [];
fetch("/api/meta").then(r=>r.json()).then(m=>{
  if(!m.hasPanel) { run(); return; }
  labels = m.labels;
  weights = (m.weights && m.weights.length === labels.length) ? m.weights.slice()
    : labels.map(()=>1/labels.length);
  // seed the return assumptions from the portfolio's own history.
  for (const [k, v] of [["mu", m.mu], ["sigma", m.sigma]]) {
    if (typeof v === "number") {
      state[k] = v;
      const s = document.getElementById("s_"+k);
      if (s) { s.value = v; document.getElementById("v_"+k).textContent = fmtVal(k, v); }
    }
  }
  const sel = document.createElement("label"); sel.className="ctl";
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
  const setHelp = m => { help.textContent = MODEL_HELP[m] || ""; };
  sel.querySelector("select").addEventListener("change", e=>{state.model=e.target.value;setHelp(state.model);schedule();});
  state.model = "parametric";
  setHelp("parametric");
  labels.forEach((name,i)=>{
    const d=document.createElement("label"); d.className="ctl";
    d.innerHTML=`<span class="lab"><span>${name}</span><span class="val" id="w_${i}">${Math.round(weights[i]*100)}%</span></span>
      <input type="range" min="0" max="100" step="1" value="${Math.round(weights[i]*100)}" id="al_${i}">`;
    form.appendChild(d);
    d.querySelector("input").addEventListener("input",e=>{
      weights[i]=parseFloat(e.target.value)/100;
      document.getElementById("w_"+i).textContent=e.target.value+"%"; schedule();});
  });
  run();
});
