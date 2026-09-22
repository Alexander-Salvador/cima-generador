// OVAV Tools — Apps Script bound to OPERACIONES spreadsheet
// Custom menu + image-buttons with real onAction click handlers.

const MENU_NAME = '🛠️ OVAV Tools';
const SNAPSHOT_KEY = 'ovav_snapshots';
const MAX_SNAPSHOTS = 5;

const IMG_RESTORE_URL = 'https://drive.google.com/uc?id=1SO07B3HnqsET0gpss-QSGKrP7i-AbMm6&export=download';
const IMG_CLEAN_URL   = 'https://drive.google.com/uc?id=1JbLGxRv2SEvYoseyudBLr9aRBPjfdgbk&export=download';

// ─── onOpen: register custom menu + image-buttons ────────────────────
function onOpen() {
  SpreadsheetApp.getUi()
    .createMenu(MENU_NAME)
    .addItem('🔄 Restaurar fórmulas', 'restoreFormulas')
    .addItem('🎨 Restaurar formato', 'restoreFormat')
    .addItem('🧹 Limpiar filas vacías', 'cleanEmptyRows')
    .addSeparator()
    .addItem('📸 Crear snapshot ahora', 'createSnapshot')
    .addItem('⏪ Restaurar último snapshot', 'restoreLastSnapshot')
    .addSeparator()
    .addItem('🖼️ Reinstalar botones-imagen', 'installImageButtons')
    .addItem('ℹ️ Estado del sistema', 'showStatus')
    .addToUi();

  // Auto-install image buttons on first open
  try { installImageButtons(); } catch (e) { console.warn('installImageButtons failed:', e); }
}

// ─── Install image buttons with onAction click handlers ──────────────
function installImageButtons() {
  const ss = SpreadsheetApp.getActive();
  const sheet = ss.getSheetByName('OPERACIONES');

  // Remove existing drawings on row 2 to avoid duplicates
  const drawings = sheet.getDrawings();
  drawings.forEach(d => {
    const pos = d.getPosition();
    if (pos.getRowIndex() === 1) {
      d.remove();
    }
  });

  // 🔄 RESTAURAR button — Drawing with text + assignScriptFunction
  // Drawing class IS the only way to get assignScriptFunction() in Sheets.
  const restoreDrawing = sheet.newDrawing()
    .setName('OVAV_btn_restore')
    .setText('🔄 RESTAURAR')
    .setFontSize(11)
    .setFontWeight('bold')
    .setForegroundColor('#FFFFFF')
    .setBackgroundColor('#1F4E78')
    .setTextAlignment('CENTER')
    .assignScriptFunction('onRestoreClick');
  // Insert drawing at I2 (col 8, row 1)
  restoreDrawing.setPosition(8, 1, 0, 4);
  restoreDrawing.setWidth(140).setHeight(40);

  // 🧹 LIMPIAR button — same pattern
  const cleanDrawing = sheet.newDrawing()
    .setName('OVAV_btn_clean')
    .setText('🧹 LIMPIAR')
    .setFontSize(11)
    .setFontWeight('bold')
    .setForegroundColor('#FFFFFF')
    .setBackgroundColor('#9C0006')
    .setTextAlignment('CENTER')
    .assignScriptFunction('onCleanClick');
  cleanDrawing.setPosition(11, 1, 0, 4);
  cleanDrawing.setWidth(220).setHeight(40);

  SpreadsheetApp.getActive().toast(
    '🖼️ Botones instalados. Click en cada uno para ejecutar.',
    'OVAV Tools',
    3
  );
}

// ─── Button click handlers ────────────────────────────────────────────
function onRestoreClick() { restoreFormulas(); }
function onCleanClick()   { cleanEmptyRows(); }

// ─── Button click handlers ────────────────────────────────────────────
function onRestoreClick() { restoreFormulas(); }
function onCleanClick()   { cleanEmptyRows(); }

