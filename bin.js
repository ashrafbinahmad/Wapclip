#!/usr/bin/env node
const { spawn } = require('child_process');
const path = require('path');

const cmd = 'go';
const args = ['run', path.join(__dirname, 'cmd')].concat(process.argv.slice(2));

const child = spawn(cmd, args, { stdio: 'inherit' });
child.on('exit', (code) => {
    process.exit(code);
});
