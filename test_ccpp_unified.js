// Test v3.7.4: CCPP dropdown IDÉNTICO a depto/prov/dist, solo items match exacto con dot
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

const mapMatch = html.match(/AC_STATE\.items\.map\(function \(item, i\)\s*\{([\s\S]*?)\}\)\.join\(''\)/);
const itemBody = mapMatch[1];
const renderItem = new Function(
  'item', 'i', 'query',
  'd6ForMark', 'ZONAS_APAGADAS',
  'acHighlight', 'acEscape',
  itemBody
);

const ZONAS_APAGADAS = {
  ccpp_keys: new Set(['160109:CHILIQUIN']),
  districts_with_off: new Set(['160109'])
};
const acEscape = (s) => s;
const acHighlight = (t) => t;

console.log('=== v3.7.4 — CCPP dropdown IDÉNTICO a depto/prov/dist ===');

console.log('\n--- Test 1: CCPP dropdown en distrito APAGADO (IQUITOS) ---');
console.log('  Items NO match exacto: deben verse NORMALES (como depto/prov/dist)');
const h1 = renderItem('Bellavista', 0, '', '160109', ZONAS_APAGADAS, acHighlight, acEscape);
console.log('  Output:', h1);
tT(h1 === '<div class="ac-item" role="option" data-idx="0" data-val="Bellavista">Bellavista</div>', 'Item normal en distrito apagado: clase base ac-item');
tT(!h1.includes('ac-item-off'), 'NO tiene ac-item-off');
tT(!h1.includes('ac-item-zone-off'), 'NO tiene ac-item-zone-off (eliminado)');
tT(!h1.includes('ac-off-dot'), 'NO tiene dot indicator');

console.log('\n--- Test 2: CCPP dropdown en distrito NO apagado (otro) ---');
const h2 = renderItem('Bellavista', 0, '', '010101', ZONAS_APAGADAS, acHighlight, acEscape);
tT(h2 === '<div class="ac-item" role="option" data-idx="0" data-val="Bellavista">Bellavista</div>', 'Item normal en distrito NO apagado: clase base ac-item');
tT(!h2.includes('ac-off-dot'), 'NO tiene dot');

console.log('\n--- Test 3: Item match EXACTO con ccpp_keys ---');
const h3 = renderItem('CHILIQUIN', 0, '', '160109', ZONAS_APAGADAS, acHighlight, acEscape);
console.log('  Output:', h3);
tT(h3.includes('ac-item-off'), 'Match exacto: clase ac-item-off (rojo)');
tT(h3.includes('ac-off-dot'), 'Match exacto: dot indicator');
tT(h3.includes('title="CCPP exacto en zona de apagado Movistar"'), 'Match exacto: title atributo');
tT(!h3.includes('ac-item-zone-off'), 'Match exacto: NO zone-off (eliminado)');

console.log('\n--- Test 4: Estructura idéntica depto/prov/dist vs CCPP ---');
const deptoItems = [
  ['AMAZONAS', 0],
  ['LIMA', 1],
  ['LORETO', 2]
].map(([n, i]) => renderItem(n, i, '', null, ZONAS_APAGADAS, acHighlight, acEscape));
// Departamento no recibe d6ForMark ni ZONAS_APAGADAS — pero el código actual del render
// SI los recibe como null. Verificamos que devuelve estructura estándar.
console.log('  Items depto:', deptoItems);
tT(deptoItems.every(h => h.startsWith('<div class="ac-item" role="option"')), 'Todos los items depto inician con <div class="ac-item" role="option"');
tT(deptoItems.every(h => !h.includes('ac-item-off')), 'Items depto NO tienen ac-item-off');
tT(deptoItems.every(h => !h.includes('ac-item-zone-off')), 'Items depto NO tienen ac-item-zone-off');
tT(deptoItems.every(h => !h.includes('ac-off-dot')), 'Items depto NO tienen dot indicator');

console.log('\n--- Test 5: CSS de .ac-item-zone-off eliminada ---');
tT(!html.includes('.ac-item.ac-item-zone-off'), 'CSS .ac-item-zone-off eliminada');
tT(!html.includes('background: color-mix(in srgb, #f59e0b 6%, transparent)'), 'CSS zone-off bg eliminada');
tT(!html.includes('color: #b87b0a'), 'CSS zone-off color eliminada');
tT(!html.includes('ac-item-zone-off:hover'), 'CSS zone-off hover eliminada');

console.log('\n--- Test 6: CSS de .ac-item-off PRESERVADA (match exacto) ---');
tT(html.includes('.ac-item.ac-item-off'), 'CSS .ac-item-off preservada');
tT(html.includes('color: var(--danger)'), 'CSS item-off color preservada');
tT(html.includes('background: color-mix(in srgb, var(--danger) 8%, transparent)'), 'CSS item-off bg preservada (8% danger)');
tT(html.includes('.ac-off-dot'), 'CSS .ac-off-dot preservada');

console.log('\n--- Test 7: CSS de .ac-item base IDÉNTICA (4 dropdowns) ---');
const acItemMatch = html.match(/\.ac-item\s*\{[^}]+\}/);
tT(acItemMatch !== null, 'CSS .ac-item existe');
tT(acItemMatch[0].includes('padding: 6px 12px'), 'Padding 6px 12px');
tT(acItemMatch[0].includes('font-size: 11px'), 'Font-size 11px');
tT(acItemMatch[0].includes('display: flex'), 'display: flex');
tT(acItemMatch[0].includes('gap: 8px'), 'gap 8px');

console.log('\n--- Test 8: Servidor local ---');
http.get('http://127.0.0.1:8765/index.html', (res) => {
  tT(res.statusCode === 200, 'GET /index.html → 200');
  res.resume();
  res.on('end', () => {
    console.log(`\n${'='.repeat(50)}`);
    console.log(`${pass} PASS / ${fail} FAIL`);
    process.exit(fail > 0 ? 1 : 0);
  });
});
