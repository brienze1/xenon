(function (host) {
  'use strict';

  const LEVELS = ['log', 'warn', 'debug', 'info', 'error', 'clear'];
  const nativeStringify = JSON.stringify;
  const TYPED_ARRAYS = [Int8Array, Uint8Array, Uint8ClampedArray, Int16Array, Uint16Array, Int32Array, Uint32Array, Float32Array, Float64Array];

  function functionLabel (value) {
    const type = value.constructor && value.constructor.name ? value.constructor.name : 'Function';
    return value.name ? '[' + type + ': ' + value.name + ']' : '[' + type + ']';
  }

  function isTypedArray (value) {
    return TYPED_ARRAYS.some((Type) => value instanceof Type);
  }

  function isBuffer (value) {
    return typeof value === 'object' && value !== null && value._isBuffer === true;
  }

  function indexed (values) {
    const result = {};
    Array.prototype.forEach.call(values, (item, index) => { result[String(index)] = item; });
    return result;
  }

  function replace (value) {
    if (typeof value === 'function') {
      return functionLabel(value);
    }
    if (value instanceof WeakMap) {
      return '[WeakMap]';
    }
    if (value instanceof WeakSet) {
      return '[WeakSet]';
    }
    if (value instanceof ArrayBuffer) {
      return '[ArrayBuffer { byteLength: ' + value.byteLength + ' }]';
    }
    return value;
  }

  function canonical (holder, key, ancestors) {
    const original = holder[key];
    let value = original;
    if (value !== null && typeof value === 'object' && typeof value.toJSON === 'function') {
      value = value.toJSON(key);
    }
    value = replace(value);
    switch (typeof value) {
      case 'undefined':
        return { $type: 'undefined' };
      case 'number':
        if (!Number.isFinite(value)) {
          return { $type: 'number', value: String(value) };
        }
        return value === 0 ? 0 : value;
      case 'bigint':
        return { $type: 'bigint', value: value.toString() };
      case 'symbol':
        return { $type: 'symbol', value: 'Symbol(' + (value.description === undefined ? '' : value.description) + ')' };
      case 'string':
        if (original instanceof Date) {
          return { $type: 'date', value };
        }
        return value;
      case 'boolean':
        return value;
      default:
        break;
    }
    if (value === null) {
      return null;
    }
    if (value.type === 'Buffer' && value.data && isBuffer(original)) {
      return indexed(value.data);
    }
    if (value instanceof RegExp) {
      return {};
    }
    if (ancestors.indexOf(value) !== -1) {
      return { $type: 'circular' };
    }
    ancestors.push(value);
    let result;
    if (value instanceof Map) {
      result = { $type: 'map', value: Array.from(value.entries()).map((entry) => [canonical(entry, 0, ancestors), canonical(entry, 1, ancestors)]) };
    }
    else if (value instanceof Set) {
      const items = Array.from(value.values());
      result = { $type: 'set', value: items.map((item, index) => canonical(items, index, ancestors)) };
    }
    else if (isTypedArray(value)) {
      result = indexed(Array.from(value));
    }
    else if (Array.isArray(value)) {
      result = [];
      for (let index = 0; index < value.length; index++) {
        result.push(canonical(value, index, ancestors));
      }
    }
    else {
      result = {};
      Object.keys(value).forEach((name) => { result[name] = canonical(value, name, ancestors); });
    }
    ancestors.pop();
    return result;
  }

  function text (value) {
    if (typeof value === 'string') {
      return value;
    }
    if (typeof value === 'function') {
      return functionLabel(value);
    }
    if (value instanceof Error) {
      return String(value);
    }
    if (value === null || typeof value !== 'object') {
      return String(value);
    }
    try {
      const encoded = nativeStringify(value, (name, item) => replace(item));
      if (encoded !== undefined) {
        return encoded;
      }
    }
    catch (error) {
      return nativeStringify(canonical({ value }, 'value', []));
    }
    return String(value);
  }

  function dispatch (level, args) {
    const values = args.map((value, index) => canonical(args, index, []));
    host.log(level, nativeStringify(values), args.map(text));
  }

  const console = {};
  LEVELS.forEach((level) => {
    console[level] = (...args) => dispatch(level, args);
  });

  return { console, canonical: (value) => canonical({ value }, 'value', []) };
})
