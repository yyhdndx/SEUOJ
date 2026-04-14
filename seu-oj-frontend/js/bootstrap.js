// Frontend bootstrap is loaded last so route modules are available before first render.
window.addEventListener("hashchange", renderRouteSafely);
window.addEventListener("beforeunload", () => {
  stopSubmissionPolling();
  stopSubmissionsPolling();
  stopContestPolling();
});
window.addEventListener("error", (event) => {
  renderFatalError(event.error || new Error(event.message || "Unknown runtime error"), "runtime");
});
window.addEventListener("unhandledrejection", (event) => {
  const reason = event.reason instanceof Error ? event.reason : new Error(String(event.reason || "Unhandled rejection"));
  renderFatalError(reason, "promise");
});

if (document.readyState === "loading") {
  document.addEventListener("DOMContentLoaded", bootstrapApp, { once: true });
} else {
  bootstrapApp();
}



