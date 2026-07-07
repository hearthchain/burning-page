"use strict";
// The only deployment knob of the static site: where the burn API lives.
// Empty string means same-origin (genesis.hearth.tech/api behind one proxy).
// For local development, point it elsewhere once per browser:
//   localStorage.setItem("hearthApiBase", "http://localhost:8080")
window.HEARTH_API_BASE = localStorage.getItem("hearthApiBase") ?? "";
