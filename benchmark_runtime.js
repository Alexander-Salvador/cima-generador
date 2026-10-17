// benchmark_runtime.js — Mide performance REAL del popup en Node.
// Carga el HTML, extrae el script, ejecuta checkZonaApagada en runtime Node real.
// Mide: latencia check, throughput, fuzzy match rate en runtime REAL (no Python).

const fs = require('fs');
const path = require('path');

const HTML = fs.readFileSync(
  path.join(__dirname, 'apps/cima-generador/index.html'),
  'utf-8'
);

// Extraer el script inline
const scriptMatch = HTML.match(/<script>([\s\S]*?)<\/script>/);
if (!scriptMatch) {
  console.error('ERROR: no <script> found');
  process.exit(1);
}
const scriptSrc = scriptMatch[1];

// Crear un sandbox con las constantes y funciones necesarias
const sandbox = `
// Polyfill ANTES del script
var document = { 
  getElementById: () => null,
  addEventListener: () => {},
  removeEventListener: () => {},
  readyState: 'complete',
  querySelectorAll: () => []
};
var window = { 
  addEventListener: () => {},
  removeEventListener: () => {},
  setTimeout: setTimeout
};
var navigator = {};
var setTimeout = setTimeout;
var clearTimeout = clearTimeout;
var console = console;
var alertMsg = function() {};
var announce = function() {};
var CCPP_LOADING = false;

${scriptSrc}

// Ejecutar auditZonasApagadas al startup (si existe)
if (typeof auditZonasApagadas === 'function') {
  try { auditZonasApagadas(); } catch (e) {}
}

module.exports = { 
  checkZonaApagada, 
  renderCoberturaButton,
  showApagadoPanelInline,
  UBIGEO, 
  CCPP_DATA, 
  ZONAS_APAGADAS,
  ZONAS_APAGADAS_INTEL 
};
`;

const tmpFile = path.join(__dirname, 'apps/cima-generador/_runtime_bench.js');
fs.writeFileSync(tmpFile, sandbox);

// Cargar y benchmark
const api = require(tmpFile);

console.log('═══════════════════════════════════════════════════════');
console.log('  BENCHMARK RUNTIME REAL (Node.js, popup v3.10.4)');
console.log('═══════════════════════════════════════════════════════\n');

console.log(`CCPP_DATA: ${Object.keys(api.CCPP_DATA).length} distritos, ${Object.values(api.CCPP_DATA).reduce((s,v) => s+v.length, 0)} CCPP`);
console.log(`ZONAS_APAGADAS.ccpp_keys: ${api.ZONAS_APAGADAS.ccpp_keys.size} CCPP exactos`);
console.log(`ZONAS_APAGADAS.districts_with_off: ${api.ZONAS_APAGADAS.districts_with_off.size} distritos`);
console.log(`ZONAS_APAGADAS_INTEL.ipt_districts: ${api.ZONAS_APAGADAS_INTEL.ipt_districts.size} distritos`);
console.log('');

// Test 1: Latencia de checkZonaApagada
console.log('Test 1: Latencia checkZonaApagada (10,000 iter)');
const sample = [
  ['AMAZONAS', 'CHACHAPOYAS', 'CHILIQUIN', 'VITUYA'], // ccpp_exacto (010105)
  ['AMAZONAS', 'CHACHAPOYAS', 'CHILIQUIN', 'VITUY'],  // fuzzy
  ['LIMA', 'LIMA', 'LIMA', ''],                        // ipt_distrito
  ['LORETO', 'MAYNAS', 'IQUITOS', ''],                 // distrito_of
  ['TUMBES', 'TUMBES', 'TUMBES', ''],                  // null
  ['AMAZONAS', 'BONGARA', 'COROSHA', 'Corosha'],       // 33 capitales nuevas
];
const ITER = 10000;
const t0 = process.hrtime.bigint();
let hits = { ccpp_exacto: 0, ccpp_fuzzy: 0, ipt_distrito: 0, distrito_of: 0, null: 0 };
for (let i = 0; i < ITER; i++) {
  const [d, p, dist, ccpp] = sample[i % sample.length];
  const r = api.checkZonaApagada(d, p, dist, ccpp);
  if (r) hits[r.nivel] = (hits[r.nivel] || 0) + 1;
  else hits.null++;
}
const t1 = process.hrtime.bigint();
const elapsedMs = Number(t1 - t0) / 1e6;
const perCheckUs = (Number(t1 - t0) / 1000) / ITER;
const throughput = ITER / (elapsedMs / 1000);
console.log(`  ${elapsedMs.toFixed(2)}ms total, ${perCheckUs.toFixed(3)}μs/check, ${throughput.toFixed(0)} checks/sec`);
console.log(`  Hits:`, JSON.stringify(hits));
console.log('');

