(function (bridge, shims) {
  'use strict';

  const { Request, Response } = shims;

  function parsed(text) {
    return text === undefined ? undefined : JSON.parse(text);
  }

  const variables = Object.freeze({
    has: (key) => bridge.variables.has(String(key)),
    get: (key) => parsed(bridge.variables.get(String(key))),
    set: (key, value) => {
      const encoded = JSON.stringify(value);
      bridge.variables.set(String(key), encoded === undefined ? 'null' : encoded);
    },
    unset: (key) => bridge.variables.unset(String(key)),
    clear: () => bridge.variables.clear(),
    replaceIn: (template) => bridge.variables.replaceIn(String(template)),
    toObject: () => parsed(bridge.variables.toObject()),
  });

  const execution = Object.freeze({
    skipRequest: () => bridge.skipRequest(),
  });

  function sendRequest(definition, callback) {
    const request = definition instanceof Request ? definition : new Request(definition);
    const pending = new Promise((resolve, reject) => {
      bridge.send(JSON.stringify(request.toJSON()), (message, data) => {
        if (message !== null) {
          reject(new Error(message));
          return;
        }
        resolve(new Response(JSON.parse(data)));
      });
    });
    if (typeof callback !== 'function') {
      return pending;
    }
    pending.then((response) => callback(null, response), (error) => callback(error));
    return undefined;
  }

  const request = new Request(JSON.parse(bridge.request));
  const pm = { variables, execution, request, sendRequest };

  return {
    pm,
    request: () => JSON.stringify(pm.request.toJSON()),
  };
})
