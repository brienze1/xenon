(function (sdk, host, context, canonical) {
  'use strict';

  const _ = sdk.util;
  const { VariableScope, Request, Response, CookieList } = sdk;
  const SCOPES = ['environment', 'collectionVariables', 'globals', 'local'];

  function revive (value) {
    if (Array.isArray(value)) {
      return value.map(revive);
    }
    if (value === null || typeof value !== 'object') {
      return value;
    }
    switch (value.$type) {
      case 'undefined':
        return undefined;
      case 'number':
        return Number(value.value);
      case 'date':
        return new Date(value.value);
      case 'bigint':
        return BigInt(value.value);
      case 'map':
        return new Map(value.value.map((entry) => [revive(entry[0]), revive(entry[1])]));
      case 'set':
        return new Set(value.value.map(revive));
      default:
        break;
    }
    const result = {};
    Object.keys(value).forEach((key) => { result[key] = revive(value[key]); });
    return result;
  }

  function scopeOf (definition) {
    const source = definition || {};
    const values = (source.values || []).map((variable) => Object.assign({}, variable, { value: revive(variable.value) }));
    const scope = new VariableScope({ values });
    if (source.name !== undefined) {
      scope.name = source.name;
    }
    return scope;
  }

  function readonly (target, properties) {
    Object.keys(properties).forEach((name) => {
      if (properties[name] !== undefined) {
        Object.defineProperty(target, name, { value: properties[name], writable: false });
      }
    });
    return target;
  }

  const scopes = {};
  SCOPES.forEach((name) => { scopes[name] = scopeOf(context.scopes && context.scopes[name]); });
  const environmentActive = !(context.scopes && context.scopes.environment && context.scopes.environment.active === false);
  if (!environmentActive) {
    ['set', 'unset', 'clear'].forEach((method) => {
      Object.defineProperty(scopes.environment, method, {
        value (key) {
          const argument = method === 'clear' ? '' : JSON.stringify(String(key));
          host.notice('warn', 'pm.environment.' + method + '(' + argument + ') was ignored: no environment is active');
        },
        writable: true,
        configurable: true
      });
    });
  }
  const iterationData = new VariableScope();
  const local = scopes.local;
  local.addLayer(iterationData.values);
  local.addLayer(scopes.environment.values);
  local.addLayer(scopes.collectionVariables.values);
  local.addLayer(scopes.globals.values);

  const vault = Object.assign({}, context.vault || {});
  const tests = [];
  const phase = context.phase;
  const request = context.request ? new Request(context.request) : undefined;
  const response = context.response ? new Response(context.response) : undefined;
  const cookies = new CookieList(null, context.cookies || []);
  cookies.jar = function () {
    throw new Error('cookie jar not available yet');
  };

  let chai;
  function loadChai () {
    if (!chai) {
      chai = host.load('chai');
      if (!chai.__xenonPostman) {
        chai.use(sdk.chaiPostman);
        Object.defineProperty(chai, '__xenonPostman', { value: true });
      }
    }
    return chai;
  }

  function Postman () {
    let assertionIndex = 0;
    const record = (assertion) => {
      tests.push({
        name: assertion.name,
        index: assertion.index,
        passed: assertion.passed,
        skipped: assertion.skipped,
        async: assertion.async,
        error: assertion.error ? errorMessage(assertion.error) : null
      });
    };
    const create = (name, skipped) => ({ name: String(name), async: false, skipped: Boolean(skipped), passed: true, error: null, index: assertionIndex++ });
    const fail = (assertion, error) => {
      assertion.error = error;
      assertion.passed = false;
    };

    const execution = readonly({}, {
      skipRequest () {
        if (phase !== 'prerequest') {
          throw new TypeError('pm.execution.skipRequest is not a function');
        }
        host.skip();
      },
      location: readonly((context.location && context.location.path ? context.location.path.slice() : []), { current: context.location ? context.location.current : undefined }),
      setNextRequest (next) {
        return next && undefined;
      },
      runRequest (requestId, options) {
        return new Promise((resolve, reject) => {
          if (!requestId) {
            reject(new Error('runRequest: collection request id not provided'));
            return;
          }
          host.run(String(requestId), JSON.stringify(options || {}), currentScopes(), (message, data, returned) => {
            if (message !== null) {
              reject(new Error(message));
              return;
            }
            if (returned) {
              adoptScopes(JSON.parse(returned));
            }
            resolve(data === null ? null : new Response(JSON.parse(data)));
          });
        });
      }
    });

    readonly(this, {
      info: readonly({}, {
        eventName: phase,
        iteration: context.info ? context.info.iteration : undefined,
        iterationCount: context.info ? context.info.iterationCount : undefined,
        requestName: context.info ? context.info.requestName : undefined,
        requestId: context.info ? context.info.requestId : undefined
      }),
      vault: {
        get: (key) => Promise.resolve(_.has(vault, key) ? vault[key] : null),
        set: (key, value) => {
          vault[key] = value;
          return Promise.resolve();
        },
        unset: (key) => {
          delete vault[key];
          return Promise.resolve();
        }
      },
      datasets () {
        throw host.unsupported('pm.datasets');
      },
      globals: scopes.globals,
      environment: scopes.environment,
      collectionVariables: scopes.collectionVariables,
      variables: local,
      iterationData,
      request,
      response,
      cookies,
      visualizer: {
        set (template, data, options) {
          if (typeof template !== 'string') {
            throw new Error('Invalid template. Template must be of type string, found ' + typeof template);
          }
          if (data && typeof data !== 'object') {
            throw new Error('Invalid data. Data must be an object, found ' + typeof data);
          }
          if (options && typeof options !== 'object') {
            throw new Error('Invalid options. Options must be an object, found ' + typeof options);
          }
          host.notice('info', 'pm.visualizer is not rendered in Xenon');
        },
        clear () {
          return undefined;
        }
      },
      sendRequest (definition, callback) {
        let history;
        const self = this;
        const pending = () => new Promise((resolve, reject) => {
          if (!definition) {
            reject(new Error('sendrequest: nothing to request'));
            return;
          }
          const outgoing = Request.isRequest(definition) ? definition : new Request(definition);
          host.send(JSON.stringify(exportRequest(outgoing)), currentScopes(), (message, data) => {
            history = { cookies: new CookieList({}, []) };
            if (message !== null) {
              reject(new Error(message));
              return;
            }
            resolve(new Response(JSON.parse(data)));
          });
        });
        if (_.isFunction(callback)) {
          pending().then((received) => callback.call(self, null, received, history), (error) => callback.call(self, error));
          return self;
        }
        return pending();
      },
      execution
    });

    const pm = this;
    this.test = function (name, assert) {
      const assertion = create(name, false);
      if (typeof assert !== 'function') {
        record(assertion);
        return pm;
      }
      if (assert.length) {
        try {
          assertion.async = true;
          assert(function (error) {
            if (assertion.error && assertion.passed === false) {
              return;
            }
            if (error) {
              fail(assertion, error);
            }
            record(assertion);
          });
        }
        catch (error) {
          fail(assertion, error);
          record(assertion);
        }
        return pm;
      }
      try {
        assert();
      }
      catch (error) {
        fail(assertion, error);
      }
      record(assertion);
      return pm;
    };
    this.test.skip = function (name) {
      record(create(name, true));
      return pm;
    };
    this.test.index = function () {
      return assertionIndex;
    };

    if (this.response) {
      Object.defineProperty(this.response, 'to', { get () { return loadChai().expect(this).to; } });
    }
    if (this.request) {
      Object.defineProperty(this.request, 'to', { get () { return loadChai().expect(this).to; } });
    }
  }

  Object.defineProperty(Postman.prototype, 'expect', { get () { return loadChai().expect; } });

  function errorMessage (error) {
    if (error && typeof error === 'object' && 'message' in error) {
      return error.message === undefined ? null : String(error.message);
    }
    return null;
  }

  function serialised (value) {
    return value && typeof value.toJSON === 'function' ? value.toJSON() : value;
  }

  function exportRequest (outgoing) {
    return {
      method: outgoing.method,
      url: String(outgoing.url),
      header: serialised(outgoing.headers),
      body: serialised(outgoing.body),
      auth: serialised(outgoing.auth)
    };
  }

  function transported (value) {
    if (value === undefined) {
      return undefined;
    }
    const encoded = JSON.stringify(value);
    return encoded === undefined ? undefined : JSON.parse(encoded);
  }

  function scopeValues (scope) {
    return scope.values.all().map((variable) => {
      const entry = { key: variable.key, value: canonical(transported(variable.valueOf())) };
      if (variable.type && variable.type !== 'any') {
        entry.type = variable.type;
      }
      if (variable.disabled) {
        entry.disabled = true;
      }
      if (variable.secret) {
        entry.secret = true;
      }
      return entry;
    });
  }

  function currentScopes () {
    const current = {};
    SCOPES.forEach((name) => { current[name] = { values: scopeValues(scopes[name]) }; });
    return JSON.stringify(current);
  }

  function adoptScopes (returned) {
    SCOPES.forEach((name) => {
      if (name === 'local' || !returned[name]) {
        return;
      }
      const values = returned[name].values.map((variable) => Object.assign({}, variable, { value: revive(variable.value) }));
      scopes[name].values.repopulate(values);
    });
  }

  const pm = new Postman();

  return {
    pm,
    record (assertions) {
      assertions.forEach((assertion) => tests.push(assertion));
    },
    result () {
      return JSON.stringify({
        request: pm.request ? exportRequest(pm.request) : null,
        scopes: JSON.parse(currentScopes()),
        vault,
        tests
      });
    }
  };
})
