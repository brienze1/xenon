const queue = (fn, ...args) => Promise.resolve().then(() => fn(...args));

export const process = {
    env: {},
    argv: [],
    version: '',
    versions: {},
    platform: 'browser',
    browser: true,
    title: 'browser',
    nextTick: queue,
    cwd: () => '/',
    emitWarning: () => undefined
};
