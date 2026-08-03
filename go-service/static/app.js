// Matches the Go formatter exactly so a value does not change shape
// when the countup animation finishes.
function fmtINR(n) {
  n = Number(n);
  if (Math.abs(n) < 0.5) n = 0;
  return "Rs " + Math.round(n).toLocaleString("en-IN");
}

// Short form, used only for slider readouts where space is tight.
function fmtShort(n) {
  n = Number(n);
  var abs = Math.abs(n);
  if (abs >= 1e7) return "Rs " + (n / 1e7).toFixed(2) + " Cr";
  if (abs >= 1e5) return "Rs " + (n / 1e5).toFixed(2) + " L";
  return "Rs " + Math.round(n).toLocaleString("en-IN");
}

function showTab(name) {
  document.getElementById('calcTab').style.display = name === 'calc' ? '' : 'none';
  document.getElementById('suggestionTabWrapper').style.display = name === 'suggestion' ? '' : 'none';
  document.getElementById('tabCalcBtn').classList.toggle('active', name === 'calc');
  document.getElementById('tabSuggBtn').classList.toggle('active', name === 'suggestion');
}

function reducedMotion() {
  return window.matchMedia && window.matchMedia('(prefers-reduced-motion: reduce)').matches;
}

function animateCount(el) {
  var raw = el.getAttribute('data-value');
  if (raw === null) return;
  var target = Number(raw);
  if (!isFinite(target)) return;
  if (reducedMotion()) {
    el.textContent = fmtINR(target);
    return;
  }
  var dur = 700;
  var t0 = performance.now();
  function step(t) {
    var x = Math.min(1, (t - t0) / dur);
    var e = 1 - Math.pow(1 - x, 3);
    el.textContent = fmtINR(target * e);
    if (x < 1) requestAnimationFrame(step);
  }
  requestAnimationFrame(step);
}

function runCountups(root) {
  var scope = root && root.querySelectorAll ? root : document;
  scope.querySelectorAll('.countup').forEach(animateCount);
}

if (document.readyState === 'loading') {
  document.addEventListener('DOMContentLoaded', function () { runCountups(document); });
} else {
  runCountups(document);
}

document.addEventListener('htmx:afterSwap', function (e) {
  runCountups(e.target);
  runCountups(document);
});
