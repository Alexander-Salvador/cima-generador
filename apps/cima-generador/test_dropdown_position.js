// Test v3.7.6: dropdown SIEMPRE abajo del input, sin flip
const fs = require('fs');
const http = require('http');
const html = fs.readFileSync('/home/braka/Systems/projects/work/cima/apps/cima-generador/index.html', 'utf8');

let pass = 0, fail = 0;
function tT(cond, msg) {
  if (cond) { pass++; console.log(`  ✓ ${msg}`); }
  else { fail++; console.log(`  ✗ ${msg}`); }
}

function extractFunc(name) {
  const startIdx = html.indexOf(`function ${name}(`);
  if (startIdx < 0) return null;
  let i = html.indexOf('{', startIdx);
  let depth = 0;
  for (let j = i; j < Math.min(i + 10000, html.length); j++) {
    if (html[j] === '{') depth++;
    else if (html[j] === '}') {
      depth--;
      if (depth === 0) return html.substring(startIdx, j + 1);
    }
  }
  return null;
}

console.log('=== v3.7.6 — Dropdown SIEMPRE abajo (sin flip) ===');

console.log('\n--- Test 1: Lógica de flip eliminada de acFitToGrid ---');
tT(!html.includes('FLIP_THRESHOLD'), 'No hay FLIP_THRESHOLD');
tT(!html.includes('flipCap'), 'No hay variable flipCap');
tT(!html.includes('rect.top - 4 - flipCap'), 'No hay fórmula de flip-arriba');
tT(!html.includes("classList.add('is-flipped')"), 'No hay classList.add(is-flipped)');

console.log('\n--- Test 2: acFitToGrid SIEMPRE abajo ---');
const acFitFn = extractFunc('acFitToGrid');
tT(acFitFn !== null, 'Función acFitToGrid existe');
tT(acFitFn.includes("rect.bottom + 4"), 'top = rect.bottom + 4 (debajo)');
tT(acFitFn.includes("classList.remove('is-flipped')"), 'SIEMPRE remueve is-flipped');

console.log('\n--- Test 3: Sim con espacio limitado (CCPP en row-2) ---');
const acFitSim = new Function('drop', 'input', 'inputRect', `
  var rect = inputRect || input.getBoundingClientRect();
  var grid = input.closest('.grid-block');
  var gridBottom = grid ? grid.getBoundingClientRect().bottom : window.innerHeight;
  var MARGIN = 8;
  var available = gridBottom - rect.bottom - MARGIN;
  var cap = Math.max(80, Math.min(220, available));
  drop.style.top = (rect.bottom + 4) + 'px';
  drop.style.maxHeight = cap + 'px';
  drop.classList.remove('is-flipped');
  return { top: drop.style.top, maxHeight: drop.style.maxHeight, removed: drop._removed || false };
`);
const drop = { style: {}, classList: { remove: () => { drop._removed = true } } };
const input = { getBoundingClientRect: () => ({ top: 680, bottom: 700, left: 0, right: 100, width: 100, height: 20 }) };
const grid = { getBoundingClientRect: () => ({ top: 0, bottom: 750, left: 0, right: 100, width: 100, height: 750 }) };
input.closest = () => grid;
const r = acFitSim(drop, input, null);
console.log('  Sim: input bottom=700, grid bottom=750, available=42');
console.log('  Result:', r);
tT(r.top === '704px', 'top = 704px (debajo del input)');
tT(r.maxHeight === '80px', 'maxHeight = 80px (mínimo, antes flip)');
tT(r.removed === true, 'is-flipped removido (no flip)');

console.log('\n--- Test 4: HTML del input CCPP ---');
const ccppInput = html.match(/<input id="ccpp"[^>]+>/);
tT(ccppInput !== null, 'Input CCPP existe');
tT(!ccppInput[0].includes('style='), 'No tiene style inline (no se reposiciona)');

console.log('\n--- Test 5: Servidor local ---');
http.get('http://127.0.0.1:8765/index.html', (res) => {
  tT(res.statusCode === 200, 'GET /index.html → 200');
  res.resume();
  res.on('end', () => {
    console.log(`\n${'='.repeat(50)}`);
    console.log(`${pass} PASS / ${fail} FAIL`);
    process.exit(fail > 0 ? 1 : 0);
  });
});
