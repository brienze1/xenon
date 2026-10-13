(function (host, global) {
  'use strict';

  const methods = host.load('sugar-methods');
  const stubs = [];
  let loaded = false;

  function load () {
    if (loaded) {
      return;
    }
    loaded = true;
    stubs.forEach((stub) => { delete stub.target[stub.name]; });
    host.load('sugar');
  }

  Object.keys(methods).forEach((path) => {
    const parts = path.split('.');
    const owner = global[parts[0]];
    const target = parts[1] ? owner.prototype : owner;
    methods[path].forEach((name) => {
      if (Object.prototype.hasOwnProperty.call(target, name)) {
        return;
      }
      stubs.push({ target, name });
      Object.defineProperty(target, name, {
        get () {
          load();
          return target[name];
        },
        set (value) {
          Object.defineProperty(target, name, { value, writable: true, configurable: true, enumerable: false });
        },
        configurable: true,
        enumerable: false
      });
    });
  });

  return {};
})
