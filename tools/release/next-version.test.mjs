import assert from 'node:assert/strict';
import { spawnSync } from 'node:child_process';
import { mkdtempSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { dirname, join } from 'node:path';
import { test } from 'node:test';
import { fileURLToPath } from 'node:url';

import { nextVersion } from './next-version.mjs';

const script = join(dirname(fileURLToPath(import.meta.url)), 'next-version.mjs');

test('first release of a line is patch 0', () => {
  assert.equal(nextVersion('0.1\n', []), '0.1.0');
});

test('next patch follows the highest released patch', () => {
  assert.equal(nextVersion('0.1', ['v0.1.0', 'v0.1.1']), '0.1.2');
});

test('a new VERSION line starts at patch 0', () => {
  assert.equal(nextVersion('0.2', ['v0.1.0', 'v0.1.5']), '0.2.0');
});

test('patches compare numerically, not as text', () => {
  assert.equal(nextVersion('0.1', ['v0.1.9', 'v0.1.10']), '0.1.11');
});

test('tags of other lines, without v, or with a suffix are ignored', () => {
  const tags = ['v0.2.7', 'v1.1.3', '0.1.4', 'v0.1.3-beta', 'release-0.1.9', 'v0.1.1'];
  assert.equal(nextVersion('0.1', tags), '0.1.2');
});

test('an invalid VERSION throws', () => {
  assert.throws(() => nextVersion('abc', []), /VERSION must be/);
  assert.throws(() => nextVersion('0.1.0', []), /VERSION must be/);
});

function runCli(versionContent, stdin) {
  const dir = mkdtempSync(join(tmpdir(), 'xenon-next-version-'));
  try {
    const versionFile = join(dir, 'VERSION');
    writeFileSync(versionFile, versionContent);
    return spawnSync(process.execPath, [script], {
      input: stdin,
      encoding: 'utf8',
      env: { ...process.env, XENON_VERSION_FILE: versionFile },
    });
  } finally {
    rmSync(dir, { recursive: true, force: true });
  }
}

test('the CLI reads VERSION and the tags on stdin', () => {
  const result = runCli('0.1\n', 'v0.1.0\nv0.1.1\nv0.0.4\n');
  assert.equal(result.status, 0, result.stderr);
  assert.equal(result.stdout, '0.1.2\n');
});

test('the CLI exits 1 on an invalid VERSION', () => {
  const result = runCli('abc\n', '');
  assert.equal(result.status, 1);
  assert.equal(result.stdout, '');
  assert.match(result.stderr, /VERSION must be/);
});
