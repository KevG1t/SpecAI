#!/usr/bin/env node
'use strict';

/**
 * Postinstall script for spec-ai.
 *
 * Downloads the correct platform binary from the latest GitHub Release into
 * the dist/ directory. Sets the executable bit on Unix platforms.
 * Uses only Node.js built-in modules — no external npm dependencies.
 */

const fs = require('fs');
const path = require('path');
const https = require('https');
const { detectPlatform } = require('./detect-platform');

const REPO_OWNER = 'KevG1t';
const REPO_NAME = 'SpecAI';
const DIST_DIR = path.join(__dirname, '..', 'dist');

/**
 * Follows HTTP redirects and returns the final response.
 * Handles GitHub's redirect chain (API → CDN).
 *
 * @param {string} url
 * @returns {Promise<import('http').IncomingMessage>}
 */
function getFollowingRedirects(url) {
  return new Promise((resolve, reject) => {
    const options = {
      headers: {
        'User-Agent': 'spec-ai-npm-installer',
        Accept: 'application/octet-stream',
      },
    };

    const request = https.get(url, options, (res) => {
      if (res.statusCode === 301 || res.statusCode === 302 || res.statusCode === 307) {
        const location = res.headers.location;
        if (!location) {
          reject(new Error(`Redirect received without Location header from ${url}`));
          return;
        }
        // Drain the redirect response body before following.
        res.resume();
        resolve(getFollowingRedirects(location));
        return;
      }
      resolve(res);
    });

    request.on('error', reject);
  });
}

/**
 * Downloads a binary from a URL and writes it to outputPath.
 *
 * @param {string} url
 * @param {string} outputPath
 * @returns {Promise<void>}
 */
function downloadBinary(url, outputPath) {
  return new Promise((resolve, reject) => {
    getFollowingRedirects(url)
      .then((res) => {
        if (res.statusCode !== 200) {
          res.resume();
          reject(new Error(`Download failed: HTTP ${res.statusCode} for ${url}`));
          return;
        }

        const file = fs.createWriteStream(outputPath);
        res.pipe(file);

        file.on('finish', () => {
          file.close();
          resolve();
        });

        file.on('error', (err) => {
          fs.unlink(outputPath, () => {});
          reject(err);
        });

        res.on('error', (err) => {
          fs.unlink(outputPath, () => {});
          reject(err);
        });
      })
      .catch(reject);
  });
}

/**
 * Main postinstall entry point.
 */
async function install() {
  let platformKey;
  try {
    platformKey = detectPlatform();
  } catch (err) {
    console.error(`spec-ai: skipping binary download — ${err.message}`);
    process.exit(0);
  }

  const isWindows = platformKey.startsWith('windows');
  const binaryName = isWindows ? `specai-${platformKey}.exe` : `specai-${platformKey}`;
  const binaryUrl =
    `https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/latest/download/${binaryName}`;
  const outputPath = path.join(DIST_DIR, binaryName);

  // Ensure the dist directory exists.
  if (!fs.existsSync(DIST_DIR)) {
    fs.mkdirSync(DIST_DIR, { recursive: true });
  }

  console.log(`spec-ai: downloading ${binaryName} from GitHub Releases...`);

  try {
    await downloadBinary(binaryUrl, outputPath);
  } catch (err) {
    console.error(`spec-ai: download failed — ${err.message}`);
    console.error(
      'You can install manually: ' +
      `https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/latest`
    );
    process.exit(1);
  }

  // Set executable bit on Unix platforms.
  if (!isWindows) {
    try {
      fs.chmodSync(outputPath, 0o755);
    } catch (err) {
      console.warn(`spec-ai: could not set executable bit on ${outputPath}: ${err.message}`);
    }
  }

  console.log(`spec-ai: installed to ${outputPath}`);
}

install();
