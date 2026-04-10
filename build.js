const { execSync } = require('child_process');
const fs = require('fs');
const path = require('path');

const distDir = path.join(__dirname, 'dist');
if (!fs.existsSync(distDir)) {
    fs.mkdirSync(distDir);
}

const targets = [
    { os: 'windows', arch: 'amd64', out: 'wapclip-windows.exe' },
    { os: 'darwin',  arch: 'arm64', out: 'wapclip-mac-arm' },
    { os: 'darwin',  arch: 'amd64', out: 'wapclip-mac-intel' },
    { os: 'linux',   arch: 'amd64', out: 'wapclip-linux' }
];

console.log('Starting cross-platform build...');

targets.forEach(t => {
    console.log(`Building for ${t.os}/${t.arch}...`);
    try {
        execSync(`go build -ldflags="-s -w" -o dist/${t.out} ./cmd/`, {
            env: {
                ...process.env,
                GOOS: t.os,
                GOARCH: t.arch,
                CGO_ENABLED: '0'
            },
            stdio: 'inherit'
        });
    } catch (err) {
        console.error(`Failed to build for ${t.os}/${t.arch}: ${err.message}`);
        process.exit(1);
    }
});

console.log('Build complete!');
