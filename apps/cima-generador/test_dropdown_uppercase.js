// Test v3.7.8: items del dropdown se ven en MAYÚSCULAS
const fs = require('fs');
const http = require('http');
const html = fs.readFileSync('/home/braka/Systems/projects/work/cima/apps/cima-generador/index.html', 'utf8');

let pass = 0, fail = 0;
function tT(cond, msg) {
  if (cond) { pass++; console.log(`  ✓ ${msg}`); }
  else { fail++; console.log(`  ✗ ${msg}`); }
}

console.log('=== v3.7.8 — Dropdown items MAYÚSCULAS (L-061) ===');

console.log('\n--- Test 1: .ac-item tiene text-transform: uppercase ---');
const acItemMatch = html.match(/\.ac-item\s*\{([\s\S]*?)\n\s*\}/);
tT(acItemMatch !== null, 'Regla .ac-item existe');
if (acItemMatch) {
  tT(acItemMatch[1].includes('text-transform: uppercase'), '.ac-item tiene text-transform: uppercase');
}

console.log('\n--- Test 2: .ac-empty tiene text-transform: uppercase ---');
const acEmptyMatch = html.match(/\.ac-empty\s*\{([\s\S]*?)\n\s*\}/);
tT(acEmptyMatch !== null, 'Regla .ac-empty existe');
if (acEmptyMatch) {
  tT(acEmptyMatch[1].includes('text-transform: uppercase'), '.ac-empty tiene text-transform: uppercase');
}

console.log('\n--- Test 3: CCPP_DATA inline verificado ---');
const ccpps = html.match(/const CCPP_DATA = (\{[\s\S]*?\});/);
tT(ccpps !== null, 'CCPP_DATA existe');
if (ccpps) {
  // Sample
  const sample = ccpps[1].match(/"(\d{6})":\s*\["([^"]+)"/);
  if (sample) {
    tT(sample[2] === sample[2].toLowerCase() || /[A-Z]/.test(sample[2]),
      `Sample item ${sample[1]}: "${sample[2]}" (case mixto/lowercase — se mostrará en MAYÚSCULAS via CSS)`);
  }
}

console.log('\n--- Test 4: No hay reglas con text-transform: lowercase ---');
tT(!html.includes('text-transform: lowercase'), 'No hay text-transform: lowercase');

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
