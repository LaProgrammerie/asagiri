#!/usr/bin/env node
'use strict';

const path = require('path');
const { buildAuditSources, getApplicationRoot } = require('./optimization-audit-sources');
const { writeReportSafely } = require('./report-writer');

const appRoot = getApplicationRoot();
const sources = buildAuditSources(appRoot);

const checks = [
    { name: 'aria labels', ok: /aria-label|aria-labelledby/.test(sources.twig), critical: false },
    { name: 'navigation role', ok: /role="navigation"/.test(sources.twig), critical: false },
    { name: 'focus-visible styles', ok: /focus-visible|:focus/.test(sources.css), critical: false },
    { name: 'skip links', ok: /skip-link|skip-links/.test(sources.twig), critical: false },
];

const report = { generatedAt: new Date().toISOString(), checks };
writeReportSafely(path.join(appRoot, 'tests/E2E/cypress/screenshots/accessibility-test-report.json'), report);

const failedCritical = checks.filter((c) => !c.ok && c.critical);
if (failedCritical.length > 0) {
    process.exit(1);
}
console.log('Accessibility checks completed');
