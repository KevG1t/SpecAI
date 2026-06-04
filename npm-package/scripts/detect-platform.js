'use strict';

/**
 * Maps the current Node.js platform and architecture to one of the six
 * supported binary targets used by the spec-ai release workflow.
 *
 * Supported targets: darwin-amd64, darwin-arm64, linux-amd64, linux-arm64,
 * windows-amd64, windows-arm64.
 *
 * @returns {string} Platform key in the form "{os}-{arch}".
 * @throws {Error} When the current platform/arch combination is not supported.
 */
function detectPlatform() {
  const platformMap = {
    darwin: 'darwin',
    linux: 'linux',
    win32: 'windows',
  };

  const archMap = {
    x64: 'amd64',
    arm64: 'arm64',
  };

  const os = platformMap[process.platform];
  if (!os) {
    throw new Error(
      `Unsupported platform: ${process.platform}. ` +
      `spec-ai supports darwin, linux, and win32.`
    );
  }

  const arch = archMap[process.arch];
  if (!arch) {
    throw new Error(
      `Unsupported architecture: ${process.arch} on ${process.platform}. ` +
      `spec-ai supports x64 and arm64.`
    );
  }

  return `${os}-${arch}`;
}

module.exports = { detectPlatform };
