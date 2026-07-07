"use strict";
// The only deployment knob of the static site: where the burn API lives.
// Empty string means same-origin (genesis.hearth.tech/api behind one proxy).
// On localhost the API is assumed on :8080 (cmd/api's default listenAddr);
// localStorage.setItem("hearthApiBase", ...) still overrides everything.
window.HEARTH_API_BASE = localStorage.getItem("hearthApiBase")
  ?? (["localhost", "127.0.0.1"].includes(location.hostname) ? "http://" + location.hostname + ":8080" : "");
