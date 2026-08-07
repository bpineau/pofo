// The live composer's front end (served at /composer.js).
//
// This is the skeleton: it bootstraps from the data attributes the server
// rendered into the panel and leaves the panel in a clean, usable static
// state (the native <details> toggles collapsed/open on its own). The editing
// behaviour, the /catalog.json autocomplete, live validation, the link-budget
// meter and the "Run comparison" navigation land in the next task; init() is
// the stub they fill.
//
// Everything runs inside one try/catch so a bootstrap failure degrades to the
// server-rendered static panel rather than a broken page: the composer is a
// convenience over the /view URL grammar, never a hard dependency.
(function () {
  "use strict";

  // read parses a JSON data attribute, returning null on a missing or
  // malformed value rather than throwing (the server writes it, but a proxy
  // or an extension could still corrupt it).
  function read(el, name) {
    var raw = el.getAttribute(name);
    if (!raw) return null;
    try {
      return JSON.parse(raw);
    } catch (e) {
      return null;
    }
  }

  // boot reads the server-rendered state off the panel: the guardrail caps
  // and, for each forkable read-only card, its data-fork-<i> payload
  // ({name, p, dropped}). It returns the parsed model init() will drive.
  function boot(panel) {
    var caps = read(panel, "data-caps") || {};
    var forks = [];
    var cards = panel.querySelectorAll(".pcard");
    for (var i = 0; i < cards.length; i++) {
      var fork = read(cards[i], "data-fork-" + i);
      if (fork) forks.push({ index: i, card: cards[i], fork: fork });
    }
    return { panel: panel, caps: caps, forks: forks };
  }

  // init wires the interactive behaviour. Skeleton stub: the next task fills
  // it (fork -> editable rows, autocomplete, validation, budget meter, run).
  function init(state) {
    // Intentionally empty for now. `state` carries the parsed panel model.
    void state;
  }

  function main() {
    var panel = document.getElementById("composer");
    if (!panel) return; // not a composer page
    init(boot(panel));
  }

  try {
    if (document.readyState === "loading") {
      document.addEventListener("DOMContentLoaded", function () {
        try {
          main();
        } catch (e) {
          /* leave the static panel in place */
        }
      });
    } else {
      main();
    }
  } catch (e) {
    /* bootstrap failed: the server-rendered panel stays usable */
  }
})();
