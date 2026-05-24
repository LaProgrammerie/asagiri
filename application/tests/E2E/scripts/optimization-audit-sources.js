#!/usr/bin/env node
'use strict';

const fs = require('fs');
const path = require('path');

function getApplicationRoot() {
    return path.resolve(__dirname, '../../..');
}

function readUtf8IfExists(root, relativePath) {
    const fullPath = path.join(root, relativePath);
    if (!fs.existsSync(fullPath)) {
        return '';
    }
    return fs.readFileSync(fullPath, 'utf8');
}

function buildAuditSources(applicationRoot = getApplicationRoot()) {
    const twigParts = [
        readUtf8IfExists(applicationRoot, 'templates/base.html.twig'),
        readUtf8IfExists(applicationRoot, 'templates/default/index.html.twig'),
    ];
    const cssParts = [
        readUtf8IfExists(applicationRoot, 'assets/styles/app.css'),
        readUtf8IfExists(applicationRoot, 'assets/styles/main.css'),
    ];
    const jsParts = [
        readUtf8IfExists(applicationRoot, 'assets/app.js'),
        readUtf8IfExists(applicationRoot, 'assets/bootstrap.js'),
    ];

    return {
        twig: twigParts.join('\n'),
        css: cssParts.join('\n'),
        js: jsParts.join('\n'),
    };
}

module.exports = {
    buildAuditSources,
    getApplicationRoot,
    readUtf8IfExists,
};
