import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';

await import('../public/wasm_exec.js');

assert.equal(typeof globalThis.Go, 'function', 'Go WebAssembly bridge did not load');
const go = new globalThis.Go();
const bytes = await readFile(new URL('../public/tanebi.wasm', import.meta.url));
const result = await WebAssembly.instantiate(bytes, go.importObject);
void go.run(result.instance);

for (let attempt = 0; attempt < 100 && !globalThis.tanebiRun; attempt += 1) {
  await new Promise((resolve) => setTimeout(resolve, 10));
}

assert.equal(typeof globalThis.tanebiRun, 'function', 'TANEBI WebAssembly entry point did not start');
const execution = globalThis.tanebiRun('let 種火 = 2\nrepeat 種火 { print "灯る" }\n');
assert.deepEqual(execution, { output: '灯る\n灯る\n', error: '' });
console.log('[ok] TANEBI WebAssembly executed Unicode source deterministically.');
process.exit(0);
