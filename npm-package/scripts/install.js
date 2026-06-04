#!/usr/bin/env node
'use strict';

/**
 * Postinstall script for spec-ai.
 *
 * The platform binary ships INSIDE the npm tarball under dist/ — it is
 * extracted by the GitHub Actions release workflow before `npm publish` runs.
 * This script does NOT download anything at install time.
 *
 * What it does:
 *   1. Detect the current platform via detect-platform.js.
 *   2. Locate the matching binary already present in dist/.
 *   3. On POSIX, chmod 0o755 that binary so it is executable.
 *
 * On any failure (unsupported platform, missing binary, chmod error) the
 * script prints a WARNING to stderr and exits 0. Postinstall must be
 * non-fatal — bin/spec-ai.js already has a PATH fallback.
 *
 * Uses only Node.js built-in modules.
 */

const fs = require('fs');
const path = require('path');
const { detectPlatform } = require('./detect-platform');

const DIST_DIR = path.join(__dirname, '..', 'dist');

/**
 * Main postinstall entry point.
 */
function install() {
  // 1. Detect platform.
  let platformKey;
  try {
    platformKey = detectPlatform();
  } catch (err) {
    process.stderr.write(
      `spec-ai WARNING: unsupported platform — ${err.message}\n` +
      'The CLI will fall back to the system PATH binary named "specai".\n'
    );
    process.exit(0);
  }

  // 2. Locate the bundled binary in dist/.
  const isWindows = platformKey.startsWith('windows');
  const binaryName = isWindows ? `specai-${platformKey}.exe` : `specai-${platformKey}`;
  const binaryPath = path.join(DIST_DIR, binaryName);

  if (!fs.existsSync(binaryPath)) {
    process.stderr.write(
      `spec-ai WARNING: expected bundled binary not found at ${binaryPath}\n` +
      'This usually means the package was installed from a source checkout rather\n' +
      'than the published npm tarball. The CLI will fall back to the system PATH\n' +
      'binary named "specai".\n'
    );
    process.exit(0);
  }

  // 3. On POSIX, set the executable bit.
  if (!isWindows) {
    try {
      fs.chmodSync(binaryPath, 0o755);
    } catch (err) {
      process.stderr.write(
        `spec-ai WARNING: could not set executable bit on ${binaryPath}: ${err.message}\n` +
        'The CLI may not be executable. Try: chmod 755 ' + binaryPath + '\n'
      );
      process.exit(0);
    }
  }

  console.log(`spec-ai: binary ready at ${binaryPath}`);
}

install();
