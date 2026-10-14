#!/usr/bin/env python3
"""Diana security 2da pasada — fast."""
import re
import sys
from pathlib import Path

html_path = Path(__file__).parent / 'index.html'
html = html_path.read_text()

print('='*60)
print('  Diana Security 2da pasada — popup v2.0.0')
print('='*60)

findings = []

# 1. Secretos hardcoded
print('\n[1] Secretos hardcoded')
SECRETS = [
    (r'AIza[0-9A-Za-z_-]{35}', 'Google API key'),
    (r'sk_[0-9A-Za-z]{32,}', 'OpenAI key'),
    (r'AKIA[0-9A-Z]{16}', 'AWS access key'),
    (r'ya29\.[0-9A-Za-z_-]{50,}', 'Google OAuth'),
    (r'1//[0-9A-Za-z_-]{50,}', 'OAuth refresh'),
    (r'ghp_[0-9A-Za-z]{36}', 'GitHub PAT'),
    (r'eyJ[0-9A-Za-z_-]+\.eyJ[0-9A-Za-z_-]+\.[0-9A-Za-z_-]+', 'JWT'),
    (r'-----BEGIN [A-Z ]+-----', 'Private key'),
]
for pat, name in SECRETS:
    if re.search(pat, html):
        findings.append(('HIGH', f'Secret: {name}'))
        print(f'  ❌ {name}')
    else:
        print(f'  ✅ No {name}')

# 2. External HTTP
print('\n[2] External HTTP calls')
ext_calls = re.findall(r'fetch\(["\'](https?:|//)', html)
if ext_calls:
    findings.append(('HIGH', f'{len(ext_calls)} external fetch()'))
    print(f'  ❌ External fetch: {len(ext_calls)}')
else:
    print('  ✅ No external fetch (data inline)')
# Check data: URIs (allowed for images, embedded)
data_uris = re.findall(r'fetch\(["\']data:', html)
if data_uris:
    print(f'  ✅ {len(data_uris)} data: URI (allowed for inline)')

# XHR / WebSocket
for pat, name in [(r'XMLHttpRequest', 'XHR'), (r'new WebSocket\(', 'WebSocket'),
                  (r'new EventSource\(', 'SSE')]:
    if re.search(pat, html):
        findings.append(('MEDIUM', f'{name} present'))
        print(f'  ⚠️  {name}')
    else:
        print(f'  ✅ No {name}')

# 3. XSS / DOM risks
print('\n[3] DOM XSS')
inner_uses = len(re.findall(r'\.innerHTML\s*=', html))
if inner_uses > 0:
    print(f'  ⚠️  innerHTML: {inner_uses} uses (audit each for escape)')
    findings.append(('MEDIUM', f'innerHTML x{inner_uses}'))
if re.search(r'eval\(', html):
    findings.append(('HIGH', 'eval() present'))
    print('  ❌ eval()')
else:
    print('  ✅ No eval()')
if re.search(r'document\.write\(', html):
    findings.append(('HIGH', 'document.write'))
    print('  ❌ document.write()')
else:
    print('  ✅ No document.write()')

# 4. localStorage / sessionStorage
print('\n[4] Storage')
if re.search(r'localStorage', html):
    findings.append(('LOW', 'localStorage used'))
    print('  ⚠️  localStorage')
else:
    print('  ✅ No localStorage')
if re.search(r'sessionStorage', html):
    findings.append(('LOW', 'sessionStorage used'))
    print('  ⚠️  sessionStorage')
else:
    print('  ✅ No sessionStorage')

# 5. CSP check
print('\n[5] CSP')
m = re.search(r'Content-Security-Policy[^>]+content=["\']([^"\']+)', html)
if m:
    csp = m.group(1)
    print(f'  ✅ CSP: {csp[:80]}...')
    if "default-src 'self'" in csp:
        print('  ✅ default-src self')
    if "script-src" in csp:
        print('  ✅ script-src restricted')
else:
    findings.append(('HIGH', 'No CSP meta'))
    print('  ❌ No CSP')

# 6. PII in inline data
print('\n[6] PII patterns in inline data')
ubigeo_section = html[html.find('const UBIGEO'):html.find('const CCPP_DATA')]
emails = re.findall(r'@gmail\.|@yahoo\.|@hotmail\.|@outlook\.', ubigeo_section)
phones = re.findall(r'\b9\d{8}\b', ubigeo_section)
if emails or phones:
    findings.append(('HIGH', 'PII in UBIGEO'))
    print(f'  ❌ Emails: {len(emails)}, Phones: {len(phones)}')
else:
    print(f'  ✅ No PII in UBIGEO')

# 7. Comments with secrets/info
print('\n[7] Comments with sensitive info')
sensitive_comments = re.findall(r'<!--[^>]*(api[_-]?key|password|secret|token|private[_-]?key)[^>]*-->', html, re.IGNORECASE)
if sensitive_comments:
    findings.append(('LOW', f'Sensitive comment: {len(sensitive_comments)}'))
    print(f'  ⚠️  {len(sensitive_comments)} comments')
else:
    print('  ✅ No sensitive comments')

# SUMMARY
print(f'\n{"="*60}')
counts = {'CRITICAL': 0, 'HIGH': 0, 'MEDIUM': 0, 'LOW': 0}
for sev, _ in findings:
    counts[sev] += 1
print(f'  DIANA RESULTS: {len(findings)} findings')
for sev in ['CRITICAL', 'HIGH', 'MEDIUM', 'LOW']:
    print(f'    {sev}: {counts[sev]}')
print(f'{"="*60}')

if findings:
    print('\nDetails:')
    for sev, detail in findings:
        print(f'  [{sev}] {detail}')

critical_high = counts['CRITICAL'] + counts['HIGH']
verdict = '🟢 CLEAN' if critical_high == 0 else f'🔴 {critical_high} CRITICAL/HIGH'
print(f'\nVerdict: {verdict}')

sys.exit(0 if critical_high == 0 else 1)
