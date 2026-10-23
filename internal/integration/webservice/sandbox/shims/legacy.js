(function (sdk, host, pm, console, global, lazy, record) {
  'use strict';

  const _ = sdk.util;
  const LEGACY_GLOBS = ['tests', 'globals', 'environment', 'data', 'request', 'responseCookies', 'responseHeaders', 'responseTime',
    'responseCode', 'responseBody', 'iteration', 'postman', '_', 'CryptoJS', 'tv4', 'xml2Json', 'Backbone', 'cheerio'];
  const ALTERNATIVES = {
    tests: '"pm.test()"',
    globals: '"pm.globals"',
    environment: '"pm.environment"',
    data: '"pm.iterationData"',
    request: '"pm.request"',
    responseCookies: '"pm.cookies"',
    responseHeaders: '"pm.response.headers"',
    responseTime: '"pm.response.responseTime"',
    responseCode: '"pm.response.code"',
    responseBody: '"pm.response.text()"',
    iteration: '"pm.info.iteration"',
    _: '"require(\'lodash\')"',
    CryptoJS: 'global "crypto" object',
    tv4: '"require(\'ajv\')"',
    xml2Json: '"require(\'xml2js\')"',
    cheerio: '"require(\'cheerio\')"',
    'postman.setNextRequest': '"pm.execution.setNextRequest()"',
    'postman.setEnvironmentVariable': '"pm.environment.set()"',
    'postman.getEnvironmentVariable': '"pm.environment.get()"',
    'postman.clearEnvironmentVariables': '"pm.environment.clear()"',
    'postman.clearEnvironmentVariable': '"pm.environment.unset()"',
    'postman.setGlobalVariable': '"pm.globals.set()"',
    'postman.getGlobalVariable': '"pm.globals.get()"',
    'postman.clearGlobalVariables': '"pm.globals.clear()"',
    'postman.clearGlobalVariable': '"pm.globals.unset()"',
    'postman.getResponseCookie': '"pm.cookies.get()"',
    'postman.getResponseHeader': '"pm.response.headers.get()"'
  };
  const MAX_RESPONSE_SIZE = 50 * 1024 * 1024;

  const nativeJSON = global.JSON;
  const liquidJSON = {
    parse (text, reviver, relaxed) {
      try {
        return nativeJSON.parse(text, reviver);
      }
      catch (error) {
        global.JSON = nativeJSON;
        try {
          return host.load('liquid-json').parse(text, reviver, relaxed);
        }
        finally {
          global.JSON = liquidJSON;
        }
      }
    },
    stringify () {
      try {
        return nativeJSON.stringify.apply(nativeJSON, arguments);
      }
      catch (error) {
        error.name = 'JSONError';
        throw error;
      }
    }
  };
  global.JSON = liquidJSON;

  const used = new Set();
  let warnings = true;

  function deprecated (key) {
    if (!warnings || used.has(key)) {
      return;
    }
    used.add(key);
    if (ALTERNATIVES[key]) {
      console.warn('Using "' + key + '" is deprecated. Use ' + ALTERNATIVES[key] + ' instead.');
    }
    else if (LEGACY_GLOBS.includes(key)) {
      console.warn('Using "' + key + '" is deprecated.');
    }
  }

  function tracked (key, value) {
    if (!(value && ['object', 'function'].includes(typeof value))) {
      return value;
    }
    return new Proxy(value, {
      set (target, property, item) {
        deprecated(key);
        target[property] = item;
        return true;
      },
      get (target, property) {
        if (key !== 'postman') {
          deprecated(key);
        }
        return target[property];
      },
      apply (target, thisArg, args) {
        deprecated(key);
        return target.apply(thisArg, args);
      }
    });
  }

  function bodyData (request) {
    const body = request.body;
    if (!body || body.isEmpty()) {
      return {};
    }
    let content = body[body.mode];
    if (content && _.isFunction(content.all)) {
      content = content.all();
    }
    const collect = (accumulator, param) => {
      if (param.disabled) {
        return accumulator;
      }
      if (accumulator[param.key]) {
        if (Array.isArray(accumulator[param.key])) {
          accumulator[param.key].push(param.value);
        }
        else {
          accumulator[param.key] = [accumulator[param.key], param.value];
        }
      }
      else {
        accumulator[param.key] = param.value;
      }
      return accumulator;
    };
    if (body.mode === 'raw') {
      return content || {};
    }
    if (body.mode === 'urlencoded' || body.mode === 'formdata') {
      return (content || []).reduce(collect, {});
    }
    if (body.mode === 'file') {
      return (body.file && body.file.content) || {};
    }
    return {};
  }

  function xml2Json (text) {
    let json = {};
    host.load('xml2js').parseString(text, { explicitArray: false, async: false, trim: true, mergeAttrs: false }, (error, result) => {
      json = result;
    });
    return json;
  }

  function LegacyInterface (globalvars) {
    Object.defineProperty(this, '__globalvars', { value: globalvars });
  }

  Object.assign(LegacyInterface.prototype, {
    setEnvironmentVariable (key, value) {
      if ((value === false || value) && typeof (value && value.toString) === 'function') {
        value = value.toString();
      }
      this.__globalvars.environment[key] = value;
      return pm.environment.set(key, value);
    },
    getEnvironmentVariable (key) {
      return pm.environment.get(key);
    },
    clearEnvironmentVariables () {
      Object.keys(this.__globalvars.environment).forEach((property) => { delete this.__globalvars.environment[property]; });
      return pm.environment.clear();
    },
    clearEnvironmentVariable (key) {
      if (key) {
        delete this.__globalvars.environment[key];
      }
      return pm.environment.unset(key);
    },
    setGlobalVariable (key, value) {
      if ((value === false || value) && typeof (value && value.toString) === 'function') {
        value = value.toString();
      }
      this.__globalvars.globals[key] = value;
      return pm.globals.set(key, value);
    },
    getGlobalVariable (key) {
      return pm.globals.get(key);
    },
    clearGlobalVariables () {
      Object.keys(this.__globalvars.globals).forEach((property) => { delete this.__globalvars.globals[property]; });
      return pm.globals.clear();
    },
    clearGlobalVariable (key) {
      if (key) {
        delete this.__globalvars.globals[key];
      }
      return pm.globals.unset(key);
    },
    setNextRequest (next) {
      pm.execution.setNextRequest(next);
    }
  });

  function LegacyTestInterface (globalvars) {
    LegacyInterface.call(this, globalvars);
  }

  LegacyTestInterface.prototype = Object.create(LegacyInterface.prototype, { constructor: { value: LegacyTestInterface } });

  Object.assign(LegacyTestInterface.prototype, {
    getResponseCookie (name) {
      return pm.cookies ? pm.cookies.one(String(name)) : undefined;
    },
    getResponseHeader (name) {
      const header = pm.response && pm.response.headers && pm.response.headers.one(name);
      return header ? header.value : undefined;
    }
  });

  const globalvars = {
    tests: {},
    globals: pm.globals.syncVariablesTo(),
    environment: pm.environment.syncVariablesTo(),
    data: {},
    iteration: pm.info.iteration,
    xml2Json
  };
  if (pm.request) {
    globalvars.request = {
      id: pm.info.requestId,
      name: pm.info.requestName,
      description: pm.request.description ? pm.request.description.toString() : undefined,
      headers: pm.request.headers.toObject(true, false, true, true),
      method: pm.request.method,
      url: pm.request.url.toString(),
      data: bodyData(pm.request)
    };
  }
  else {
    globalvars.request = {};
  }
  if (pm.response) {
    globalvars.responseCookies = pm.cookies || [];
    globalvars.responseHeaders = {};
    pm.response.headers.each((header) => {
      if (header && !header.disabled) {
        globalvars.responseHeaders[header.key] = header.value;
      }
    });
    globalvars.responseTime = pm.response.responseTime;
    globalvars.responseCode = Object.assign({}, pm.response.details());
    globalvars.responseBody = pm.response.responseSize > MAX_RESPONSE_SIZE ? undefined : pm.response.text();
  }
  const legacy = new (pm.response ? LegacyTestInterface : LegacyInterface)(globalvars);
  globalvars.postman = new Proxy(legacy, {
    get (target, property) {
      deprecated('postman.' + String(property));
      return target[property];
    }
  });

  Object.keys(globalvars).forEach((key) => {
    global[key] = tracked(key, globalvars[key]);
  });
  Object.keys(lazy).forEach((key) => {
    let loaded = false;
    let loading = false;
    let value;
    Object.defineProperty(global, key, {
      get () {
        if (!loaded && !loading) {
          loading = true;
          try {
            value = tracked(key, lazy[key]());
            loaded = true;
          }
          finally {
            loading = false;
          }
        }
        return value;
      },
      set (replacement) {
        loaded = true;
        value = replacement;
      },
      configurable: true,
      enumerable: false
    });
  });

  return {
    finish () {
      warnings = false;
      const tests = global.tests;
      if (!tests || typeof tests !== 'object' || Object.keys(tests).length === 0) {
        return;
      }
      let index = pm.test.index();
      record(Object.keys(tests).map((key) => {
        const passed = Boolean(tests[key]);
        return {
          name: String(key),
          index: index++,
          passed,
          skipped: false,
          async: false,
          error: passed ? null : 'expected ' + String(tests[key]) + ' to be truthy'
        };
      }));
    }
  };
})
