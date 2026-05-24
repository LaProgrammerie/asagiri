#!/usr/bin/env node
'use strict';

const path = require('path');
const { buildAuditSources, getApplicationRoot } = require('./optimization-audit-sources');
const { writeReportSafely } = require('./report-writer');

const appRoot = getApplicationRoot();
const sources = buildAuditSources(appRoot);

const checks = [
    { name: 'open graph', ok: /og:title|og:description|og:image/.test(sources.twig), critical: true },
    { name: 'schema ld+json', ok: /application\/ld\+json/.test(sources.twig), critical: false },
    { name: 'lazy loading', ok: /loading="lazy"/.test(sources.twig), critical: false },
    { name: 'js perf hints', ok: /requestAnimationFrame|passive:\s*true/.test(sources.js), critical: false },
];

const report = {
    generatedAt: new Date().toISOString(),
    checks,
};

writeReportSafely(
    path.join(appRoot, 'tests/E2E/cypress/screenshots/optimization-test-report.json'),
    report
);

const failedCritical = checks.filter((c) => !c.ok && c.critical);
if (failedCritical.length > 0) {
    console.error('Critical optimization checks failed:', failedCritical.map((c) => c.name).join(', '));
    process.exit(1);
}
console.log('Optimization checks passed');
