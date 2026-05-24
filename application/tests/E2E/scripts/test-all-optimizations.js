#!/usr/bin/env node
'use strict';

const { execSync } = require('child_process');
const path = require('path');
const { writeReportSafely } = require('./report-writer');

const root = path.resolve(__dirname, '..');
const scripts = [
    { cmd: 'node scripts/test-optimizations.js', category: 'optimization' },
    { cmd: 'node scripts/test-accessibility.js', category: 'accessibility' },
    { cmd: 'node scripts/test-mobile-ux.js', category: 'mobile-ux' },
];

const results = [];
let errors = 0;
for (const script of scripts) {
    try {
        execSync(script.cmd, { cwd: root, stdio: 'inherit' });
        results.push({ category: script.category, status: 'success' });
    } catch (e) {
        results.push({ category: script.category, status: 'error', error: String(e.message || e) });
        errors += 1;
    }
}

writeReportSafely(path.join(root, 'cypress/screenshots/all-optimizations-report.json'), {
    generatedAt: new Date().toISOString(),
    results,
});

process.exit(errors > 0 ? 1 : 0);
