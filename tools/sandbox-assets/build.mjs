import { createHash } from 'node:crypto';
import vm from 'node:vm';
import { mkdirSync, readFileSync, readdirSync, rmSync, writeFileSync } from 'node:fs';
import { createRequire } from 'node:module';
import { dirname, join, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';
import * as esbuild from 'esbuild';

const here = dirname(fileURLToPath(import.meta.url));
const require = createRequire(import.meta.url);
const outDir = resolve(here, '../../internal/integration/webservice/sandbox/assets');
const check = process.argv.includes('--check');

const builtins = ['assert', 'buffer', 'events', 'path', 'punycode', 'querystring', 'stream', 'string_decoder', 'timers', 'url', 'util'];

const builtinPackages = {
    assert: 'assert',
    buffer: 'buffer',
    events: 'events',
    path: 'path-browserify',
    punycode: 'punycode',
    querystring: 'querystring-es3',
    stream: 'stream-browserify',
    string_decoder: 'string_decoder',
    timers: 'timers-browserify',
    url: 'url',
    util: 'util'
};

const assets = [
    { module: 'ajv', package: 'ajv', vendor: 'ajv/dist/ajv.bundle.js' },
    { module: 'chai', package: 'chai', vendor: 'chai/chai.js' },
    { module: 'lodash', package: 'lodash', vendor: 'lodash/lodash.js' },
    { module: 'lodash3', package: 'lodash3', vendor: 'lodash3/index.js' },
    { module: 'moment', package: 'moment', vendor: 'moment/min/moment.min.js' },
    { module: 'tv4', package: 'tv4', vendor: 'tv4/tv4.js' },
    { module: 'uuid', package: 'postman-sandbox', vendor: 'postman-sandbox/lib/vendor/uuid.js' },
    { module: 'backbone', package: 'backbone', bundle: 'backbone' },
    { module: 'cheerio', package: 'cheerio', bundle: 'cheerio' },
    { module: 'crypto-js', package: 'crypto-js', bundle: 'crypto-js' },
    { module: 'csv-parse/lib/sync', package: 'csv-parse', bundle: 'csv-parse/lib/sync' },
    { module: 'xml2js', package: 'xml2js', bundle: 'xml2js' },
    { module: 'liquid-json', package: 'liquid-json', bundle: 'liquid-json' },
    { module: 'http-reasons', package: 'http-reasons', bundle: 'http-reasons' },
    { module: 'mime-format', package: 'mime-format', bundle: 'mime-format' },
    { module: 'mime', package: 'mime', bundle: 'mime' },
    { module: 'sugar', package: 'postman-sandbox', vendor: 'postman-sandbox/lib/vendor/sugar.js' },
    { module: 'sugar-methods', package: 'postman-sandbox', generate: sugarMethods },
    ...builtins.map((name) => ({ module: name, package: builtinPackages[name], bundle: builtinPackages[name], builtin: true, ownDependencies: name === 'assert' ? ['util'] : [] })),
    { module: 'intl', package: '@formatjs/intl-numberformat', bundle: join(here, 'intl-entry.js'), intl: true }
];

const intlPackages = [
    '@formatjs/intl-getcanonicallocales',
    '@formatjs/intl-locale',
    '@formatjs/intl-pluralrules',
    '@formatjs/intl-numberformat',
    '@formatjs/intl-datetimeformat'
];

const sugarTargets = ['Object', 'Array', 'String', 'Number', 'Date', 'RegExp', 'Function', 'Boolean'];

function sugarMethods () {
    const source = readFileSync(require.resolve('postman-sandbox/lib/vendor/sugar.js'), 'utf8');
    const context = vm.createContext({});
    const names = () => {
        const result = {};
        for (const target of sugarTargets) {
            result[target] = vm.runInContext('Object.getOwnPropertyNames(' + target + ')', context);
            result[target + '.prototype'] = vm.runInContext('Object.getOwnPropertyNames(' + target + '.prototype)', context);
        }
        return result;
    };
    const before = names();
    vm.runInContext(source, context);
    const after = names();
    const added = {};
    for (const target of Object.keys(after)) {
        const list = after[target].filter((name) => !before[target].includes(name));
        if (list.length) { added[target] = list; }
    }
    return 'module.exports = ' + JSON.stringify(added, null, 2) + ';\n';
}

function fileName (module) {
    return module.replace(/\//g, '-') + '.js';
}

function packageVersion (name) {
    return JSON.parse(readFileSync(join(here, 'node_modules', name, 'package.json'), 'utf8')).version;
}

function normalize (text) {
    return text.replace(/\r\n/g, '\n');
}

function externalBuiltins (names) {
    const filter = new RegExp('^(' + names.join('|') + ')/?$');
    return {
        name: 'external-builtins',
        setup (pluginBuild) {
            pluginBuild.onResolve({ filter }, (args) => ({ path: args.path.replace(/\/$/, ''), external: true }));
        }
    };
}

async function bundle (asset) {
    const external = builtins.filter((name) => !(asset.builtin && name === asset.module) && !(asset.ownDependencies || []).includes(name));
    const result = await esbuild.build({
        entryPoints: [asset.bundle.startsWith('/') || /^[A-Za-z]:\\/.test(asset.bundle) ? asset.bundle : require.resolve(asset.bundle)],
        bundle: true,
        write: false,
        format: 'cjs',
        platform: 'browser',
        target: 'es2017',
        charset: 'utf8',
        legalComments: 'inline',
        logLevel: 'error',
        plugins: [externalBuiltins(external)],
        inject: [join(here, 'process-shim.js')],
        define: { global: 'globalThis' },
        absWorkingDir: here
    });
    return normalize(result.outputFiles[0].text).replaceAll(here.replace(/\\/g, '/'), '.');
}

async function build () {
    const files = [];
    for (const asset of assets) {
        let content;
        if (asset.generate) {
            content = asset.generate();
        }
        else if (asset.vendor) {
            content = normalize(readFileSync(require.resolve(asset.vendor), 'utf8'));
        }
        else {
            content = await bundle(asset);
        }
        const version = asset.intl ?
            intlPackages.map((name) => name + '@' + packageVersion(name)).join(' ') :
            packageVersion(asset.package);
        files.push({
            module: asset.module,
            file: fileName(asset.module),
            name: asset.intl ? 'formatjs-intl' : asset.package,
            version,
            sha256: createHash('sha256').update(content, 'utf8').digest('hex'),
            content
        });
    }
    return files;
}

function manifestOf (files) {
    const entries = files.map(({ content, ...entry }) => entry);
    return JSON.stringify({ reference: 'postman-sandbox@6.7.4', files: entries }, null, 2) + '\n';
}

const files = await build();
const manifest = manifestOf(files);

if (check) {
    const drift = [];
    const expected = new Set(files.map((file) => file.file).concat('MANIFEST.json'));
    for (const file of files) {
        let current = '';
        try { current = readFileSync(join(outDir, file.file), 'utf8'); } catch { current = null; }
        if (current !== file.content) { drift.push(file.file); }
    }
    let currentManifest = null;
    try { currentManifest = readFileSync(join(outDir, 'MANIFEST.json'), 'utf8'); } catch { currentManifest = null; }
    if (currentManifest !== manifest) { drift.push('MANIFEST.json'); }
    for (const name of readdirSync(outDir)) {
        if (!expected.has(name) && name.endsWith('.js')) { drift.push(name + ' (unexpected)'); }
    }
    if (drift.length > 0) {
        console.error('Sandbox assets drifted from the build: ' + drift.join(', '));
        console.error('Run `make sandbox-assets` and commit the result.');
        process.exit(1);
    }
    console.log('Sandbox assets match the build (' + files.length + ' files).');
} else {
    rmSync(outDir, { recursive: true, force: true });
    mkdirSync(outDir, { recursive: true });
    for (const file of files) {
        writeFileSync(join(outDir, file.file), file.content);
    }
    writeFileSync(join(outDir, 'MANIFEST.json'), manifest);
    console.log('Wrote ' + files.length + ' sandbox assets to ' + outDir);
}
