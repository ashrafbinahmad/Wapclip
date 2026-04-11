#!/usr/bin/env node
const { spawn } = require('child_process');
const path = require('path');
const os = require('os');
const fs = require('fs');

const platform = os.platform();
const arch = os.arch();

const isDev = process.env.WAPCLIP_DEV === '1';
let executable;
let args = process.argv.slice(2);

if (isDev) {
    executable = 'go';
    args = ['run', './cmd/', ...args];
} else {
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

    executable = path.join(__dirname, 'dist', `wapclip-${binSuffix}`);

    if (!fs.existsSync(executable)) {
        console.error(`Could not find the compiled binary for your system at ${executable}`);
        console.error(`Please compile it first via 'npm run build' or check your installation.`);
        process.exit(1);
    }
}

const child = spawn(executable, args, { stdio: 'inherit', shell: platform === 'win32' && isDev });

child.on('error', (err) => {
    console.error(`Failed to start wapclip: ${err.message}`);
    process.exit(1);
});

child.on('exit', (code) => {
    process.exit(code || 0);
});
