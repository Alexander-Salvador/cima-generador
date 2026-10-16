// Test v3.7.5: footer "X de Y" eliminado del dropdown
const fs = require('fs');
const http = require('http');
const html = fs.readFileSync('/home/braka/Systems/projects/work/cima/apps/cima-generador/index.html', 'utf8');

let pass = 0, fail = 0;
function tT(cond, msg) {
  if (cond) { pass++; console.log(`  ✓ ${msg}`); }
  else { fail++; console.log(`  ✗ ${msg}`); }
}

console.log('=== v3.7.5 — Footer "X de Y" eliminado ===');

console.log('\n--- Test 1: No hay código de generación del footer ---');
tT(!html.includes("AC_STATE.items.length + ' de ' + total"), 'No hay código JS que genere "X de Y"');
tT(!html.includes('<div class="ac-meta">'), 'No hay template <div class="ac-meta">');

console.log('\n--- Test 2: CSS .ac-meta eliminada ---');
tT(!html.match(/\.ac-meta\s*\{/), 'CSS .ac-meta { eliminada');

console.log('\n--- Test 3: Render de items sigue funcionando ---');
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
const mapMatch = html.match(/AC_STATE\.items\.map\(function \(item, i\)\s*\{([\s\S]*?)\}\)\.join\(''\)/);
const renderItem = new Function('item', 'i', 'query', 'd6ForMark', 'ZONAS_APAGADAS', 'acHighlight', 'acEscape', mapMatch[1]);
const ZONAS_APAGADAS = { ccpp_keys: new Set(), districts_with_off: new Set() };
const acEscape = (s) => s;
const acHighlight = (t) => t;
const h = renderItem('LIMA', 0, '', null, ZONAS_APAGADAS, acHighlight, acEscape);
tT(h === '<div class="ac-item" role="option" data-idx="0" data-val="LIMA">LIMA</div>', 'Item normal: formato correcto sin footer');
tT(!h.includes(' de '), 'Output NO contiene " de " (footer eliminado)');

console.log('\n--- Test 4: Servidor local ---');
http.get('http://127.0.0.1:8765/index.html', (res) => {
  tT(res.statusCode === 200, 'GET /index.html → 200');
  res.resume();
  res.on('end', () => {
    console.log(`\n${'='.repeat(50)}`);
    console.log(`${pass} PASS / ${fail} FAIL`);
    process.exit(fail > 0 ? 1 : 0);
  });
});
