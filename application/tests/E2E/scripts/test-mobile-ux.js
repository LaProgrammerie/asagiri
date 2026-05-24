#!/usr/bin/env node
'use strict';

const path = require('path');
const { buildAuditSources, getApplicationRoot } = require('./optimization-audit-sources');
const { writeReportSafely } = require('./report-writer');

const appRoot = getApplicationRoot();
const sources = buildAuditSources(appRoot);
const bundle = `${sources.twig}\n${sources.css}`;

const checks = [
    { name: 'responsive classes', ok: /sm:|md:|lg:|@media/.test(bundle), critical: true },
    { name: 'mobile nav toggle', ok: /nav-toggle|mobile-menu/.test(bundle), critical: false },
    { name: 'touch target sizing', ok: /min-h-\[44px\]|min-height:\s*44px/.test(bundle), critical: false },
];

const report = { generatedAt: new Date().toISOString(), checks };
writeReportSafely(path.join(appRoot, 'tests/E2E/cypress/screenshots/mobile-ux-test-report.json'), report);

const failedCritical = checks.filter((c) => !c.ok && c.critical);
if (failedCritical.length > 0) {
    process.exit(1);
}
console.log('Mobile UX checks completed');
