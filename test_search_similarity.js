const fs = require('fs');
const http = require('http');
const html = fs.readFileSync('/home/braka/Systems/projects/work/cima/apps/cima-generador/index.html', 'utf8');

let pass = 0, fail = 0;
function t(actual, expected, msg) {
  const ok = JSON.stringify(actual) === JSON.stringify(expected);
  if (ok) { pass++; console.log(`  ✓ ${msg}`); }
  else { fail++; console.log(`  ✗ ${msg}: actual=${JSON.stringify(actual).slice(0,200)} expected=${JSON.stringify(expected).slice(0,200)}`); }
}
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

const code = `
${extractFunc('normSearch')}
${extractFunc('subsequenceMatch')}
${extractFunc('acHighlight')}
${extractFunc('acFilter')}
${extractFunc('acEscape')}
return { normSearch, subsequenceMatch, acHighlight, acFilter, acEscape };
`;
const lib = (new Function(code))();

console.log('=== v3.7.3 — Búsqueda INTELIGENTE por SCORING (L-056) ===');

console.log('\n--- Test 1: Highlight NUNCA parte el texto ---');
tT(lib.acHighlight('CHILIQUIN', 'C') === 'CHILIQUIN', '1 char: texto plano sin <mark>');
tT(lib.acHighlight('CHILIQUIN', 'CH') === 'CHILIQUIN', '2 chars: texto plano sin <mark>');
tT(lib.acHighlight('CHILIQUIN', 'CHI') === 'CHILIQUIN', '3 chars: texto plano sin <mark>');
tT(lib.acHighlight('San José', 'JOSE') === 'San José', '4 chars: texto plano sin <mark>');
tT(!lib.acHighlight('CHILIQUIN', 'CHI').includes('<mark>'), 'Sin <mark> en el output');
tT(!lib.acHighlight('San José', 'JOSE').includes('<mark>'), 'Sin <mark> en output con tildes');
tT(!lib.acHighlight('CHILIQUIN', '').includes('<mark>'), 'Sin <mark> en query vacío');

console.log('\n--- Test 2: Scoring — startsWith primero (score 0.95) ---');
const depts = ['AMAZONAS', 'ANCASH', 'APURIMAC', 'AREQUIPA', 'AYACUCHO', 'CAJAMARCA', 'CUSCO', 'LORETO'];
t(lib.acFilter(depts, 'CA'), ['CAJAMARCA', 'ANCASH'], 'CA: CAJAMARCA (startsWith 0.95) + ANCASH (contains 0.80)');
tT(lib.acFilter(depts, 'CA')[0] === 'CAJAMARCA', 'CA: CAJAMARCA primero (startsWith)');

console.log('\n--- Test 3: Scoring — AM (Amazonas 1º, luego contains) ---');
const r = lib.acFilter(depts, 'AM');
tT(r[0] === 'AMAZONAS', 'AM: AMAZONAS primero (startsWith)');
tT(r.includes('APURIMAC'), 'AM: APURIMAC contiene "am"');
tT(!r.includes('LORETO'), 'AM: LORETO NO contiene am (correcto)');

console.log('\n--- Test 4: Fuzzy typo (subsequence) ---');
const ccpps = ['San José', 'San Juan', 'San Pedro', 'Santa María', 'San Luis', 'Huancayo', 'Cajamarca', 'Lima', 'Cusco'];
tT(lib.acFilter(ccpps, 'JSE').includes('San José'), 'JSE: matchea San José via subsequence');
tT(lib.acFilter(ccpps, 'SNT').includes('Santa María'), 'SNT: matchea Santa');
tT(lib.acFilter(ccpps, 'LUS').includes('San Luis'), 'LUS: matchea San Luis');
tT(lib.acFilter(ccpps, 'XX').length === 0, 'XX: no matchea (sin x)');

console.log('\n--- Test 5: Sin falsos positivos fuzzy ---');
const test5 = ['AMAZONAS', 'MADRE DE DIOS', 'LORETO', 'UCAYALI'];
tT(!lib.acFilter(test5, 'DM').includes('MADRE DE DIOS'), 'DM NO matchea MADRE (D antes que M)');
tT(lib.acFilter(test5, 'MD').includes('MADRE DE DIOS'), 'MD matchea MADRE (M antes que D)');

console.log('\n--- Test 6: Orden por score (DESC) ---');
// Empezando con 1 char
const r1 = lib.acFilter(depts, 'L');
tT(r1[0] === 'LORETO', 'L: LORETO primero (startsWith 0.95)');
// Luego contains
tT(r1[0] === 'LORETO' && !r1.includes('AMAZONAS'), 'L: solo LORETO matchea (AMAZONAS no contiene L)');

console.log('\n--- Test 7: Empty query → primeros 200 ---');
tT(lib.acFilter(depts, '').length === 8, 'Empty query: todos los items');
tT(lib.acFilter(depts, '').length === 8, 'Empty query: todos los items (8)');

console.log('\n--- Test 8: Items sin match NO aparecen (score 0) ---');
tT(!lib.acFilter(depts, 'XYZ').length, 'XYZ: nada matchea (no XYZ en ningún depto)');
tT(!lib.acFilter(depts, 'QQQ').length, 'QQQ: nada matchea');

console.log('\n--- Test 9: Servidor local ---');
http.get('http://127.0.0.1:8765/index.html', (res) => {
  tT(res.statusCode === 200, 'GET /index.html → 200');
  res.resume();
  res.on('end', () => {
    console.log(`\n${'='.repeat(50)}`);
    console.log(`${pass} PASS / ${fail} FAIL`);
    process.exit(fail > 0 ? 1 : 0);
  });
});
