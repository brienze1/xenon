import { mkdirSync, readdirSync, readFileSync, writeFileSync } from 'node:fs';
import { basename, dirname, join, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';
import { createRequire } from 'node:module';

const require = createRequire(import.meta.url);
const Sandbox = require('postman-sandbox');
const sdk = require('postman-collection');

const here = dirname(fileURLToPath(import.meta.url));
const defaultCorpus = resolve(here, '../../test/integration/resources/scripts/corpus');
const corpusDir = resolve(process.argv[2] || defaultCorpus);
const outDir = resolve(process.argv[3] || join(corpusDir, 'expected'));
const only = process.argv[4];
const EXECUTION_TIMEOUT = 10000;

function canonical (value, seen = new Set()) {
    if (value === undefined) { return { $type: 'undefined' }; }
    if (value === null || typeof value === 'string' || typeof value === 'boolean') { return value; }
    if (typeof value === 'number') { return Number.isFinite(value) ? value : { $type: 'number', value: String(value) }; }
    if (typeof value === 'bigint') { return { $type: 'bigint', value: value.toString() }; }
    if (typeof value === 'function') { return { $type: 'function' }; }
    if (typeof value === 'symbol') { return { $type: 'symbol', value: value.toString() }; }
    if (value instanceof Date) { return { $type: 'date', value: isNaN(value) ? 'Invalid Date' : value.toISOString() }; }
    if (seen.has(value)) { return { $type: 'circular' }; }
    seen.add(value);
    let result;
    if (Array.isArray(value)) {
        result = value.map((item) => canonical(item, seen));
    }
    else if (value instanceof Map) {
        result = { $type: 'map', value: [...value.entries()].map(([k, v]) => [canonical(k, seen), canonical(v, seen)]) };
    }
    else if (value instanceof Set) {
        result = { $type: 'set', value: [...value.values()].map((v) => canonical(v, seen)) };
    }
    else {
        result = {};
        for (const key of Object.keys(value)) { result[key] = canonical(value[key], seen); }
    }
    seen.delete(value);
    return result;
}

function scopeValues (scope) {
    const values = {};
    const list = scope && scope.values ? scope.values : [];
    for (const variable of list) {
        if (variable && !variable.disabled) { values[variable.key] = canonical(variable.value); }
    }
    return values;
}

function keyValues (list) {
    return (list || []).filter((item) => item && !item.disabled).map((item) => ({ key: item.key, value: item.value }));
}

function bodyOf (body) {
    if (!body || !body.mode) { return null; }
    switch (body.mode) {
        case 'raw': return { mode: 'raw', raw: body.raw || '' };
        case 'urlencoded': return { mode: 'urlencoded', urlencoded: keyValues(body.urlencoded) };
        case 'formdata': return { mode: 'formdata', formdata: keyValues(body.formdata) };
        case 'graphql': return { mode: 'graphql', graphql: canonical(body.graphql || {}) };
        default: return { mode: body.mode };
    }
}

function requestOf (raw) {
    if (!raw) { return null; }
    const request = typeof raw === 'string' ? new sdk.Request(raw) : new sdk.Request(raw);
    const json = request.toJSON();
    return {
        method: request.method,
        url: request.url.toString(),
        headers: keyValues(json.header),
        body: bodyOf(json.body)
    };
}

function responseOf (fixture) {
    return new sdk.Response({
        code: fixture.code,
        status: fixture.status,
        header: fixture.header || [],
        body: fixture.body || '',
        responseTime: fixture.responseTime || 0
    }).toJSON();
}

function contextOf (testCase) {
    const context = {
        request: testCase.request || { url: 'https://example.com/', method: 'GET' },
        environment: { name: (testCase.environment || {}).name, values: (testCase.environment || {}).values || [] },
        collectionVariables: { values: testCase.collectionVariables || [] },
        globals: { values: testCase.globals || [] },
        _variables: { values: testCase.variables || [] },
        cookies: testCase.cookies || []
    };
    if (testCase.phase === 'test') {
        context.response = responseOf(testCase.response || { code: 200, status: 'OK' });
    }
    return context;
}

function errorOf (err) {
    if (!err) { return null; }
    return { name: err.name || 'Error', message: err.message || String(err) };
}

function runCase (testCase) {
    return new Promise((resolveCase, rejectCase) => {
        Sandbox.createContext({ timeout: EXECUTION_TIMEOUT }, (createErr, ctx) => {
            if (createErr) { return rejectCase(createErr); }
            const id = testCase.id;
            const outcome = {
                id,
                error: null,
                skipped: false,
                tests: [],
                console: [],
                variables: {},
                request: null,
                sentRequests: [],
                vault: {}
            };
            const sendFixtures = [...(testCase.sendRequest || [])];
            const runFixtures = [...(testCase.runRequest || [])];
            const vault = { ...(testCase.vault || {}) };

            ctx.on('console', (cursor, level, ...args) => {
                outcome.console.push({ level, args: args.map((arg) => canonical(arg)) });
            });
            ctx.on('execution.assertion', (cursor, assertions) => {
                for (const assertion of assertions) {
                    outcome.tests.push({
                        name: assertion.name,
                        passed: Boolean(assertion.passed),
                        skipped: Boolean(assertion.skipped),
                        async: Boolean(assertion.async),
                        error: assertion.error ? assertion.error.message : null
                    });
                }
            });
            ctx.on('execution.error', (cursor, err) => {
                outcome.error = outcome.error || errorOf(err);
            });
            ctx.on('execution.skipRequest.' + id, () => { outcome.skipped = true; });
            ctx.on('execution.request.' + id, (cursor, executionId, eventId, request) => {
                outcome.sentRequests.push(requestOf(request));
                const fixture = sendFixtures.shift();
                if (!fixture) {
                    return ctx.dispatch('execution.response.' + id, eventId, { name: 'Error', message: 'no fixture response' });
                }
                if (fixture.error) {
                    return ctx.dispatch('execution.response.' + id, eventId, { name: 'Error', message: fixture.error });
                }
                ctx.dispatch('execution.response.' + id, eventId, null, responseOf(fixture), {});
            });
            ctx.on('execution.run_collection_request.' + id, (cursor, executionId, eventId, requestId) => {
                outcome.sentRequests.push({ runRequest: requestId });
                const fixture = runFixtures.shift();
                if (!fixture) {
                    return ctx.dispatch('execution.run_collection_request_response.' + id, eventId, { name: 'Error', message: 'no fixture response' });
                }
                ctx.dispatch('execution.run_collection_request_response.' + id, eventId, null, responseOf(fixture), { variableMutations: {} });
            });
            ctx.on('execution.vault.' + id, (eventId, method, key, value) => {
                if (method === 'get') { return ctx.dispatch('execution.vault.' + id, eventId, null, vault[key]); }
                if (method === 'set') { vault[key] = value; return ctx.dispatch('execution.vault.' + id, eventId, null); }
                if (method === 'unset') { delete vault[key]; return ctx.dispatch('execution.vault.' + id, eventId, null); }
                ctx.dispatch('execution.vault.' + id, eventId, { message: 'unknown vault method ' + method });
            });

            ctx.execute({
                listen: testCase.phase,
                script: { exec: testCase.script.split('\n') }
            }, {
                id,
                timeout: EXECUTION_TIMEOUT,
                context: contextOf(testCase)
            }, (err, result) => {
                outcome.error = outcome.error || errorOf(err);
                if (result) {
                    outcome.variables = {
                        environment: scopeValues(result.environment),
                        collectionVariables: scopeValues(result.collectionVariables),
                        globals: scopeValues(result.globals),
                        local: scopeValues(result._variables)
                    };
                    if (testCase.phase === 'prerequest') { outcome.request = requestOf(result.request); }
                }
                outcome.vault = Object.fromEntries(Object.entries(vault).sort(([a], [b]) => a.localeCompare(b)));
                ctx.dispose();
                resolveCase(outcome);
            });
        });
    });
}

function readCases () {
    return readdirSync(corpusDir)
        .filter((name) => name.endsWith('.json'))
        .sort()
        .map((name) => {
            const testCase = JSON.parse(readFileSync(join(corpusDir, name), 'utf8'));
            if (testCase.id !== basename(name, '.json')) {
                throw new Error(name + ': id must match the file name');
            }
            if (!['prerequest', 'test'].includes(testCase.phase)) {
                throw new Error(name + ': phase must be prerequest or test');
            }
            return testCase;
        });
}

const cases = readCases().filter((testCase) => !only || testCase.id === only);
mkdirSync(outDir, { recursive: true });
for (const testCase of cases) {
    const outcome = await runCase(testCase);
    writeFileSync(join(outDir, testCase.id + '.json'), JSON.stringify(outcome, null, 2) + '\n');
}
console.log('Oracle (postman-sandbox 6.7.4) wrote ' + cases.length + ' outcomes to ' + outDir);
