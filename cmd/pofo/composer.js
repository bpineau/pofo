// The live composer's front end (served at /composer.js).
//
// Two layers, kept apart on purpose:
//
//   1. A pure-function core over a plain state object: parseSearch (a /view
//      query string -> state), serialize (state -> query string, the exact
//      inverse for non-opaque state), and byteLen (UTF-8 length of one p=
//      value). These never touch the DOM and never throw: an unparseable p=
//      degrades to an {kind:"opaque", raw} port whose raw is passed through
//      verbatim. composerSelfTest exercises their round trips.
//
//   2. A thin DOM layer that hydrates the server-rendered panel, keeps the
//      state object in sync with every edit, and rewrites the URL live
//      (history.replaceState) so the link always reflects the editor.
//
// Everything runs inside one try/catch so a bootstrap failure degrades to the
// server-rendered static panel rather than a broken page: the composer is a
// convenience over the /view URL grammar, never a hard dependency. The
// server-side gates in view.go stay authoritative; the client caps only warn
// early.
//
// state = {
//   ports: [ {kind:"ex", name}
//          | {kind:"p", raw, name, nameSet, holdings:[{id,w}], metas:{k:v}}
//          | {kind:"opaque", raw} ],
//   globals: {currency, rebalance, sim, bench, start, end}  // "" = server default
// }
(function () {
  "use strict";

  // ---- pure core ----------------------------------------------------------

  // encP encodes one query value the way the /view grammar is written by hand:
  // ":", "," and "!" stay literal (they are legal in a query and keep links
  // readable), everything else is percent-encoded. It is the inverse of the
  // decoding URLSearchParams performs on the way in.
  function encP(v) {
    return encodeURIComponent(v).replace(/%3A/g, ":").replace(/%2C/g, ",");
  }

  // byteLen is the UTF-8 byte length of one serialized p= value, the unit the
  // server caps at maxViewSpecLen.
  function byteLen(p) {
    return new TextEncoder().encode(p).length;
  }

  // normCurrency folds the currency value onto the grammar's tokens: a present
  // but empty value or "native" (any case) is the native sentinel; a code is
  // upper-cased; absent stays "".
  function normCurrency(v) {
    var u = v.trim().toUpperCase();
    if (u === "" || u === "NATIVE") return "native";
    return u;
  }

  // parsePValue parses one decoded p= value ("ID:WEIGHT,...!name:..!k:v..")
  // into a p port, or an opaque port when it does not fit the grammar. Shared
  // by parseSearch and Fork so both accept exactly what the server accepts.
  function parsePValue(raw) {
    var segs = raw.split("!");
    var holdings = [];
    var pairs = segs[0].split(",");
    for (var i = 0; i < pairs.length; i++) {
      var pair = pairs[i].trim();
      if (pair === "") continue;
      var c = pair.indexOf(":");
      if (c < 0) return { kind: "opaque", raw: raw };
      var id = pair.slice(0, c).trim();
      var w = pair.slice(c + 1).trim();
      if (id === "" || w === "" || !isFinite(Number(w))) return { kind: "opaque", raw: raw };
      holdings.push({ id: id, w: w });
    }
    if (holdings.length === 0) return { kind: "opaque", raw: raw };
    var name = "", nameSet = false, metas = {};
    for (var s = 1; s < segs.length; s++) {
      var meta = segs[s].trim();
      if (meta === "") continue;
      var mc = meta.indexOf(":");
      if (mc < 0) return { kind: "opaque", raw: raw };
      var k = meta.slice(0, mc).trim(), v = meta.slice(mc + 1).trim();
      if (k === "name") { name = v; nameSet = v !== ""; continue; }
      metas[k] = v;
    }
    return { kind: "p", raw: raw, name: name, nameSet: nameSet, holdings: holdings, metas: metas };
  }

  // parseSearch turns a /view query string into a state object. Ports are
  // grouped exs-first then ps, mirroring the server's parseViewQuery so the
  // client ports line up one-to-one with the rendered cards.
  function parseSearch(search) {
    var q = new URLSearchParams(search);
    var ports = [];
    q.getAll("ex").forEach(function (v) { ports.push({ kind: "ex", name: v }); });
    q.getAll("p").forEach(function (v) { ports.push(parsePValue(v)); });
    var g = { currency: "", rebalance: "", sim: "", bench: "", start: "", end: "" };
    if (q.has("currency")) g.currency = normCurrency(q.get("currency"));
    if (q.get("rebalance")) g.rebalance = q.get("rebalance");
    if (q.get("sim")) g.sim = q.get("sim");
    if (q.get("bench")) g.bench = q.get("bench");
    if (q.get("start")) g.start = q.get("start");
    if (q.get("end")) g.end = q.get("end");
    return { ports: ports, globals: g };
  }

  // serializeP renders one port back into its p=/ex= value. Opaque raws pass
  // through verbatim; a p port rebuilds holdings, then the explicit name, then
  // its metas in encounter order.
  function serializeP(port) {
    if (port.kind === "opaque") return port.raw;
    var s = port.holdings.map(function (h) { return h.id + ":" + h.w; }).join(",");
    if (port.nameSet && port.name !== "") s += "!name:" + port.name;
    Object.keys(port.metas).forEach(function (k) { s += "!" + k + ":" + port.metas[k]; });
    return s;
  }

  // serialize is the exact inverse of parseSearch for non-opaque state: ports
  // in order (ex= then p=), then any set globals in a fixed order. Empty
  // globals are left out so the server default applies.
  function serialize(state) {
    var parts = [];
    state.ports.forEach(function (port) {
      if (port.kind === "ex") parts.push("ex=" + encP(port.name));
      else parts.push("p=" + encP(serializeP(port)));
    });
    var g = state.globals;
    if (g.currency) parts.push("currency=" + encP(g.currency));
    if (g.rebalance) parts.push("rebalance=" + encP(g.rebalance));
    if (g.sim) parts.push("sim=" + encP(g.sim));
    if (g.bench) parts.push("bench=" + encP(g.bench));
    if (g.start) parts.push("start=" + encP(g.start));
    if (g.end) parts.push("end=" + encP(g.end));
    return parts.join("&");
  }

  // composerSelfTest round-trips serialize(parseSearch(x)) === x on the cases
  // the brief enumerates. Runs only from #composer-selftest; Task 7 drives it
  // headless. Reports each case to the console.
  function composerSelfTest() {
    var cases = [
      "p=NTSX:46,VWCE:30!name:Core!rebalance:30&p=IWDA:60,IGLN:40!sim:on",
      "p=40%20VUAA,30%20IB01",
      "ex=claude-dragonlite&p=IWDA:60,IGLN:40",
      "ex=claude-dragonlite&currency=EUR&rebalance=90&sim=on&bench=IWDA&start=2010-01-01&end=2026-06-30",
      ""
    ];
    var ok = 0;
    cases.forEach(function (x) {
      var got = serialize(parseSearch(x));
      var pass = got === x;
      if (pass) ok++;
      console.log((pass ? "PASS" : "FAIL") + " " + JSON.stringify(x) + (pass ? "" : " -> " + JSON.stringify(got)));
    });
    console.log("composerSelfTest: " + ok + "/" + cases.length + " passed");
    return ok === cases.length;
  }

  // ---- DOM layer ----------------------------------------------------------

  var CAP_PORTS = 6, CAP_HOLD = 20, CAP_BYTES = 2000; // overwritten from data-caps
  var state = { ports: [], globals: {} };
  var catalog = null;          // [{id,name,class,alt}] once /catalog.json loads
  var known = null;            // Set of lower-cased ids and alts for validation
  var byKey = null;            // lower-cased id/alt -> asset, for fill and name
  var panelEl = null;

  // commit rewrites the URL to the current state without a navigation.
  function commit() {
    history.replaceState(null, "", "/view?" + serialize(state));
    refreshBudget();
  }

  // run navigates to the freshly serialized comparison.
  function run() {
    location.assign("/view?" + serialize(state));
  }

  // portOf climbs from an event target to the state port its card carries.
  function portOf(node) {
    var card = node.closest ? node.closest(".pcard") : null;
    return card ? card.__cmpPort : null;
  }

  // fmtNum renders a weight without trailing zeros ("46", "33.33").
  function fmtNum(n) { return String(Number(n.toFixed(2))); }

  // sumOf totals a port's weights (numbers, blanks as zero).
  function sumOf(port) {
    return port.holdings.reduce(function (a, h) { return a + (Number(h.w) || 0); }, 0);
  }

  // syncCard reads the editable name and holding rows of a p card back into its
  // port. Metas and the ex/opaque raw are never in the DOM, so they are left
  // untouched.
  function syncCard(port) {
    var el = port._el;
    var nameInput = el.querySelector(".pname");
    if (nameInput && !nameInput.readOnly) port.name = nameInput.value.trim();
    var hs = [];
    el.querySelectorAll(".hrow").forEach(function (r) {
      hs.push({
        id: r.querySelector(".idbox .field").value.trim(),
        w: r.querySelector(".wt").value.trim()
      });
    });
    port.holdings = hs;
  }

  // ---- rendering the editable rows ----------------------------------------

  // makeRow builds one editable holding row for a p card.
  function makeRow(h) {
    var row = document.createElement("div");
    row.className = "hrow";
    var idbox = document.createElement("div");
    idbox.className = "idbox";
    var id = document.createElement("input");
    id.className = "field id";
    id.value = h.id;
    id.setAttribute("autocomplete", "off");
    id.setAttribute("spellcheck", "false");
    idbox.appendChild(id);
    var rn = document.createElement("span");
    rn.className = "rn";
    var wt = document.createElement("input");
    wt.className = "field wt";
    wt.setAttribute("inputmode", "decimal");
    wt.value = h.w;
    var rm = document.createElement("button");
    rm.className = "rm";
    rm.type = "button";
    rm.textContent = "×";
    row.appendChild(idbox);
    row.appendChild(rn);
    row.appendChild(wt);
    row.appendChild(rm);
    refreshName(id, rn);
    validateId(id);
    return row;
  }

  // renderBody repaints a p card's body from its port: one editable row per
  // holding, plus the add-holding button.
  function renderBody(port) {
    var body = port._el.querySelector(".pcard-body");
    body.textContent = "";
    port.holdings.forEach(function (h) { body.appendChild(makeRow(h)); });
    var add = document.createElement("button");
    add.className = "add";
    add.type = "button";
    add.textContent = "+ add holding";
    body.appendChild(add);
  }

  // enhanceHead adds the live sum badge, the Normalize button and a remove
  // control to an editable card head (the server renders the plain head).
  function enhanceHead(port) {
    var head = port._el.querySelector(".pcard-head");
    if (!head || head.querySelector(".bal")) return;
    var bal = document.createElement("span");
    bal.className = "bal";
    var norm = document.createElement("button");
    norm.className = "norm";
    norm.type = "button";
    norm.textContent = "Normalize to 100";
    var drop = document.createElement("button");
    drop.className = "pdrop";
    drop.type = "button";
    drop.title = "Remove this portfolio";
    drop.textContent = "×";
    head.appendChild(bal);
    head.appendChild(norm);
    head.appendChild(drop);
    refreshBadge(port);
  }

  // refreshBadge updates a card's sum badge, flagging any sum that is not 100.
  function refreshBadge(port) {
    var bal = port._el.querySelector(".bal");
    if (!bal) return;
    var s = sumOf(port);
    bal.textContent = "Σ " + fmtNum(s);
    var ok = Math.abs(s - 100) < 0.005;
    bal.classList.toggle("ok", ok);
    bal.classList.toggle("off", !ok);
  }

  // normalize rescales a card's weights to sum 100, rounding to two decimals
  // and letting the last row absorb the rounding residue.
  function normalize(port) {
    var vals = port.holdings.map(function (h) { return Number(h.w) || 0; });
    var total = vals.reduce(function (a, b) { return a + b; }, 0);
    if (total <= 0 || port.holdings.length === 0) return;
    var scaled = vals.map(function (v) { return Math.round((v / total) * 10000) / 100; });
    var s = scaled.reduce(function (a, b) { return a + b; }, 0);
    var last = scaled.length - 1;
    scaled[last] = Math.round((scaled[last] + (100 - s)) * 100) / 100;
    var rows = port._el.querySelectorAll(".hrow .wt");
    port.holdings.forEach(function (h, i) {
      h.w = fmtNum(scaled[i]);
      if (rows[i]) rows[i].value = h.w;
    });
    refreshBadge(port);
    commit();
  }

  // ---- autocomplete + validation ------------------------------------------

  // refreshName fills a row's readout with the id's catalog name once known.
  function refreshName(input, rn) {
    if (!byKey) { rn.textContent = ""; return; }
    var a = byKey[input.value.trim().toLowerCase()];
    rn.textContent = a ? a.name : (input.value.trim() ? "unknown identifier" : "");
  }

  // validateId reds an id input whose value is not in the catalog (only once
  // the catalog has loaded; a failed fetch means no validation at all).
  function validateId(input) {
    if (!known) { input.classList.remove("bad"); return; }
    var v = input.value.trim().toLowerCase();
    input.classList.toggle("bad", v !== "" && !known.has(v));
  }

  var acBox = null, acInput = null, acItems = [], acPos = -1;

  // closeAC dismisses the autocomplete dropdown.
  function closeAC() {
    if (acBox && acBox.parentNode) acBox.parentNode.removeChild(acBox);
    acBox = null; acInput = null; acItems = []; acPos = -1;
  }

  // openAC shows catalog matches for the focused id input: a case-insensitive
  // substring over id, name and alternates. Keyboard and mouse both select.
  function openAC(input) {
    closeAC();
    if (!catalog) return;
    var q = input.value.trim().toLowerCase();
    if (q === "") return;
    var hits = [];
    for (var i = 0; i < catalog.length && hits.length < 8; i++) {
      var a = catalog[i];
      var hay = (a.id + " " + (a.name || "") + " " + (a.alt || []).join(" ")).toLowerCase();
      if (hay.indexOf(q) >= 0) hits.push(a);
    }
    hits.sort(function (x, y) {
      var sx = x.id.toLowerCase().indexOf(q) === 0 ? 0 : 1;
      var sy = y.id.toLowerCase().indexOf(q) === 0 ? 0 : 1;
      return sx - sy;
    });
    if (hits.length === 0) return;
    acBox = document.createElement("div");
    acBox.className = "ac";
    acItems = hits;
    hits.forEach(function (a, i) {
      var d = document.createElement("div");
      if (i === 0) d.className = "on";
      var t = document.createElement("span");
      t.className = "t";
      t.textContent = a.id;
      var dd = document.createElement("span");
      dd.className = "d";
      dd.textContent = a.name || "";
      d.appendChild(t);
      d.appendChild(dd);
      d.__pick = a;
      acBox.appendChild(d);
    });
    acPos = 0;
    acInput = input;
    input.parentNode.appendChild(acBox);
  }

  // moveAC walks the highlighted row up or down.
  function moveAC(delta) {
    if (!acBox) return;
    var rows = acBox.children;
    acPos = (acPos + delta + rows.length) % rows.length;
    for (var i = 0; i < rows.length; i++) rows[i].className = i === acPos ? "on" : "";
  }

  // pickAC fills the input with a catalog id, refreshes the row and commits.
  function pickAC(input, asset) {
    input.value = asset.id;
    var row = input.closest(".hrow");
    refreshName(input, row.querySelector(".rn"));
    validateId(input);
    closeAC();
    var port = portOf(input);
    if (port) { syncCard(port); refreshBadge(port); commit(); }
  }

  // ---- link budget --------------------------------------------------------

  // refreshBudget drives the foot meter off the largest p= value: the server
  // caps each p= at CAP_BYTES, so the binding portfolio sets the reading.
  function refreshBudget() {
    if (!panelEl) return;
    var max = 0;
    state.ports.forEach(function (port) {
      if (port.kind !== "ex") max = Math.max(max, byteLen(serializeP(port)));
    });
    var budget = panelEl.querySelector(".budget");
    var bytes = panelEl.querySelector(".cmp-bytes");
    var meter = panelEl.querySelector(".budget .meter i");
    var hint = panelEl.querySelector(".cmp-hint");
    if (bytes) bytes.textContent = max.toLocaleString() + " / " + CAP_BYTES.toLocaleString() + " B";
    if (meter) meter.style.width = Math.min(100, (max / CAP_BYTES) * 100) + "%";
    var amber = max > CAP_BYTES - 200, red = max >= CAP_BYTES;
    if (budget) {
      budget.classList.toggle("warn", amber && !red);
      budget.classList.toggle("over", red);
    }
    if (hint) {
      hint.textContent = red ? "At the shareable-link cap. Trim a portfolio."
        : amber ? "Near the shareable-link cap." : "";
    }
  }

  // overCap reports whether a port's serialized p= exceeds the byte cap.
  function overCap(port) {
    return port.kind !== "ex" && byteLen(serializeP(port)) > CAP_BYTES;
  }

  // ---- globals ------------------------------------------------------------

  // buildGlobals replaces the server-rendered global controls with the exact
  // grammar tokens: selects for currency, rebalance and sim (each with a
  // leading "server default" option), text inputs for bench, start and end.
  function buildGlobals() {
    var row = panelEl.querySelector(".globals");
    if (!row) return;
    row.textContent = "";
    var gt = document.createElement("span");
    gt.className = "gt";
    gt.textContent = "Globals";
    row.appendChild(gt);
    var g = state.globals;
    row.appendChild(selectField("currency", "currency", g.currency, [
      ["", "default"], ["EUR", "EUR"], ["USD", "USD"], ["GBP", "GBP"], ["CHF", "CHF"], ["native", "native"]
    ]));
    row.appendChild(selectField("rebalance", "rebalance", g.rebalance, [
      ["", "default"], ["30", "30 d"], ["90", "90 d"], ["180", "180 d"], ["365", "365 d"], ["0", "never"]
    ]));
    row.appendChild(selectField("sim", "sim", g.sim, [
      ["", "default"], ["on", "on"], ["off", "off"]
    ]));
    row.appendChild(inputField("benchmark", "bench", g.bench, "", "9rem"));
    row.appendChild(inputField("start", "start", g.start, "YYYY-MM-DD", "8rem"));
    row.appendChild(inputField("end", "end", g.end, "YYYY-MM-DD", "8rem"));
  }

  // selectField builds one labelled <select> global control.
  function selectField(label, key, value, opts) {
    var f = document.createElement("div");
    f.className = "gfield";
    var l = document.createElement("label");
    l.textContent = label;
    var sel = document.createElement("select");
    sel.className = "field";
    opts.forEach(function (o) {
      var opt = document.createElement("option");
      opt.value = o[0];
      opt.textContent = o[1];
      if (o[0] === value) opt.selected = true;
      sel.appendChild(opt);
    });
    sel.addEventListener("change", function () {
      state.globals[key] = key === "currency" ? normCurrency(sel.value) : sel.value;
      commit();
    });
    f.appendChild(l);
    f.appendChild(sel);
    return f;
  }

  // inputField builds one labelled text global control.
  function inputField(label, key, value, placeholder, width) {
    var f = document.createElement("div");
    f.className = "gfield";
    var l = document.createElement("label");
    l.textContent = label;
    var inp = document.createElement("input");
    inp.className = "field";
    inp.value = value;
    if (placeholder) inp.placeholder = placeholder;
    if (width) inp.style.width = width;
    inp.addEventListener("input", function () { state.globals[key] = inp.value.trim(); commit(); });
    f.appendChild(l);
    f.appendChild(inp);
    return f;
  }

  // ---- fork ---------------------------------------------------------------

  // fork replaces a read-only ex= card with an editable p= card parsed from its
  // data-fork payload, surfaces any dropped content as a dismissible note and
  // rewrites the URL.
  function fork(port, payload) {
    var parsed = parsePValue(payload.p);
    if (parsed.kind !== "p") return;
    if (payload.name) { parsed.name = payload.name; parsed.nameSet = true; }
    var el = port._el;
    // Swap the state port in place, keeping the DOM index alignment.
    for (var k in port) if (Object.prototype.hasOwnProperty.call(port, k)) delete port[k];
    Object.assign(port, parsed);
    port._el = el;
    el.__cmpPort = port;
    el.removeAttribute("data-fork-" + indexOfPort(port));

    var kind = el.querySelector(".kind");
    if (kind) { kind.textContent = "p="; kind.className = "kind p"; }
    var name = el.querySelector(".pname");
    if (name) { name.readOnly = false; name.value = parsed.nameSet ? parsed.name : ""; }
    var forkBtn = el.querySelector(".fork");
    if (forkBtn) forkBtn.parentNode.removeChild(forkBtn);

    renderBody(port);
    enhanceHead(port);
    if (payload.dropped && payload.dropped.length) addNote(el, payload.dropped);
    commit();
  }

  // indexOfPort returns a port's position (its data-fork index).
  function indexOfPort(port) {
    return state.ports.indexOf(port);
  }

  // addNote shows the dropped-content list under a forked card, dismissible.
  function addNote(el, dropped) {
    var body = el.querySelector(".pcard-body");
    var note = document.createElement("div");
    note.className = "note";
    var txt = document.createElement("span");
    txt.textContent = "Dropped on fork: " + dropped.join(", ");
    var x = document.createElement("button");
    x.className = "note-x";
    x.type = "button";
    x.textContent = "×";
    x.addEventListener("click", function () { note.parentNode.removeChild(note); });
    note.appendChild(txt);
    note.appendChild(x);
    body.parentNode.insertBefore(note, body);
  }

  // ---- add / remove -------------------------------------------------------

  // addHolding appends a blank editable row within the holdings cap.
  function addHolding(port) {
    if (port.holdings.length >= CAP_HOLD) return;
    port.holdings.push({ id: "", w: "" });
    var body = port._el.querySelector(".pcard-body");
    body.insertBefore(makeRow({ id: "", w: "" }), body.querySelector(".add"));
    refreshBadge(port);
    commit();
  }

  // removeHolding drops one row.
  function removeHolding(port, row) {
    row.parentNode.removeChild(row);
    syncCard(port);
    refreshBadge(port);
    commit();
  }

  // removePort drops a whole portfolio card (opaque and editable alike).
  function removePort(port) {
    var i = state.ports.indexOf(port);
    if (i >= 0) state.ports.splice(i, 1);
    if (port._el && port._el.parentNode) port._el.parentNode.removeChild(port._el);
    commit();
  }

  // ---- wiring -------------------------------------------------------------

  // accepted remembers the last within-cap value of each field, so an edit
  // that would push a p= past the byte cap can be reverted (the input refused).
  var accepted = new WeakMap();

  // wire attaches the delegated listeners once, after hydration.
  function wire() {
    // Weight, id and name edits, refusing any edit past the byte cap.
    panelEl.addEventListener("input", function (e) {
      var t = e.target;
      var port = portOf(t);
      if (!port) return;
      var wasName = t.classList.contains("pname");
      if (wasName) port.nameSet = true;
      syncCard(port);
      if (overCap(port)) {
        // Refuse: restore the field to its last accepted value.
        if (accepted.has(t)) t.value = accepted.get(t);
        syncCard(port);
        refreshBudget();
        return;
      }
      accepted.set(t, t.value);
      if (t.classList.contains("id")) {
        validateId(t);
        refreshName(t, t.closest(".hrow").querySelector(".rn"));
        openAC(t);
      }
      if (t.classList.contains("wt") || t.classList.contains("id")) refreshBadge(port);
      commit();
    });

    panelEl.addEventListener("focusin", function (e) {
      if (e.target.classList && e.target.classList.contains("id")) openAC(e.target);
      if (e.target.classList && e.target.classList.contains("field")) accepted.set(e.target, e.target.value);
      if (e.target.classList && e.target.classList.contains("pname")) accepted.set(e.target, e.target.value);
    });
    panelEl.addEventListener("focusout", function (e) {
      if (e.target.classList && e.target.classList.contains("id")) {
        // Delay so a mousedown on a suggestion registers before the box closes.
        setTimeout(closeAC, 120);
      }
    });

    panelEl.addEventListener("mousedown", function (e) {
      var item = e.target.closest ? e.target.closest(".ac div") : null;
      if (item && item.__pick && acInput) { e.preventDefault(); pickAC(acInput, item.__pick); }
    });

    panelEl.addEventListener("click", function (e) {
      var t = e.target;
      var port = portOf(t);
      if (t.classList.contains("rm") && port) { removeHolding(port, t.closest(".hrow")); return; }
      if (t.classList.contains("add") && port) { addHolding(port); return; }
      if (t.classList.contains("norm") && port) { normalize(port); return; }
      if (t.classList.contains("pdrop") && port) { removePort(port); return; }
      if (t.classList.contains("fork") && port && port._fork) { fork(port, port._fork); return; }
      if (t.classList.contains("btn-run")) { run(); return; }
    });

    panelEl.addEventListener("keydown", function (e) {
      var t = e.target;
      if (t.classList && t.classList.contains("id") && acBox) {
        if (e.key === "ArrowDown") { e.preventDefault(); moveAC(1); return; }
        if (e.key === "ArrowUp") { e.preventDefault(); moveAC(-1); return; }
        if (e.key === "Enter") {
          e.preventDefault();
          var pick = acBox.children[acPos] && acBox.children[acPos].__pick;
          if (pick) pickAC(t, pick);
          return;
        }
        if (e.key === "Escape") { closeAC(); return; }
      }
      if (e.key === "Enter" && t.classList && (t.classList.contains("field") || t.classList.contains("pname"))) {
        e.preventDefault();
        run();
      }
    });
  }

  // ---- boot ---------------------------------------------------------------

  // hydrate binds the parsed state to the server-rendered cards (ex-first then
  // p, one-to-one), wires the forkable payloads, and paints the live controls.
  function hydrate(panel, forks) {
    var caps = readJSON(panel, "data-caps") || {};
    CAP_PORTS = caps.ports || CAP_PORTS;
    CAP_HOLD = caps.holdings || CAP_HOLD;
    CAP_BYTES = caps.bytes || CAP_BYTES;
    panelEl = panel;
    state = parseSearch(location.search);
    var cards = panel.querySelectorAll(".pcard");
    for (var i = 0; i < state.ports.length && i < cards.length; i++) {
      var port = state.ports[i];
      port._el = cards[i];
      cards[i].__cmpPort = port;
      if (port.kind === "p") { renderBody(port); enhanceHead(port); }
    }
    forks.forEach(function (f) {
      if (f.card.__cmpPort) f.card.__cmpPort._fork = f.fork;
    });
    buildGlobals();
    wire();
    refreshBudget();
    loadCatalog();
  }

  // loadCatalog fetches /catalog.json and lights up autocomplete and inline id
  // validation. A failed fetch leaves both off (never a false red).
  function loadCatalog() {
    fetch("/catalog.json").then(function (r) { return r.json(); }).then(function (data) {
      catalog = data;
      known = new Set();
      byKey = {};
      data.forEach(function (a) {
        known.add(a.id.toLowerCase());
        byKey[a.id.toLowerCase()] = a;
        (a.alt || []).forEach(function (alt) { known.add(alt.toLowerCase()); byKey[alt.toLowerCase()] = a; });
      });
      panelEl.querySelectorAll(".hrow").forEach(function (r) {
        var id = r.querySelector(".idbox .field");
        validateId(id);
        refreshName(id, r.querySelector(".rn"));
      });
    }).catch(function () { /* no catalog: no validation, no autocomplete */ });
  }

  // readJSON parses a JSON data attribute, null on missing or malformed value.
  function readJSON(el, name) {
    var raw = el.getAttribute(name);
    if (!raw) return null;
    try { return JSON.parse(raw); } catch (e) { return null; }
  }

  // boot collects the caps and each forkable card's data-fork payload.
  function boot(panel) {
    var forks = [];
    var cards = panel.querySelectorAll(".pcard");
    for (var i = 0; i < cards.length; i++) {
      var fork = readJSON(cards[i], "data-fork-" + i);
      if (fork) forks.push({ index: i, card: cards[i], fork: fork });
    }
    return { panel: panel, forks: forks };
  }

  function main() {
    if (location.hash === "#composer-selftest") { composerSelfTest(); return; }
    var panel = document.getElementById("composer");
    if (!panel) return; // not a composer page
    var b = boot(panel);
    hydrate(b.panel, b.forks);
  }

  try {
    if (document.readyState === "loading") {
      document.addEventListener("DOMContentLoaded", function () {
        try { main(); } catch (e) { /* leave the static panel in place */ }
      });
    } else {
      main();
    }
  } catch (e) {
    /* bootstrap failed: the server-rendered panel stays usable */
  }
})();
