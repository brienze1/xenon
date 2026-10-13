(function () {
  'use strict';

  const hasOwn = Object.prototype.hasOwnProperty;
  const PATH_PART = /[^.[\]]+|\[(?:(-?\d+(?:\.\d+)?)|(["'])((?:(?!\2)[^\\]|\\.)*?)\2)\]|(?=(?:\.|\[\])(?:\.|\[\]|$))/g;

  const isObject = (value) => value !== null && (typeof value === 'object' || typeof value === 'function');
  const isString = (value) => typeof value === 'string';
  const isFunction = (value) => typeof value === 'function';
  const isNil = (value) => value === null || value === undefined;
  const isNumber = (value) => typeof value === 'number' || value instanceof Number;
  const has = (object, key) => object !== null && object !== undefined && hasOwn.call(object, key);

  function isPlainObject (value) {
    if (value === null || typeof value !== 'object' || Object.prototype.toString.call(value) !== '[object Object]') {
      return false;
    }
    const proto = Object.getPrototypeOf(value);
    if (proto === null) {
      return true;
    }
    const constructor = hasOwn.call(proto, 'constructor') && proto.constructor;
    return typeof constructor === 'function' && Function.prototype.toString.call(constructor) === Function.prototype.toString.call(Object);
  }

  function isEmpty (value) {
    if (isNil(value)) {
      return true;
    }
    if (typeof value === 'string' || Array.isArray(value) || (typeof value === 'object' && typeof value.length === 'number' && typeof value.splice === 'function')) {
      return value.length === 0;
    }
    if (value instanceof Map || value instanceof Set) {
      return value.size === 0;
    }
    if (typeof value === 'object') {
      return Object.keys(value).length === 0;
    }
    return true;
  }

  function choose () {
    for (let index = 0; index < arguments.length; index++) {
      if (!isEmpty(arguments[index])) {
        return arguments[index];
      }
    }
    return undefined;
  }

  function mergeDefined (target, source) {
    for (const key in source) {
      if (has(source, key) && source[key] !== undefined) {
        target[key] = source[key];
      }
    }
    return target;
  }

  function assignHidden (object, name, value) {
    Object.defineProperty(object, name, { value, configurable: true, enumerable: false, writable: true });
    return object;
  }

  function inherit (child, base) {
    Object.defineProperty(child, 'super_', { value: base, configurable: false, enumerable: false, writable: false });
    child.prototype = Object.create(base.prototype, { constructor: { value: child, enumerable: false, writable: true, configurable: true } });
    return child;
  }

  function inSuperChain (constructor, key, value) {
    return constructor ? (constructor[key] === value || inSuperChain(constructor.super_, key, value)) : false;
  }

  function ensureEncoded (text) {
    try {
      text = decodeURIComponent(text);
    }
    catch (error) {
      text = String(text);
    }
    try {
      return encodeURIComponent(text);
    }
    catch (error) {
      return text;
    }
  }

  function cloneElement (value) {
    if (value && isFunction(value.toJSON)) {
      return value.toJSON();
    }
    if (Array.isArray(value)) {
      return value.map(cloneElement);
    }
    if (value instanceof Date) {
      return new Date(value.getTime());
    }
    if (isPlainObject(value)) {
      const copy = {};
      Object.keys(value).forEach((key) => { copy[key] = cloneElement(value[key]); });
      return copy;
    }
    return value;
  }

  function toPath (path) {
    if (Array.isArray(path)) {
      return path.map(String);
    }
    const text = String(path);
    const result = [];
    if (text.charCodeAt(0) === 46) {
      result.push('');
    }
    text.replace(PATH_PART, (match, number, quote, quoted) => {
      result.push(quote ? quoted.replace(/\\(\\)?/g, '$1') : (number || match));
      return match;
    });
    return result;
  }

  function get (object, path, fallback) {
    if (object !== null && object !== undefined && !Array.isArray(path) && isObject(object) && path in Object(object)) {
      return object[path] === undefined ? fallback : object[path];
    }
    let current = object;
    for (const key of toPath(path)) {
      if (current === null || current === undefined) {
        return fallback;
      }
      current = current[key];
    }
    return current === undefined ? fallback : current;
  }

  function hasPath (object, path) {
    if (object === null || object === undefined) {
      return false;
    }
    if (!Array.isArray(path) && has(object, path)) {
      return true;
    }
    let current = object;
    for (const key of toPath(path)) {
      if (current === null || current === undefined || !hasOwn.call(Object(current), key)) {
        return false;
      }
      current = current[key];
    }
    return true;
  }

  function isEqual (left, right, seen) {
    if (left === right || (left !== left && right !== right)) {
      return true;
    }
    if (!isObject(left) || !isObject(right) || typeof left === 'function' || typeof right === 'function') {
      return false;
    }
    const tag = Object.prototype.toString.call(left);
    if (tag !== Object.prototype.toString.call(right)) {
      return false;
    }
    if (left instanceof Date) {
      return +left === +right;
    }
    if (left instanceof RegExp || left instanceof String) {
      return String(left) === String(right);
    }
    if (left instanceof Number || left instanceof Boolean) {
      return isEqual(left.valueOf(), right.valueOf());
    }
    const visited = seen || new Map();
    if (visited.get(left) === right) {
      return true;
    }
    visited.set(left, right);
    if (left instanceof Map || left instanceof Set) {
      if (left.size !== right.size) {
        return false;
      }
      const rightEntries = Array.from(right.entries());
      return Array.from(left.entries()).every((entry, index) => isEqual(entry, rightEntries[index], visited));
    }
    if (Array.isArray(left)) {
      return left.length === right.length && left.every((item, index) => isEqual(item, right[index], visited));
    }
    if (tag === '[object Object]' && Object.getPrototypeOf(left) !== Object.getPrototypeOf(right) && left.constructor !== right.constructor) {
      return false;
    }
    const leftKeys = Object.keys(left);
    const rightKeys = Object.keys(right);
    return leftKeys.length === rightKeys.length && leftKeys.every((key) => hasOwn.call(right, key) && isEqual(left[key], right[key], visited));
  }

  function forOwn (object, iterator) {
    Object.keys(object || {}).forEach((key) => iterator(object[key], key));
  }

  function isNumeric (value) {
    return !isNaN(parseFloat(value)) && isFinite(value);
  }

  function serialise (property) {
    const json = {};
    Object.keys(property).forEach((key) => {
      const value = property[key];
      if (value === undefined) {
        return;
      }
      let name = key;
      if (value && value._postman_propertyIsList && !value._postman_proprtyIsSerialisedAsPlural && name.endsWith('s')) {
        name = name.slice(0, -1);
      }
      if (value && isFunction(value.toJSON)) {
        json[name] = value.toJSON();
        return;
      }
      json[name] = isString(value) ? value : cloneElement(value);
    });
    return json;
  }

  return {
    util: {
      isObject, isString, isFunction, isNil, isNumber, isPlainObject, isEmpty, has, choose, mergeDefined, assignHidden,
      inherit, inSuperChain, ensureEncoded, cloneElement, toPath, get, hasPath, isEqual, forOwn, isNumeric, serialise
    }
  };
})
