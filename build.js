const { execSync } = require('child_process');
const fs = require('fs');
const path = require('path');
const os = require('os');

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
        let ldflags = "-s -w";
        if (t.os === 'windows') {
            ldflags += " -H=windowsgui";
        }

        // Enable CGO only when building for Windows on a Windows host to support systray
        const cgoEnabled = (os.platform() === 'win32' && t.os === 'windows') ? '1' : '0';

        execSync(`go build -ldflags="${ldflags}" -o dist/${t.out} ./cmd/`, {
            env: {
                ...process.env,
                GOOS: t.os,
                GOARCH: t.arch,
                CGO_ENABLED: cgoEnabled
            },
            stdio: 'inherit'
        });
    } catch (err) {
        console.warn(`Warning: Failed to build for ${t.os}/${t.arch}: ${err.message}`);
        console.warn(`This is likely due to platform-specific dependencies (like systray requiring CGO on non-Windows).`);
    }
});

console.log('Build complete!');
