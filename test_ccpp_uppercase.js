// Test v3.7.7: input CCPP mantiene MAYÚSCULAS explícitamente
const fs = require('fs');
const http = require('http');
const html = fs.readFileSync('/home/braka/Systems/projects/work/cima/apps/cima-generador/index.html', 'utf8');

let pass = 0, fail = 0;
function tT(cond, msg) {
  if (cond) { pass++; console.log(`  ✓ ${msg}`); }
  else { fail++; console.log(`  ✗ ${msg}`); }
}

console.log('=== v3.7.7 — Input CCPP MAYÚSCULAS explícitas (L-060) ===');

console.log('\n--- Test 1: text-transform en clases ccpp-off-* (main) ---');
tT(html.match(/\.field input\.ccpp-off-confirmed,[^}]*?text-transform: uppercase/s) !== null, 
  'ccpp-off-confirmed tiene text-transform: uppercase (main)');
tT(html.match(/\.field input\.ccpp-off-ipt,[^}]*?text-transform: uppercase/s) !== null,
  'ccpp-off-ipt tiene text-transform: uppercase (main)');

console.log('\n--- Test 2: text-transform en clases ccpp-off-* (popup) ---');
tT(html.match(/\.field input\.ccpp-off-confirmed,?\s*\.field input\.ccpp-off-confirmed\.filled\s*\{[\s\S]*?text-transform: uppercase/) !== null,
  'ccpp-off-confirmed tiene text-transform: uppercase (popup)');
tT(html.match(/\.field input\.ccpp-off-ipt,?\s*\.field input\.ccpp-off-ipt\.filled\s*\{[\s\S]*?text-transform: uppercase/) !== null,
  'ccpp-off-ipt tiene text-transform: uppercase (popup)');

console.log('\n--- Test 3: No hay reglas con text-transform: lowercase ---');
tT(!html.includes('text-transform: lowercase'), 'No hay text-transform: lowercase');

console.log('\n--- Test 4: .field input base tiene uppercase ---');
tT(html.match(/\.field input, \.field textarea\s*\{[^}]*?text-transform: uppercase/s) !== null,
  'Regla base .field input tiene text-transform: uppercase');

console.log('\n--- Test 5: Conteo de text-transform: uppercase en CSS ---');
const count = (html.match(/text-transform: uppercase/g) || []).length;
console.log(`  Total: ${count}`);
tT(count >= 15, `Mínimo 15 reglas (era 13, ahora +2) — actual ${count}`);

console.log('\n--- Test 6: Servidor local ---');
http.get('http://127.0.0.1:8765/index.html', (res) => {
  tT(res.statusCode === 200, 'GET /index.html → 200');
  res.resume();
  res.on('end', () => {
    console.log(`\n${'='.repeat(50)}`);
    console.log(`${pass} PASS / ${fail} FAIL`);
    process.exit(fail > 0 ? 1 : 0);
  });
});
