// Computes the next release version: major.minor from VERSION, patch from the existing v* tags.
// Usage: git tag -l 'v*' | node tools/release/next-version.mjs
import { readFileSync } from 'node:fs';
import { dirname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

const VERSION_LINE = /^(\d+)\.(\d+)$/;
const RELEASE_TAG = /^v(\d+)\.(\d+)\.(\d+)$/;

/**
 * Returns `<major>.<minor>.<patch>` where patch is the highest released patch of the
 * VERSION line plus one, or 0 when that line has no release yet.
 * Throws when versionLine is not `major.minor`.
 */
export function nextVersion(versionLine, tags) {
  const line = String(versionLine).trim();
  const match = VERSION_LINE.exec(line);
  if (!match) {
    throw new Error(`VERSION must be <major>.<minor> (e.g. 0.1), got "${line}"`);
  }
  const major = Number(match[1]);
  const minor = Number(match[2]);
  let highest = -1;
  for (const raw of tags) {
    const tag = RELEASE_TAG.exec(raw.trim());
    if (tag && Number(tag[1]) === major && Number(tag[2]) === minor) {
      highest = Math.max(highest, Number(tag[3]));
    }
  }
  return `${major}.${minor}.${highest + 1}`;
}

function readStdin() {
  try {
    return readFileSync(0, 'utf8');
  } catch {
    return '';
  }
}

function main() {
  const scriptDir = dirname(fileURLToPath(import.meta.url));
  const versionFile = process.env.XENON_VERSION_FILE || resolve(scriptDir, '../../VERSION');
  try {
    const versionLine = readFileSync(versionFile, 'utf8');
    const tags = readStdin().split(/\r?\n/).filter((tag) => tag.trim() !== '');
    process.stdout.write(`${nextVersion(versionLine, tags)}\n`);
  } catch (error) {
    process.stderr.write(`next-version: ${error.message}\n`);
    process.exitCode = 1;
  }
}

if (process.argv[1] && resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  main();
}
