/**
 * 01_CfgLoader.gs
 * ----------------------------------------------------------------------------
 * Lee 00_Config!B2:B9 y expone los valores vía CacheService.
 * Reduce lecturas a Sheets API en operaciones masivas (load/save/sync).
 *
 * Anchors (architecture.yaml v2):
 *   B2 = mes activo (1-12)
 *   B3 = año activo (YYYY)
 *   B4 = semana activa (W01..W53)
 *   B5 = persona filtro ("TODOS" o "P-NNN")
 *   B6 = factor HE diurno (default 1.25)
 *   B7 = factor HE nocturno (default 1.35)
 *   B8 = factor feriado (default 2.00)
 *   B9 = jornada max diaria (default 8)
 *   B11 = timestamp último refresh
 *
 * Cache TTL: 600s (10 min) · prefix: "he_lima_"
 * Authority: andres · area: platform_engineering · lead: thavren
 */

/** @constant {string} Nombre de la hoja de configuración */
const SHEET_CFG = '00_Config';

/** @constant {string} Prefijo de claves para CacheService */
const CACHE_PREFIX = 'he_lima_';

/** @constant {number} TTL del caché en segundos */
const CACHE_TTL_SEC = 600;

/** @constant {Object<string, string>} Mapa anchor → celda */
const CFG_ANCHORS = Object.freeze({
  mes: 'B2',
  anio: 'B3',
  semana: 'B4',
  persona: 'B5',
  factorDiurno: 'B6',
  factorNocturno: 'B7',
  factorFeriado: 'B8',
  jornadaMax: 'B9',
  lastRefresh: 'B11',
});

/**
 * Cache interna local para evitar llamadas repetidas a CacheService dentro
 * del mismo execution context (Apps Script no comparte caché entre runs).
 * @type {Object<string, *>}
 */
const _memCache = {};

/**
 * Obtiene la hoja 00_Config (helper reutilizable).
 * @returns {GoogleAppsScript.Spreadsheet.Sheet}
 * @throws {Error} Si la hoja no existe.
 */
function getCfgSheet_() {
  const ss = SpreadsheetApp.getActiveSpreadsheet();
  const sh = ss.getSheetByName(SHEET_CFG);
  if (!sh) {
    throw new Error('cfg_missing: hoja 00_Config no encontrada en el spreadsheet');
  }
  return sh;
}

/**
 * Lee el valor de un anchor desde 00_Config con caché.
 * Cache miss → lee Sheet + escribe CacheService.
 * @param {string} key Clave lógica ('mes', 'anio', 'factorDiurno', etc.)
 * @returns {*} Valor leído (tipo depende del anchor)
 */
function readCfg_(key) {
  const cell = CFG_ANCHORS[key];
  if (!cell) {
    throw new Error('cfg_unknown_key: ' + key);
  }
  const cacheKey = CACHE_PREFIX + 'cfg_' + key;

  // 1. Mem cache (durante mismo execution)
  if (Object.prototype.hasOwnProperty.call(_memCache, cacheKey)) {
    return _memCache[cacheKey];
  }

  // 2. CacheService (entre executions)
  const cache = CacheService.getDocumentCache();
  const cached = cache.get(cacheKey);
  if (cached !== null) {
    const parsed = JSON.parse(cached);
    _memCache[cacheKey] = parsed;
    return parsed;
  }

  // 3. Sheets API
  const sh = getCfgSheet_();
  const raw = sh.getRange(cell).getValue();

  // Normalización de tipos por anchor
  let normalized = raw;
  if (['mes', 'anio', 'jornadaMax'].indexOf(key) >= 0) {
    normalized = Number(raw);
  } else if (key === 'factorDiurno' || key === 'factorNocturno' || key === 'factorFeriado') {
    normalized = Number(raw);
  } else if (key === 'semana' && typeof raw === 'string') {
    normalized = raw; // "W##"
  }

  cache.put(cacheKey, JSON.stringify(normalized), CACHE_TTL_SEC);
  _memCache[cacheKey] = normalized;
  return normalized;
}

/**
 * Mes activo (1-12).
 * @returns {number}
 * @example const m = getMes(); // 3
 */
function getMes() {
  return readCfg_('mes');
}

/**
 * Año activo (YYYY).
 * @returns {number}
 * @example const y = getAnio(); // 2026
 */
function getAnio() {
  return readCfg_('anio');
}

/**
 * Semana activa en formato "W##".
 * @returns {string}
 * @example const w = getSemana(); // "W14"
 */
function getSemana() {
  return readCfg_('semana');
}

/**
 * Persona filtro ("TODOS" o "P-NNN").
 * @returns {string}
 */
function getPersonaFiltro() {
  return readCfg_('persona');
}

/**
 * Factores HE + jornada max en un solo bundle (optimiza 1 round-trip a caché).
 * @returns {{diurno:number, nocturno:number, feriado:number, jornadaMax:number}}
 * @example
 *   const f = getFactores();
 *   // {diurno:1.25, nocturno:1.35, feriado:2.00, jornadaMax:8}
 */
function getFactores() {
  return {
    diurno: readCfg_('factorDiurno'),
    nocturno: readCfg_('factorNocturno'),
    feriado: readCfg_('factorFeriado'),
    jornadaMax: readCfg_('jornadaMax'),
  };
}

/**
 * Invalida toda la caché de configuración (forzar re-lectura desde Sheets).
 * Llamar cuando un administrador edita manualmente 00_Config!B2:B9.
 * @returns {void}
 */
function invalidateCache() {
  const cache = CacheService.getDocumentCache();
  // Borrar todas las claves conocidas
  Object.keys(CFG_ANCHORS).forEach(function (key) {
    cache.remove(CACHE_PREFIX + 'cfg_' + key);
    delete _memCache[CACHE_PREFIX + 'cfg_' + key];
  });
  // También limpia caché de feriados (lo mantiene 02_HolidaysEngine)
  cache.remove(CACHE_PREFIX + 'feriados_anio');
}

/**
 * Escribe el timestamp de último refresh en 00_Config!B11.
 * Llamado por 05_DashboardRefresh y 08_DBSync después de operaciones pesadas.
 * @returns {void}
 */
function touchLastRefresh() {
  const sh = getCfgSheet_();
  sh.getRange(CFG_ANCHORS.lastRefresh).setValue(new Date());
  // B11 no se cachea (es efímero)
}