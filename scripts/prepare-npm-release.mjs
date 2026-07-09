#!/usr/bin/env node
// Copies goreleaser's per-platform binaries from dist/ into the npm/platforms/*
// packages, stamps every package.json (platform packages + wrapper) with the
// release version, and leaves them ready for `npm publish` in release.yml.
//
// Usage: node scripts/prepare-npm-release.mjs <version-without-v-prefix>

import { readdirSync, statSync, copyFileSync, readFileSync, writeFileSync, chmodSync } from 'node:fs';
import { join } from 'node:path';

const version = process.argv[2];
if (!version) {
  console.error('usage: prepare-npm-release.mjs <version>');
  process.exit(1);
}

const ROOT = new URL('..', import.meta.url).pathname;
const DIST = join(ROOT, 'dist');

const PLATFORMS = [
  { dir: 'darwin-arm64', goos: 'darwin', goarch: 'arm64', binary: 'pokemon-cli' },
  { dir: 'darwin-amd64', goos: 'darwin', goarch: 'amd64', binary: 'pokemon-cli' },
  { dir: 'linux-amd64', goos: 'linux', goarch: 'amd64', binary: 'pokemon-cli' },
  { dir: 'linux-arm64', goos: 'linux', goarch: 'arm64', binary: 'pokemon-cli' },
  { dir: 'windows-amd64', goos: 'windows', goarch: 'amd64', binary: 'pokemon-cli.exe' },
];

function findFilesRecursive(dir) {
  const out = [];
  for (const entry of readdirSync(dir)) {
    const full = join(dir, entry);
    const stat = statSync(full);
    if (stat.isDirectory()) out.push(...findFilesRecursive(full));
    else out.push(full);
  }
  return out;
}

const allDistFiles = findFilesRecursive(DIST);

function bumpJson(path, mutate) {
  const pkg = JSON.parse(readFileSync(path, 'utf8'));
  mutate(pkg);
  writeFileSync(path, `${JSON.stringify(pkg, null, 2)}\n`);
}

for (const platform of PLATFORMS) {
  const match = allDistFiles.find(
    (f) =>
      f.includes(`_${platform.goos}_${platform.goarch}`) &&
      f.endsWith(platform.binary === 'pokemon-cli.exe' ? '.exe' : 'pokemon-cli'),
  );
  if (!match) {
    console.error(`could not find built binary for ${platform.goos}/${platform.goarch} under dist/`);
    process.exit(1);
  }

  const pkgDir = join(ROOT, 'npm', 'platforms', platform.dir);
  const dest = join(pkgDir, platform.binary);
  copyFileSync(match, dest);
  chmodSync(dest, 0o755);

  bumpJson(join(pkgDir, 'package.json'), (pkg) => {
    pkg.version = version;
  });

  console.log(`prepared npm/platforms/${platform.dir} <- ${match}`);
}

bumpJson(join(ROOT, 'npm', 'pokemon-cli', 'package.json'), (pkg) => {
  pkg.version = version;
  for (const dep of Object.keys(pkg.optionalDependencies)) {
    pkg.optionalDependencies[dep] = version;
  }
});

console.log(`stamped npm/pokemon-cli/package.json -> ${version}`);