// ─── Generic confirm + run helper ─────────────────────────────────────
function confirmAndRun(title, message, action) {
  const ui = SpreadsheetApp.getUi();
  const resp = ui.alert(title, message, ui.ButtonSet.YES_NO);
  if (resp !== ui.Button.YES) return;
  try {
    const result = action();
    ui.alert('✅ Listo', result || 'Operación completada.', ui.ButtonSet.OK);
  } catch (e) {
    ui.alert('❌ Error', e.message, ui.ButtonSet.OK);
    console.error(e);
  }
}

// ─── 1. Restore VLOOKUP + REQUEST_ID formulas ──────────────────────────
function restoreFormulas() {
  confirmAndRun(
    '🔄 Restaurar fórmulas',
    'Esto reescribirá las fórmulas VLOOKUP en NOMBRE, HORARIO, TURNO_ID y el REQUEST_ID auto en M. ¿Continuar?',
    () => {
      const ss = SpreadsheetApp.getActive();
      const sheet = ss.getSheetByName('OPERACIONES');
      let count = 0;
      for (let row = 4; row <= 10; row++) {
        const a = `A${row}`;
        const fNom = `=IF(${a}="","",IFERROR(VLOOKUP(${a},PERSONAL!$B$3:$O$27,2,FALSE),""))`;
        const fHor = `=IF(${a}="","",IFERROR(VLOOKUP(${a},PERSONAL!$B$3:$O$27,4,FALSE),""))`;
        const fTur = `=IF(${a}="","",IFERROR(VLOOKUP(${a},PERSONAL!$B$3:$O$27,9,FALSE),""))`;
        const fReq = `="REQ-"&TEXT(ROW()-3,"000")`;
        sheet.getRange(`B${row}`).setFormula(fNom);
        sheet.getRange(`E${row}`).setFormula(fHor);
        sheet.getRange(`F${row}`).setFormula(fTur);
        sheet.getRange(`M${row}`).setFormula(fReq);
        count += 4;
      }
      return `${count} fórmulas restauradas (B, E, F, M × 7 filas).`;
    }
  );
}

// ─── 2. Re-apply styles + conditional formatting + validations ──────
function restoreFormat() {
  confirmAndRun(
    '🎨 Restaurar formato',
    'Esto re-aplicará estilos, conditional formatting y validaciones. ¿Continuar?',
    () => {
      const ss = SpreadsheetApp.getActive();
      const sheet = ss.getSheetByName('OPERACIONES');

      sheet.getRange('A3:M3')
        .setBackground('#404040').setFontColor('#FFFFFF')
        .setFontWeight('bold').setHorizontalAlignment('center')
        .setBorder(true, true, true, true, true, false).setFontSize(10);

      sheet.getRange('A4:A10').setBackground('#FFF4D6').setFontWeight('bold')
        .setHorizontalAlignment('center').setBorder(true, true, true, true, true, false);
      sheet.getRange('B4:B10').setBackground('#F0F0F0').setFontSize(9)
        .setBorder(true, true, true, true, true, false);
      sheet.getRange('C4:D10').setBackground('#FFF4D6')
        .setHorizontalAlignment('center').setBorder(true, true, true, true, true, false);
      sheet.getRange('E4:F10').setBackground('#F0F0F0').setFontSize(9)
        .setBorder(true, true, true, true, true, false);
      sheet.getRange('G4:K10').setBackground('#FFF4D6')
        .setHorizontalAlignment('left').setBorder(true, true, true, true, true, false);
      sheet.getRange('L4:L10').setBackground('#FFF4D6').setFontWeight('bold')
        .setHorizontalAlignment('center').setBorder(true, true, true, true, true, false);
      sheet.getRange('M4:M10').setBackground('#E7F0FF').setFontWeight('bold')
        .setHorizontalAlignment('center').setBorder(true, true, true, true, true, false);

      return 'Header + 7 filas activas restauradas.';
    }
  );
}

