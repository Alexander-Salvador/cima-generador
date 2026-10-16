const fs = require('fs');
const http = require('http');
const html = fs.readFileSync('/home/braka/Systems/projects/work/cima/apps/cima-generador/index.html', 'utf8');

let pass = 0, fail = 0;
function tT(cond, msg) {
  if (cond) { pass++; console.log(`  ✓ ${msg}`); }
  else { fail++; console.log(`  ✗ ${msg}`); }
}

const mapMatch = html.match(/AC_STATE\.items\.map\(function \(item, i\)\s*\{([\s\S]*?)\}\)\.join\(''\)/);
const itemBody = mapMatch[1];
const renderItem = new Function(
  'item', 'i', 'query',
  'd6ForMark', 'distritoEnApagadoSinMatch', 'ZONAS_APAGADAS',
  'acHighlight', 'acEscape',
  itemBody
);

const ZONAS_APAGADAS = {
  ccpp_keys: new Set(['160109:CHILIQUIN', '160109:BELLAVISTA']),
  districts_with_off: new Set(['160109'])
};
const acHighlight = (t) => t;
const acEscape = (s) => s;

console.log('=== v3.7.1 — Render del item CCPP ===');

const h1 = renderItem('Pueblo Nuevo', 0, '', null, false, ZONAS_APAGADAS, acHighlight, acEscape);
const h2 = renderItem('Santa María', 0, '', '160109', true, ZONAS_APAGADAS, acHighlight, acEscape);
const h3 = renderItem('CHILIQUIN', 0, '', '160109', true, ZONAS_APAGADAS, acHighlight, acEscape);

console.log('\n--- Test 1: Item normal ---');
tT(h1.includes('class="ac-item"'), 'clase base ac-item');
tT(!h1.includes('ac-off-dot'), 'NO tiene dot');
tT(!h1.includes('ac-item-off'), 'NO tiene ac-item-off');
tT(!h1.includes('ac-item-zone-off'), 'NO tiene zone-off');

console.log('\n--- Test 2: Item zone-off ---');
tT(h2.includes('ac-item-zone-off'), 'clase ac-item-zone-off');
tT(h2.includes('ac-off-dot'), 'dot indicator');
tT(h2.includes('title="Distrito con apagados Movistar"'), 'title atributo');
tT(!h2.includes('⚠️'), 'NO emoji ⚠️');

console.log('\n--- Test 3: Item match exacto ---');
tT(h3.includes('ac-item-off'), 'clase ac-item-off');
tT(h3.includes('ac-off-dot'), 'dot indicator');
tT(h3.includes('title="CCPP exacto en zona de apagado Movistar"'), 'title atributo');
tT(!h3.includes('⚠️'), 'NO emoji ⚠️');

console.log('\n--- Test 4: Estructura <div> con role=option data-idx data-val ---');
tT(h1.startsWith('<div class="ac-item" role="option" data-idx="0" data-val="Pueblo Nuevo">'), 'Item normal: estructura completa');
tT(h2.startsWith('<div class="ac-item ac-item-zone-off" role="option" data-idx="0" data-val="Santa María">'), 'Item zone-off: estructura completa');
tT(h3.startsWith('<div class="ac-item ac-item-off" role="option" data-idx="0" data-val="CHILIQUIN">'), 'Item match exacto: estructura completa');

console.log('\n--- Test 5: Dot entre data-val y texto del item ---');
const dataValIdx = h3.indexOf('data-val="');
const dataValEnd = h3.indexOf('">', dataValIdx) + 2;
const dotIdx = h3.indexOf('ac-off-dot');
const itemTextIdx = h3.indexOf('CHILIQUIN', dataValEnd);
tT(dotIdx > dataValEnd, 'dot está DESPUÉS del data-val');
tT(dotIdx < itemTextIdx, 'dot está ANTES del texto');

console.log('\n--- Test 6: data-val correcto ---');
tT(h1.includes('data-val="Pueblo Nuevo"'), 'data-val normal');
tT(h2.includes('data-val="Santa María"'), 'data-val zone-off');
tT(h3.includes('data-val="CHILIQUIN"'), 'data-val match exacto');

console.log('\n--- Test 7: Sin emoji ⚠️ (reemplazado por dot) ---');
tT(!h1.includes('⚠️') && !h2.includes('⚠️') && !h3.includes('⚠️'), 'Ninguno de los 3 outputs tiene emoji');

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
