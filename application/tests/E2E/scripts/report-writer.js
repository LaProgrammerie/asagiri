#!/usr/bin/env node
'use strict';

const fs = require('fs');
const os = require('os');
const path = require('path');

function writeReportSafely(primaryPath, reportObject, label = 'Detailed report') {
    const payload = JSON.stringify(reportObject, null, 2);
    const targets = [primaryPath];

    const fallback = path.join(os.tmpdir(), path.basename(primaryPath));
    if (fallback !== primaryPath) {
        targets.push(fallback);
    }

    for (const target of targets) {
        try {
            fs.mkdirSync(path.dirname(target), { recursive: true });
            fs.writeFileSync(target, payload);
            console.log(`\nReport saved: ${target} (${label})`);
            return target;
        } catch (error) {
            if (error && (error.code === 'EACCES' || error.code === 'EPERM')) {
                continue;
            }
            throw error;
        }
    }

    console.warn(`\nCould not write ${label}: ${primaryPath}`);
    return null;
}

module.exports = { writeReportSafely };
