window.onload = function() {
  //<editor-fold desc="Changeable Configuration Block">

  // Spec location can be overridden via the ?url= query parameter.
  var params = new URLSearchParams(window.location.search);
  var specUrl = params.get("url") || "./openapi.json";

  window.ui = SwaggerUIBundle({
    url: specUrl,
    dom_id: '#swagger-ui',
    deepLinking: true,
    docExpansion: 'list',
    defaultModelsExpandDepth: -1,
    displayOperationId: true,
    displayRequestDuration: true,
    filter: true,
    tryItOutEnabled: true,
    persistAuthorization: true,
    presets: [
      SwaggerUIBundle.presets.apis,
      SwaggerUIStandalonePreset
    ],
    plugins: [
      SwaggerUIBundle.plugins.DownloadUrl
    ],
    layout: "StandaloneLayout"
  });

  //</editor-fold>
};

// Report the rendered document height to the parent page so the docs route can
// be a single-scroll document (auto-sized iframe) instead of two scrollbars.
// The reporting page and the parent embed are cross-origin (frontend/backend),
// so postMessage must be used; height data alone is harmless to broadcast.
(function reportHeightToParent() {
  if (window.parent === window) return;

  var last = -1;
  function report(force) {
    var h = Math.ceil(document.documentElement.scrollHeight);
    if (!force && h === last) return;
    last = h;
    window.parent.postMessage({ type: "swagger:height", height: h }, "*");
  }

  window.addEventListener("load", function () {
    // Swagger UI renders async; catch the settled height.
    setTimeout(report, 100);
    setTimeout(report, 500);
  });

  if (window.ResizeObserver) {
    // Covers expand/collapse, filtering, auth modal and window resizes.
    new ResizeObserver(report).observe(document.documentElement);
  }

  // Fallback heartbeat, posted every 750ms regardless of height changes. The
  // parent attaches its listener only after React hydration, so it may miss
  // the one-off reports fired while the frame's height was settling; forcing
  // a post here guarantees a late-listening parent still receives the size.
  setInterval(function () { report(true); }, 750);
})();