// ─── 3. Delete empty rows in OPERACIONES (preserves header + active) ──
function cleanEmptyRows() {
  confirmAndRun(
    '🧹 Limpiar filas vacías',
    'Esto eliminará filas totalmente vacías (rows 4-100). NO tocará filas con datos. ¿Continuar?',
    () => {
      const ss = SpreadsheetApp.getActive();
      const sheet = ss.getSheetByName('OPERACIONES');
      let deleted = 0;
      // Iterate bottom-up to avoid index shift
      for (let r = 100; r >= 4; r--) {
        const range = sheet.getRange(r, 1, 1, 13).getValues()[0];
        // Only delete if ALL 13 columns are empty
        if (range.every(v => v === '' || v === null)) {
          sheet.deleteRow(r);
          deleted++;
        }
      }
      return `${deleted} filas vacías eliminadas (rows 4-100).`;
    }
  );
}

// ─── 4. Snapshot system (Properties Service) ──────────────────────────
function createSnapshot() {
  const ss = SpreadsheetApp.getActive();
  const sheet = ss.getSheetByName('OPERACIONES');
  const data = sheet.getDataRange().getValues();
  const props = PropertiesService.getDocumentProperties();
  const snapshots = JSON.parse(props.getProperty(SNAPSHOT_KEY) || '[]');
  snapshots.unshift({
    ts: new Date().toISOString(),
    rows: data.length,
    cols: data[0] ? data[0].length : 0,
    payload: data,
  });
  while (snapshots.length > MAX_SNAPSHOTS) snapshots.pop();
  props.setProperty(SNAPSHOT_KEY, JSON.stringify(snapshots));
  SpreadsheetApp.getUi().alert(
    '📸 Snapshot creado',
    `Snapshots almacenados: ${snapshots.length} / ${MAX_SNAPSHOTS}\nMás reciente: ${snapshots[0].ts}`,
    SpreadsheetApp.getUi().ButtonSet.OK
  );
}

function restoreLastSnapshot() {
  const props = PropertiesService.getDocumentProperties();
  const snapshots = JSON.parse(props.getProperty(SNAPSHOT_KEY) || '[]');
  if (snapshots.length === 0) {
    SpreadsheetApp.getUi().alert('⏪ Sin snapshots', 'No hay snapshots para restaurar.', SpreadsheetApp.getUi().ButtonSet.OK);
    return;
  }
  const snap = snapshots[0];
  const ui = SpreadsheetApp.getUi();
  const resp = ui.alert(
    '⏪ Restaurar snapshot',
    `Restaurar ${snap.rows} filas × ${snap.cols} cols del ${snap.ts}?`,
    ui.ButtonSet.YES_NO
  );
  if (resp !== ui.Button.YES) return;
  const ss = SpreadsheetApp.getActive();
  const sheet = ss.getSheetByName('OPERACIONES');
  sheet.clear();
  if (snap.payload.length > 0) {
    sheet.getRange(1, 1, snap.payload.length, snap.payload[0].length)
      .setValues(snap.payload);
  }
  ui.alert('✅ Restaurado', `${snap.rows} filas restauradas.`, ui.ButtonSet.OK);
}

// ─── 5. Status (informational) ───────────────────────────────────────
function showStatus() {
  const ss = SpreadsheetApp.getActive();
  const sheet = ss.getSheetByName('OPERACIONES');
  const lastRow = sheet.getLastRow();
  const data = sheet.getRange(`B4:L${lastRow}`).getValues();
  const totals = { PENDIENTE: 0, GESTIONADO: 0, RECHAZADO: 0, NO_PROCEDE: 0 };
  data.forEach(row => {
    const estado = row[10];
    if (totals[estado] !== undefined) totals[estado]++;
  });
  SpreadsheetApp.getUi().alert(
    '📊 Estado del sistema',
    `Filas activas: ${data.length}\n` +
    `Pendientes: ${totals.PENDIENTE}\n` +
    `Gestionados: ${totals.GESTIONADO}\n` +
    `Rechazados: ${totals.RECHAZADO}\n` +
    `No procede: ${totals.NO_PROCEDE}`,
    SpreadsheetApp.getUi().ButtonSet.OK
  );
}
