// Slider definitions: [key, label, min, max, step, default].
const SLIDERS = [
  ["capital","Capital (€)",800000,4000000,10000,1800000],
  ["needAnnual","Spending floor /yr (€)",24000,84000,1000,48000],
  ["bufferYears","Buffer (years)",0,10,1,3],
  ["mu","Real growth return",0.01,0.07,0.005,0.045],
  ["sigma","Volatility",0.06,0.20,0.005,0.12],
  ["df","Tail df (low=fat)",3,30,1,6],
  ["bufferReturn","Buffer real return",-0.01,0.02,0.005,0.005],
  ["years","Horizon (years)",20,45,1,40],
  ["pensionYear","Pension from year",5,20,1,12],
  ["pensionAnnual","Pension /yr (€)",0,36000,1000,12000],
  ["flexCut","Possible spending cut",0,0.40,0.05,0.25],
  ["taxRate","Flat tax on gains",0,0.35,0.01,0.314],
  ["nPaths","Simulations",500,5000,500,2000],
];
const form = document.getElementById("controls");
const state = {};
for (const [k,label,min,max,step,def] of SLIDERS) {
  state[k] = def;
  const d = document.createElement("label"); d.className = "ctl";
  d.innerHTML = `${label}: <span id="v_${k}">${def}</span>
    <input type="range" min="${min}" max="${max}" step="${step}" value="${def}" id="s_${k}">`;
  form.appendChild(d);
  d.querySelector("input").addEventListener("input", e => {
    state[k] = parseFloat(e.target.value);
    document.getElementById("v_"+k).textContent = e.target.value;
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
  for (const id of ["bufferSvg","ruinCurveSvg","surfaceSvg","recoverySvg"])
    document.getElementById(id).innerHTML = r[id];
  document.getElementById("cards").innerHTML = Object.entries(r.cards)
    .map(([k,v]) => `<div class="card"><div>${k}</div><div class="v">${v}</div></div>`).join("");
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
      if (s) { s.value = v; document.getElementById("v_"+k).textContent = v.toFixed(3); }
    }
  }
  const sel = document.createElement("label"); sel.className="ctl";
  sel.innerHTML = `Return model:
    <select id="model"><option value="parametric">parametric</option>
    <option value="bootstrap">historical bootstrap</option>
    <option value="cohorts">historical cohorts</option></select>`;
  form.prepend(sel);
  sel.querySelector("select").addEventListener("change", e=>{state.model=e.target.value;schedule();});
  state.model = "parametric";
  labels.forEach((name,i)=>{
    const d=document.createElement("label"); d.className="ctl";
    d.innerHTML=`${name}: <span id="w_${i}">${Math.round(weights[i]*100)}</span>%
      <input type="range" min="0" max="100" step="1" value="${Math.round(weights[i]*100)}" id="al_${i}">`;
    form.appendChild(d);
    d.querySelector("input").addEventListener("input",e=>{
      weights[i]=parseFloat(e.target.value)/100;
      document.getElementById("w_"+i).textContent=e.target.value; schedule();});
  });
  run();
});