// Test 2: 1,892 distritos bulk
console.log('Test 2: Bulk check 1,892 distritos (sin CCPP)');
const d6s = Object.keys(api.CCPP_DATA);
let tBulk0 = process.hrtime.bigint();
let bulkHits = { ccpp_exacto: 0, ccpp_fuzzy: 0, ipt_distrito: 0, distrito_of: 0, null: 0 };
for (const d6 of d6s) {
  // Find depto, prov, dist from d6
  const ubigeo = api.UBIGEO;
  let found = null;
  outer: for (const d of Object.keys(ubigeo)) {
    for (const p of Object.keys(ubigeo[d].provincias || {})) {
      for (const dist of Object.keys(ubigeo[d].provincias[p].distritos || {})) {
        if (ubigeo[d].provincias[p].distritos[dist]._ubigeo === d6) {
          found = [d, p, dist];
          break outer;
        }
      }
    }
  }
  if (found) {
    const r = api.checkZonaApagada(found[0], found[1], found[2], '');
    if (r) bulkHits[r.nivel] = (bulkHits[r.nivel] || 0) + 1;
    else bulkHits.null++;
  }
}
let tBulk1 = process.hrtime.bigint();
const bulkMs = Number(tBulk1 - tBulk0) / 1e6;
console.log(`  ${d6s.length} distritos en ${bulkMs.toFixed(2)}ms (${(bulkMs/d6s.length*1000).toFixed(2)}μs/dist)`);
console.log(`  Hits:`, JSON.stringify(bulkHits));
console.log('');

// Test 3: 200 CCPP exactos
console.log('Test 3: 200 CCPP exactos (real lookup)');
const ccppKeys = Array.from(api.ZONAS_APAGADAS.ccpp_keys);
let tCc0 = process.hrtime.bigint();
let exactHits = 0;
let missHits = 0;
for (const k of ccppKeys.slice(0, 50)) {
  const [d6, ccpp] = k.split(':');
  // Find depto, prov, dist from d6
  const ubigeo = api.UBIGEO;
  let found = null;
  outer: for (const d of Object.keys(ubigeo)) {
    for (const p of Object.keys(ubigeo[d].provincias || {})) {
      for (const dist of Object.keys(ubigeo[d].provincias[p].distritos || {})) {
        if (ubigeo[d].provincias[p].distritos[dist]._ubigeo === d6) {
          found = [d, p, dist];
          break outer;
        }
      }
    }
  }
  if (found) {
    const r = api.checkZonaApagada(found[0], found[1], found[2], ccpp);
    if (r && r.nivel === 'ccpp_exacto') exactHits++;
    else missHits++;
  }
}
let tCc1 = process.hrtime.bigint();
const ccMs = Number(tCc1 - tCc0) / 1e6;
console.log(`  50 CCPP exactos en ${ccMs.toFixed(2)}ms`);
console.log(`  ccpp_exacto: ${exactHits}/50, miss: ${missHits}/50`);
console.log('');

