#!/usr/bin/env node
const { spawnSync } = require('node:child_process');
const path = require('node:path');

// process.platform/arch -> published platform-package suffix (matches goreleaser GOOS/GOARCH)
const PLATFORMS = {
  'darwin-arm64': '@gyeonghokim/pokemon-cli-darwin-arm64',
  'darwin-x64': '@gyeonghokim/pokemon-cli-darwin-amd64',
  'linux-x64': '@gyeonghokim/pokemon-cli-linux-amd64',
  'linux-arm64': '@gyeonghokim/pokemon-cli-linux-arm64',
  'win32-x64': '@gyeonghokim/pokemon-cli-windows-amd64',
};

const key = `${process.platform}-${process.arch}`;
const pkgName = PLATFORMS[key];

if (!pkgName) {
  console.error(`pokemon-cli: unsupported platform ${key}`);
  process.exit(1);
}

let binPath;
try {
  const binaryName = process.platform === 'win32' ? 'pokemon-cli.exe' : 'pokemon-cli';
  binPath = path.join(path.dirname(require.resolve(`${pkgName}/package.json`)), binaryName);
} catch {
  console.error(
    `pokemon-cli: optional dependency "${pkgName}" is not installed. ` +
      'This usually means npm skipped it for your platform — try reinstalling.',
  );
  process.exit(1);
}

const result = spawnSync(binPath, process.argv.slice(2), { stdio: 'inherit' });
process.exit(result.status ?? 1);
