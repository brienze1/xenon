import { readFileSync, readdirSync, writeFileSync, mkdirSync } from 'node:fs';
import { join, dirname } from 'node:path';
import { fileURLToPath } from 'node:url';
import codegen from 'postman-code-generators';
import sdk from 'postman-collection';

const root = join(dirname(fileURLToPath(import.meta.url)), '..', '..', 'test', 'integration', 'resources', 'snippets');
const languages = {
  curl: { language: 'curl', variant: 'cURL' },
  js_fetch: { language: 'javascript', variant: 'Fetch' },
  go_native: { language: 'go', variant: 'Native' },
  python_requests: { language: 'python', variant: 'Requests' },
};
const optionSets = JSON.parse(readFileSync(join(root, 'options.json'), 'utf8'));

function generate(language, request, options) {
  return new Promise((resolve, reject) => {
    codegen.convert(language.language, language.variant, request, options, (error, snippet) => (error ? reject(error) : resolve(snippet)));
  });
}

const expectedDir = join(root, 'expected');
mkdirSync(expectedDir, { recursive: true });
for (const file of readdirSync(join(root, 'requests')).filter((name) => name.endsWith('.json')).sort()) {
  const name = file.replace(/\.json$/, '');
  const request = new sdk.Request(JSON.parse(readFileSync(join(root, 'requests', file), 'utf8')));
  for (const [id, language] of Object.entries(languages)) {
    for (const [set, options] of Object.entries(optionSets[id])) {
      const snippet = await generate(language, request, options);
      writeFileSync(join(expectedDir, `${name}.${id}.${set}.txt`), snippet);
    }
  }
}
