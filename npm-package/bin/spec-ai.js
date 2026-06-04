#!/usr/bin/env node
'use strict';

/**
 * Entry point for the spec-ai CLI.
 *
 * Resolves the platform binary from the dist/ directory that was populated by
 * the postinstall script. Falls back to the system PATH binary named "specai"
 * if the dist binary is not present (e.g. after a postinstall failure or a
 * source checkout without running npm install).
 */

const fs = require('fs');
const path = require('path');
const { execFileSync } = require('child_process');
const { spawnSync } = require('child_process');
const { detectPlatform } = require('../scripts/detect-platform');

/**
 * Resolves the absolute path to the specai binary for the current platform.
 * Returns the path to the dist binary if it exists, otherwise returns 'specai'
 * for PATH resolution.
 *
 * @returns {string} Absolute path or bare command name.
 */
function resolveBinaryPath() {
  let platformKey;
  try {
    platformKey = detectPlatform();
  } catch (_) {
    // Unsupported platform — fall back to system PATH.
    return 'specai';
  }

  const isWindows = platformKey.startsWith('windows');
  const binaryName = isWindows ? `specai-${platformKey}.exe` : `specai-${platformKey}`;
  const distPath = path.join(__dirname, '..', 'dist', binaryName);

  if (fs.existsSync(distPath)) {
    return distPath;
  }

  // Fall back to the system PATH binary.
  return 'specai';
}

/**
 * Executes the specai binary, forwarding all arguments and inheriting stdio.
 */
function run() {
  const binaryPath = resolveBinaryPath();
  const args = process.argv.slice(2);

  const result = spawnSync(binaryPath, args, {
    stdio: 'inherit',
    shell: false,
  });

  if (result.error) {
    if (result.error.code === 'ENOENT') {
      process.stderr.write(
        'spec-ai: binary not found. Run "npm install -g spec-ai" to reinstall.\n'
      );
    } else {
      process.stderr.write(`spec-ai: error executing binary: ${result.error.message}\n`);
    }
    process.exit(1);
  }

  process.exit(result.status || 0);
}

run();
