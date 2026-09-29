import {test} from 'node:test';
import assert from 'node:assert/strict';
import {mkdtempSync, writeFileSync, rmSync} from 'node:fs';
import {tmpdir} from 'node:os';
import {join} from 'node:path';
import {secret} from './secrets.js';

test('secret files reject ambiguous, empty and unavailable sources without leaking paths', () => {
  const directory = mkdtempSync(join(tmpdir(), 'cafe-secret-'));
  const file = join(directory, 'credential');
  process.env.CAFE_TEST_SECRET_FILE = file;
  try {
    writeFileSync(file, 'sensitive-value\n');
    assert.equal(secret('CAFE_TEST_SECRET'), 'sensitive-value');
    process.env.CAFE_TEST_SECRET = '';
    assert.throws(() => secret('CAFE_TEST_SECRET'), /mutually exclusive/);
    delete process.env.CAFE_TEST_SECRET;
    for (const value of ['', ' \n']) {
      writeFileSync(file, value);
      assert.throws(() => secret('CAFE_TEST_SECRET'), /must not be empty/);
    }
    rmSync(file);
    assert.throws(() => secret('CAFE_TEST_SECRET'), {message: 'CAFE_TEST_SECRET_FILE cannot be read'});
    delete process.env.CAFE_TEST_SECRET_FILE;
    process.env.CAFE_TEST_SECRET = ' value ';
    assert.equal(secret('CAFE_TEST_SECRET'), ' value ');
  } finally {
    delete process.env.CAFE_TEST_SECRET;
    delete process.env.CAFE_TEST_SECRET_FILE;
    rmSync(directory, {recursive: true, force: true});
  }
});
