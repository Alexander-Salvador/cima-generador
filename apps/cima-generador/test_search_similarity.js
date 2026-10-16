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

console.log('=== v3.7.2 — Búsqueda por SIMILITUD (L-055) ===');

console.log('\n--- Test 1: Highlight NO parte texto con query 1 char (L-055) ---');
tT(lib.acHighlight('CHILIQUIN', 'C') === 'CHILIQUIN', 'C+CHILIQUIN: NO highlight (1 char)');
tT(lib.acHighlight('COMERCIO', 'C') === 'COMERCIO', 'C+COMERCIO: NO highlight');
tT(lib.acHighlight('San José', 'C') === 'San José', 'C+San José: NO highlight (no contiene)');
tT(!lib.acHighlight('CHILIQUIN', 'C').includes('<mark>'), 'Sin <mark> en query 1 char');
tT(!lib.acHighlight('COMERCIO', 'L').includes('<mark>'), 'Sin <mark> en L (1 char)');

console.log('\n--- Test 2: Highlight con query 2+ chars ---');
t(lib.acHighlight('CHILIQUIN', 'CH'), '<mark>CH</mark>ILIQUIN', 'CH+CHILIQUIN: highlight prefijo');
t(lib.acHighlight('San José', 'JOSE'), 'San <mark>José</mark>', 'JOSE+San José: highlight palabra con tilde');
t(lib.acHighlight('CARACAS', 'CA'), '<mark>CA</mark>RACAS', 'CA+CARACAS: highlight prefijo');
t(lib.acHighlight('Amazonas', 'AM'), '<mark>Am</mark>azonas', 'AM+Amazonas: highlight preserva case (Am)');

console.log('\n--- Test 3: Filtro startsWith + contains (1 char) ---');
const depts = ['AMAZONAS', 'ANCASH', 'APURIMAC', 'AREQUIPA', 'AYACUCHO', 'CAJAMARCA', 'CUSCO', 'LORETO'];
// Query=1 char: solo contains. Orden de items preservado.
// Items con "C" en lowercase (normSearch): ANCASH (no), AMAZONAS (no, "amazonas" sin c), APURIMAC (sí, "apurimac" tiene c), AREQUIPA (no), AYACUCHO (sí), CAJAMARCA (sí), CUSCO (sí), LORETO (no)
// Espera: APURIMAC, AYACUCHO, CAJAMARCA, CUSCO
tT(lib.acFilter(depts, 'C').includes('CAJAMARCA') && lib.acFilter(depts, 'C').includes('CUSCO'), 'C: contiene CAJAMARCA y CUSCO');
tT(!lib.acFilter(depts, 'C').includes('LORETO'), 'C: NO contiene LORETO (sin c)');

console.log('\n--- Test 4: Filtro startsWith (2+ chars) ---');
t(lib.acFilter(depts, 'CA'), ['CAJAMARCA', 'ANCASH'], 'CA: CAJAMARCA (startsWith) + ANCASH (contains)');
t(lib.acFilter(depts, 'AM'), ['AMAZONAS', 'CAJAMARCA', 'APURIMAC'], 'AM: AMAZONAS startsWith + CAJAMARCA/APURIMAC contains');
t(lib.acFilter(depts, 'LOR'), ['LORETO'], 'LOR: solo LORETO startsWith');

console.log('\n--- Test 5: Filtro fuzzy con typo (L-055) ---');
const ccpps = ['San José', 'San Juan', 'San Pedro', 'Santa María', 'San Luis', 'Huancayo', 'Cajamarca', 'Lima', 'Cusco'];
t(lib.acFilter(ccpps, 'JSE'), ['San José'], 'JSE (typo): matchea "San José" via subsequence (J,S,E en orden)');
t(lib.acFilter(ccpps, 'SNT'), ['Santa María'], 'SNT (typo): matchea "Santa" via subsequence (S,N,T en orden)');
t(lib.acFilter(ccpps, 'LUS'), ['San Luis'], 'LUS: matchea "San Luis" (L,U,S en orden en "san luis")');
t(lib.acFilter(ccpps, 'XX'), [], 'XX: no matchea nada (subsequence sin X)');

console.log('\n--- Test 6: Fuzzy NO matchea si chars no están en orden ---');
const test6 = ['XYZ', 'ABC', 'CBA'];
tT(!lib.acFilter(test6, 'BA').includes('XYZ'), 'BA no matchea XYZ');
tT(lib.acFilter(test6, 'BA').includes('CBA'), 'BA matchea CBA (B antes que A)');
tT(!lib.acFilter(test6, 'BA').includes('ABC'), 'BA no matchea ABC');

console.log('\n--- Test 7: Sin falsos positivos en fuzzy ---');
const test7 = ['AMAZONAS', 'MADRE DE DIOS', 'LORETO', 'UCAYALI'];
t(lib.acFilter(test7, 'MD'), ['MADRE DE DIOS'], 'MD: solo MADRE DE DIOS (M antes que D)');
t(lib.acFilter(test7, 'DM'), [], 'DM: nada (D antes que M en "MADRE", subsequence busca MD no DM)');

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