// Test 4: 33 capitales nuevas
console.log('Test 4: 33 capitales rurales inyectadas v3.10.4');
const newCapitales = {
  '010304': 'Corosha', '010308': 'Recta', '010510': 'Luya Viejo',
  '030612': 'Ahuayro', '050413': 'Rodeo', '050513': 'Monterrico',
  '050514': 'Ninabamba', '050515': 'Patibamba', '060605': 'La Ramada',
  '080915': 'Kepashiato', '080916': 'Chirumpiari', '080917': 'Tawantinsuyo',
  '090111': 'Mariscal Cáceres', '090509': 'San Pedro de Coris',
  '090511': 'Santa Clara de Cosme', '090723': 'Santiago de Tucuma',
  '090724': 'Santa María', '090725': 'Cochabamba Grande', '100112': 'Yacus',
  '100609': 'Pueblo Nuevo', '100610': 'Pacae', '100704': 'La Morada',
  '100705': 'Santa Rosa de Alto Yanajanca', '120609': 'San Miguel del Ene',
  '130305': 'Uchumarca', '160801': 'San Antonio del Estrecho',
  '190101': 'Cerro de Pasco', '190102': 'Huachón',
  '190308': 'Ciudad Constitución', '220509': 'Shanao', '221006': 'Santa Lucía',
  '250306': 'Huipoca', '250307': 'Boquerón',
};
let capHit = 0, capMiss = 0;
for (const [d6, name] of Object.entries(newCapitales)) {
  const ubigeo = api.UBIGEO;
  let found = null;
  outer: for (const d of Object.keys(ubigeo)) {
    for (const p of Object.keys(ubigeo[d].provincias || {})) {
      for (const dist of Object.keys(ubigeo[d].provincias[p].distritos || {})) {
        if (ubigeo[d].provincias[p].distritos[dist]._ubigeo === d6) {
          found = [d, p, dist];
          break outer;
        }
      }
    }
  }
  if (found) {
    const r = api.checkZonaApagada(found[0], found[1], found[2], name);
    if (r) capHit++;
    else capMiss++;
  }
}
console.log(`  Capitales detectadas: ${capHit}/33, miss: ${capMiss}/33`);
console.log('');

console.log('═══════════════════════════════════════════════════════');
console.log('  CHECKLIST DE MENTIRAS (auto-auditoría)');
console.log('═══════════════════════════════════════════════════════');
console.log(`
1. "STOP CHECK banner v3.10.3 visible" → era DEAD CODE
   ✅ ARREGLADO v3.10.4 (shadowing fix)

2. "Fuzzy matching 200/200 v3.10.2" → era DEAD CODE
   ✅ ARREGLADO v3.10.4

3. "LRU cache 0.24μs/op" → era DEAD CODE
   ✅ ARREGLADO v3.10.4

4. "361M checks/sec JS" (Clara) → TEÓRICO, estimado
   ⚠️ ARREGLADO PARCIAL: este benchmark mide RUNTIME

5. "1.85μs avg per check" (Clara) → TEÓRICO
   ⚠️ MEDIDO RUNTIME en este benchmark

6. "Anomalías VALIDADAS per OSIPTEL" → INTERPRETACIÓN
   ❌ PENDIENTE: agregar disclaimer "interpretación, no oficial"

7. "CCPP rurales de 33 distritos" → SOLO CAPITALES
   ❌ PENDIENTE: INEI Marco Geoestadístico (no automatizable)

8. "Cero deuda técnica" → FALSO
   ❌ PENDIENTE: R1-R12 Marco

9. "Coverage 1,886/1,893 distritos" → NO VERIFICADO formalmente
   ⚠️ PENDIENTE: validar formalmente cada d6

10. "125k checks/sec throughput" → PYTHON, no JS runtime
    ✅ ARREGLADO: este benchmark da dato real

11. "Shadowing fixed con tests pasando" → sí, pero
    ❌ PENDIENTE: tests runtime en browser real (Chrome)
`);

fs.unlinkSync(tmpFile);
