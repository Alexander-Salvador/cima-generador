const fs = require('fs');
const html = fs.readFileSync('/home/braka/Systems/projects/work/cima/apps/cima-generador/index.html', 'utf8');

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
const mockInputs = {};
let writeCounts = {};
function $(id) {
  if (!mockInputs[id]) {
    mockInputs[id] = { _value: '', setAttribute: () => {} };
    Object.defineProperty(mockInputs[id], 'value', {
      get() { return this._value; },
      set(v) { 
        if (v !== this._value) { 
          this._value = v; 
          writeCounts[id] = (writeCounts[id] || 0) + 1; 
        }
      }
    });
  }
  return mockInputs[id];
}
function getWrites(id) { return writeCounts[id] || 0; }
function resetWrites(id) { delete writeCounts[id]; }
${extractFunc('upper')}
${extractFunc('normSearch')}
${extractFunc('acFilter')}
${extractFunc('acHighlight')}
${extractFunc('acEscape')}

$('departamento'); $('provincia'); $('distrito'); $('ccpp'); $('typing');

globalThis.__TEST__ = { upper, normSearch, acFilter, acHighlight, acEscape, $, getWrites, resetWrites };
`;
eval(code);
const lib = globalThis.__TEST__;

let pass = 0, fail = 0;
function t(actual, expected, msg) {
  const ok = JSON.stringify(actual) === JSON.stringify(expected);
  if (ok) { pass++; console.log(`  ✓ ${msg}`); }
  else { fail++; console.log(`  ✗ ${msg}: actual=${JSON.stringify(actual)} expected=${JSON.stringify(expected)}`); }
}

console.log('=== upper() ===');
lib.$('departamento')._value = 'Lima';
t(lib.upper('departamento'), 'LIMA', 'upper("Lima") → "LIMA"');
t(lib.getWrites('departamento'), 1, '  1 escritura (de Lima a LIMA)');

lib.resetWrites('departamento');
lib.upper('departamento');
t(lib.getWrites('departamento'), 0, 'upper("LIMA") no muta si ya es uppercase (CRIT: no resetea cursor)');

lib.$('distrito')._value = 'San Miguel';
t(lib.upper('distrito'), 'SAN MIGUEL', 'upper("San Miguel") → "SAN MIGUEL"');

console.log('\n=== normSearch() accent-insensitive (sin trim) ===');
t(lib.normSearch('Lima'), 'lima', 'normal');
t(lib.normSearch('Líme'), 'lime', 'sin tildes');
t(lib.normSearch('San Miguel'), 'san miguel', 'lowercase + space');
t(lib.normSearch('MÓVIL'), 'movil', 'uppercase + tildes');
t(lib.normSearch('Ñaña'), 'nana', 'ñ→n (sin trim en espacios)');
t(lib.normSearch('  espacios  '), '  espacios  ', 'preserva espacios (NO trimea, es semántico)');

console.log('\n=== acFilter() accent-insensitive ===');
const opts = ['LIMA', 'Lima', 'LIME', 'LIMÓN', 'LORETO', 'LAMBAYEQUE'];
t(lib.acFilter(opts, 'LIMA'), ['LIMA', 'Lima'], 'startsWith LIMA: LIMA, Lima');
t(lib.acFilter(opts, 'LIM').length, 4, 'LIM contains: LIMA, Lima, LIME, LIMÓN');
t(lib.acFilter(opts, 'lima').length, 2, 'lima: LIMA, Lima');

const ccpps = ['Bellavista', 'San José', 'María', 'Pueblo Nuevo'];
t(lib.acFilter(ccpps, 'jose'), ['San José'], 'jose → San José (sin tilde)');
t(lib.acFilter(ccpps, 'maria'), ['María'], 'maria → María (sin tilde)');
t(lib.acFilter(ccpps, 'pueblo'), ['Pueblo Nuevo'], 'pueblo → Pueblo Nuevo');
t(lib.acFilter(ccpps, 'SAN JOSE'), ['San José'], 'SAN JOSE → San José');
t(lib.acFilter(ccpps, '  maria  '), ['María'], 'query con whitespace se trimea → matchea María');

console.log('\n=== acHighlight() preserva case original ===');
t(lib.acHighlight('LIMA', 'LIM'), '<mark>LIM</mark>A', 'LIMA|LIM → "<mark>LIM</mark>A"');
t(lib.acHighlight('San José', 'jose'), 'San <mark>José</mark>', 'preserva tilde (José)');
t(lib.acHighlight('LIME', 'LIME'), '<mark>LIME</mark>', 'match exacto LIME');
t(lib.acHighlight('San Miguel', 'MIG'), 'San <mark>Mig</mark>uel', 'preserva case original (Mig, no MIG)');
t(lib.acHighlight('María', 'MAR'), '<mark>Mar</mark>ía', 'preserva lowercase Mar (no María)');
t(lib.acHighlight('Ñaña', 'NAN'), '<mark>Ñañ</mark>a', 'preserva tilde (Ñañ)');
t(lib.acHighlight('LIMA', 'LIMA'), '<mark>LIMA</mark>', 'match full LIMA');
t(lib.acHighlight('LIMA', ''), 'LIMA', 'query vacío retorna plano');

console.log('\n=== NO mutación DOM durante typing ===');
lib.resetWrites('departamento');
lib.upper('departamento');
t(lib.getWrites('departamento'), 0, 'upper repetido no muta DOM (UX cursor preservado)');

console.log(`\n${'='.repeat(50)}`);
console.log(`${pass} PASS / ${fail} FAIL`);
process.exit(fail > 0 ? 1 : 0);
