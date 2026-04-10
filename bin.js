#!/usr/bin/env node
const { spawn } = require('child_process');
const path = require('path');
const os = require('os');
const fs = require('fs');

const platform = os.platform();
const arch = os.arch();

let binSuffix = '';

if (platform === 'win32') {
    binSuffix = 'windows.exe';
} else if (platform === 'darwin') {
    if (arch === 'arm64') {
        binSuffix = 'mac-arm';
    } else {
        binSuffix = 'mac-intel';
    }
} else if (platform === 'linux') {
    binSuffix = 'linux';
} else {
    console.error(`Unsupported platform: ${platform} ${arch}`);
    process.exit(1);
}

const executable = path.join(__dirname, 'dist', `wapclip-${binSuffix}`);

if (!fs.existsSync(executable)) {
    console.error(`Could not find the compiled binary for your system at ${executable}`);
    console.error(`Please compile it first via 'npm run build' or check your installation.`);
    process.exit(1);
}

const child = spawn(executable, process.argv.slice(2), { stdio: 'inherit' });

child.on('error', (err) => {
    console.error(`Failed to start wapclip daemon binary: ${err.message}`);
    process.exit(1);
});

child.on('exit', (code) => {
    process.exit(code);
});
