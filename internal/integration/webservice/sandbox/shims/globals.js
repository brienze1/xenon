(function (host, global) {
  'use strict';

  const ERROR_NAMES = {
    IndexSizeError: 1, HierarchyRequestError: 3, WrongDocumentError: 4, InvalidCharacterError: 5, NoModificationAllowedError: 7,
    NotFoundError: 8, NotSupportedError: 9, InvalidStateError: 11, SyntaxError: 12, InvalidModificationError: 13, NamespaceError: 14,
    InvalidAccessError: 15, TypeMismatchError: 17, SecurityError: 18, NetworkError: 19, AbortError: 20, URLMismatchError: 21,
    QuotaExceededError: 22, TimeoutError: 23, InvalidNodeTypeError: 24, DataCloneError: 25
  };
  const DIGESTS = { 'SHA-1': 'SHA-1', 'SHA-256': 'SHA-256', 'SHA-384': 'SHA-384', 'SHA-512': 'SHA-512' };
  const SPECIAL_PORTS = { 'ftp:': '21', 'file:': '', 'http:': '80', 'https:': '443', 'ws:': '80', 'wss:': '443' };

  class DOMException extends Error {
    constructor (message, name) {
      super(message === undefined ? '' : String(message));
      Object.defineProperty(this, 'name', { value: name === undefined ? 'Error' : String(name), configurable: true, writable: true });
      Object.defineProperty(this, 'code', { value: ERROR_NAMES[this.name] || 0, configurable: true });
    }
  }

  function setHidden (object, name, value) {
    Object.defineProperty(object, name, { value, writable: true, configurable: true, enumerable: false });
  }

  function bytesOf (source) {
    if (source instanceof ArrayBuffer) {
      return new Uint8Array(source);
    }
    if (ArrayBuffer.isView(source)) {
      return new Uint8Array(source.buffer, source.byteOffset, source.byteLength);
    }
    throw new TypeError('The provided value is not of type \'(ArrayBuffer or ArrayBufferView)\'');
  }

  function bufferOf (bytes) {
    return bytes.buffer.slice(bytes.byteOffset, bytes.byteOffset + bytes.byteLength);
  }

  class Event {
    constructor (type, options) {
      if (arguments.length === 0) {
        throw new TypeError('The "type" argument must be specified');
      }
      const settings = options || {};
      setHidden(this, '_stop', false);
      this.type = String(type);
      this.bubbles = Boolean(settings.bubbles);
      this.cancelable = Boolean(settings.cancelable);
      this.composed = Boolean(settings.composed);
      this.defaultPrevented = false;
      this.target = null;
      this.currentTarget = null;
      this.timeStamp = Date.now();
      this.isTrusted = false;
    }

    preventDefault () {
      if (this.cancelable) {
        this.defaultPrevented = true;
      }
    }

    stopPropagation () {
      return undefined;
    }

    stopImmediatePropagation () {
      this._stop = true;
    }
  }

  class EventTarget {
    constructor () {
      setHidden(this, '_listeners', new Map());
    }

    addEventListener (type, listener, options) {
      if (!listener) {
        return;
      }
      const once = typeof options === 'object' && options !== null && Boolean(options.once);
      const list = this._listeners.get(type) || [];
      if (list.some((entry) => entry.listener === listener)) {
        return;
      }
      list.push({ listener, once });
      this._listeners.set(type, list);
    }

    removeEventListener (type, listener) {
      const list = this._listeners.get(type);
      if (list) {
        this._listeners.set(type, list.filter((entry) => entry.listener !== listener));
      }
    }

    dispatchEvent (event) {
      if (!(event instanceof Event)) {
        throw new TypeError('The "event" argument must be an instance of Event. Received ' + typeof event);
      }
      event.target = this;
      event.currentTarget = this;
      const handler = this['on' + event.type];
      const list = (this._listeners.get(event.type) || []).slice();
      if (typeof handler === 'function') {
        handler.call(this, event);
      }
      for (const entry of list) {
        if (entry.once) {
          this.removeEventListener(event.type, entry.listener);
        }
        if (typeof entry.listener === 'function') {
          entry.listener.call(this, event);
        }
        else if (entry.listener && typeof entry.listener.handleEvent === 'function') {
          entry.listener.handleEvent(event);
        }
        if (event._stop) {
          break;
        }
      }
      event.currentTarget = null;
      return !event.defaultPrevented;
    }
  }

  class AbortSignal extends EventTarget {
    constructor () {
      super();
      this.aborted = false;
      this.reason = undefined;
      this.onabort = null;
    }

    throwIfAborted () {
      if (this.aborted) {
        throw this.reason;
      }
    }

    static abort (reason) {
      const signal = new AbortSignal();
      signal.aborted = true;
      signal.reason = reason === undefined ? new DOMException('This operation was aborted', 'AbortError') : reason;
      return signal;
    }

    static timeout (milliseconds) {
      const signal = new AbortSignal();
      global.setTimeout(() => {
        signal.aborted = true;
        signal.reason = new DOMException('The operation was aborted due to timeout', 'TimeoutError');
        signal.dispatchEvent(new Event('abort'));
      }, milliseconds);
      return signal;
    }
  }

  class AbortController {
    constructor () {
      this.signal = new AbortSignal();
    }

    abort (reason) {
      if (this.signal.aborted) {
        return;
      }
      this.signal.aborted = true;
      this.signal.reason = reason === undefined ? new DOMException('This operation was aborted', 'AbortError') : reason;
      this.signal.dispatchEvent(new Event('abort'));
    }
  }

  class TextEncoder {
    get encoding () {
      return 'utf-8';
    }

    encode (input) {
      return new Uint8Array(host.encodeUtf8(input === undefined ? '' : String(input)));
    }

    encodeInto (input, destination) {
      const encoded = this.encode(input);
      const written = Math.min(encoded.length, destination.length);
      destination.set(encoded.subarray(0, written));
      return { read: String(input).length, written };
    }
  }

  class TextDecoder {
    constructor (label, options) {
      const encoding = label === undefined ? 'utf-8' : String(label).trim().toLowerCase();
      if (!['utf-8', 'utf8', 'unicode-1-1-utf-8'].includes(encoding)) {
        throw new RangeError('The "' + label + '" encoding is not supported');
      }
      const settings = options || {};
      setHidden(this, '_fatal', Boolean(settings.fatal));
      setHidden(this, '_ignoreBOM', Boolean(settings.ignoreBOM));
    }

    get encoding () {
      return 'utf-8';
    }

    get fatal () {
      return this._fatal;
    }

    get ignoreBOM () {
      return this._ignoreBOM;
    }

    decode (input) {
      if (input === undefined) {
        return '';
      }
      const decoded = host.decodeUtf8(bufferOf(bytesOf(input)), this._fatal);
      if (decoded === null) {
        throw new TypeError('The encoded data was not valid for encoding utf-8');
      }
      return !this._ignoreBOM && decoded.charCodeAt(0) === 0xFEFF ? decoded.slice(1) : decoded;
    }
  }

  function atob (data) {
    if (arguments.length === 0) {
      throw new TypeError('The "data" argument must be specified');
    }
    const decoded = host.atob(String(data));
    if (decoded === null) {
      throw new DOMException('Invalid character', 'InvalidCharacterError');
    }
    return decoded;
  }

  function btoa (data) {
    if (arguments.length === 0) {
      throw new TypeError('The "data" argument must be specified');
    }
    const encoded = host.btoa(String(data));
    if (encoded === null) {
      throw new DOMException('Invalid character', 'InvalidCharacterError');
    }
    return encoded;
  }

  class Blob {
    constructor (parts, options) {
      const chunks = [];
      (parts || []).forEach((part) => {
        if (part instanceof Blob) {
          chunks.push(part._bytes);
        }
        else if (part instanceof ArrayBuffer || ArrayBuffer.isView(part)) {
          chunks.push(new Uint8Array(bytesOf(part)));
        }
        else {
          chunks.push(new TextEncoder().encode(String(part)));
        }
      });
      const size = chunks.reduce((total, chunk) => total + chunk.length, 0);
      const bytes = new Uint8Array(size);
      let offset = 0;
      chunks.forEach((chunk) => {
        bytes.set(chunk, offset);
        offset += chunk.length;
      });
      setHidden(this, '_bytes', bytes);
      const type = options && options.type !== undefined ? String(options.type) : '';
      setHidden(this, '_type', /^[\x20-\x7E]*$/.test(type) ? type.toLowerCase() : '');
    }

    get size () {
      return this._bytes.length;
    }

    get type () {
      return this._type;
    }

    text () {
      return Promise.resolve(new TextDecoder().decode(this._bytes));
    }

    arrayBuffer () {
      return Promise.resolve(bufferOf(this._bytes));
    }

    slice (start, end, type) {
      return new Blob([this._bytes.slice(start, end)], { type });
    }
  }

  function cloneValue (value, seen) {
    if (value === null || (typeof value !== 'object' && typeof value !== 'function')) {
      if (typeof value === 'symbol') {
        throw new DOMException(String(value) + ' could not be cloned.', 'DataCloneError');
      }
      return value;
    }
    if (typeof value === 'function') {
      throw new DOMException((value.toString().split('\n')[0]) + ' could not be cloned.', 'DataCloneError');
    }
    if (seen.has(value)) {
      return seen.get(value);
    }
    let copy;
    if (value instanceof Date) {
      copy = new Date(value.getTime());
    }
    else if (value instanceof RegExp) {
      copy = new RegExp(value.source, value.flags);
    }
    else if (value instanceof ArrayBuffer) {
      copy = value.slice(0);
    }
    else if (ArrayBuffer.isView(value)) {
      copy = new value.constructor(value);
    }
    else if (value instanceof Map) {
      copy = new Map();
      seen.set(value, copy);
      value.forEach((item, key) => copy.set(cloneValue(key, seen), cloneValue(item, seen)));
      return copy;
    }
    else if (value instanceof Set) {
      copy = new Set();
      seen.set(value, copy);
      value.forEach((item) => copy.add(cloneValue(item, seen)));
      return copy;
    }
    else if (value instanceof Error) {
      copy = new (global[value.name] && global[value.name].prototype instanceof Error ? global[value.name] : Error)(value.message);
    }
    else if (Array.isArray(value)) {
      copy = [];
      seen.set(value, copy);
      value.forEach((item, index) => { copy[index] = cloneValue(item, seen); });
      return copy;
    }
    else {
      copy = {};
      seen.set(value, copy);
      Object.keys(value).forEach((key) => { copy[key] = cloneValue(value[key], seen); });
      return copy;
    }
    seen.set(value, copy);
    return copy;
  }

  function structuredClone (value) {
    if (arguments.length === 0) {
      throw new TypeError('The "value" argument must be specified');
    }
    return cloneValue(value, new Map());
  }

  function queueMicrotask (callback) {
    if (typeof callback !== 'function') {
      throw new TypeError('The "callback" argument must be of type function. Received ' + (callback === null ? 'null' : typeof callback));
    }
    Promise.resolve().then(() => callback());
  }

  const keys = new WeakMap();

  class CryptoKey {
    constructor () {
      throw new TypeError('Illegal constructor');
    }
  }

  function hashName (algorithm) {
    const name = typeof algorithm === 'string' ? algorithm : algorithm && algorithm.name;
    const normalized = DIGESTS[String(name).toUpperCase()];
    if (!normalized) {
      throw new DOMException('Unrecognized algorithm name', 'NotSupportedError');
    }
    return normalized;
  }

  function hmacKey (key) {
    const material = keys.get(key);
    if (!material) {
      throw new TypeError('The "key" argument must be an instance of CryptoKey.');
    }
    return material;
  }

  function algorithmName (algorithm) {
    return String(typeof algorithm === 'string' ? algorithm : algorithm && algorithm.name).toUpperCase();
  }

  function settle (work) {
    return new Promise((resolve) => resolve(work()));
  }

  const subtle = {
    digest (algorithm, data) {
      return settle(() => host.digest(hashName(algorithm), bufferOf(bytesOf(data))));
    },
    importKey (format, keyData, algorithm, extractable, usages) {
      return settle(() => {
        if (format !== 'raw' || algorithmName(algorithm) !== 'HMAC') {
          throw new DOMException(algorithmName(algorithm) + ' keys in the ' + format + ' format are not supported in Xenon', 'NotSupportedError');
        }
        const hash = hashName(algorithm.hash);
        const bytes = new Uint8Array(bytesOf(keyData));
        const key = Object.create(CryptoKey.prototype);
        Object.defineProperties(key, {
          type: { value: 'secret', enumerable: true },
          extractable: { value: Boolean(extractable), enumerable: true },
          algorithm: { value: { name: 'HMAC', hash: { name: hash }, length: bytes.length * 8 }, enumerable: true },
          usages: { value: Array.from(usages || []), enumerable: true }
        });
        keys.set(key, { hash, bytes: bufferOf(bytes) });
        return key;
      });
    },
    sign (algorithm, key, data) {
      return settle(() => {
        const material = hmacKey(key);
        return host.hmac(material.hash, material.bytes, bufferOf(bytesOf(data)));
      });
    },
    verify (algorithm, key, signature, data) {
      return settle(() => {
        const material = hmacKey(key);
        return host.hmacEqual(material.hash, material.bytes, bufferOf(bytesOf(signature)), bufferOf(bytesOf(data)));
      });
    },
    exportKey (format, key) {
      return settle(() => {
        const material = hmacKey(key);
        if (format !== 'raw' || !key.extractable) {
          throw new DOMException('key is not extractable', 'InvalidAccessError');
        }
        return material.bytes.slice(0);
      });
    }
  };

  const crypto = {
    getRandomValues (array) {
      if (!ArrayBuffer.isView(array) || array instanceof Float32Array || array instanceof Float64Array || array instanceof DataView) {
        throw new DOMException('The data argument must be an integer-type TypedArray', 'TypeMismatchError');
      }
      if (array.byteLength > 65536) {
        throw new DOMException('The ArrayBufferView\'s byte length (' + array.byteLength + ') exceeds the number of bytes of entropy available via this API (65536)', 'QuotaExceededError');
      }
      new Uint8Array(array.buffer, array.byteOffset, array.byteLength).set(new Uint8Array(host.randomBytes(array.byteLength)));
      return array;
    },
    randomUUID () {
      return host.uuid();
    },
    subtle
  };

  function unsupported (name) {
    return () => {
      throw host.unsupported(name);
    };
  }

  function SharedArrayBuffer (length) {
    if (arguments.length > 0 || !new.target) {
      throw host.unsupported('SharedArrayBuffer');
    }
    Object.defineProperty(this, Symbol.toStringTag, { value: 'SharedArrayBuffer' });
    Object.defineProperty(this, 'byteLength', { value: length === undefined ? 0 : length });
  }

  const Atomics = {};
  ['add', 'and', 'compareExchange', 'exchange', 'isLockFree', 'load', 'notify', 'or', 'store', 'sub', 'wait', 'waitAsync', 'xor'].forEach((name) => {
    Atomics[name] = unsupported('Atomics');
  });

  function searchParamsEncode (text) {
    return encodeURIComponent(text).replace(/%20/g, '+').replace(/[!'()~]/g, (character) => '%' + character.charCodeAt(0).toString(16).toUpperCase());
  }

  function searchParamsDecode (text) {
    const replaced = text.replace(/\+/g, ' ');
    try {
      return decodeURIComponent(replaced);
    }
    catch (error) {
      return replaced.replace(/%([0-9A-Fa-f]{2})/g, (match, hex) => String.fromCharCode(parseInt(hex, 16)));
    }
  }

  class URLSearchParams {
    constructor (init) {
      setHidden(this, '_list', []);
      setHidden(this, '_url', null);
      if (init === undefined || init === null) {
        return;
      }
      if (init instanceof URLSearchParams) {
        this._list = init._list.map((pair) => pair.slice());
      }
      else if (typeof init === 'object' && typeof init[Symbol.iterator] === 'function') {
        for (const pair of init) {
          const items = Array.from(pair);
          if (items.length !== 2) {
            throw new TypeError('Each query pair must be an iterable [name, value] tuple');
          }
          this._list.push([String(items[0]), String(items[1])]);
        }
      }
      else if (typeof init === 'object') {
        Object.keys(init).forEach((key) => this._list.push([key, String(init[key])]));
      }
      else {
        this._parse(String(init));
      }
    }

    _parse (text) {
      this._list = [];
      const query = text.charAt(0) === '?' ? text.slice(1) : text;
      query.split('&').forEach((part) => {
        if (!part) {
          return;
        }
        const index = part.indexOf('=');
        const name = index < 0 ? part : part.slice(0, index);
        const value = index < 0 ? '' : part.slice(index + 1);
        this._list.push([searchParamsDecode(name), searchParamsDecode(value)]);
      });
    }

    _update () {
      if (this._url) {
        const serialized = this.toString();
        this._url._query = serialized === '' ? null : serialized;
      }
    }

    get size () {
      return this._list.length;
    }

    append (name, value) {
      this._list.push([String(name), String(value)]);
      this._update();
    }

    delete (name, value) {
      this._list = this._list.filter((pair) => !(pair[0] === String(name) && (value === undefined || pair[1] === String(value))));
      this._update();
    }

    get (name) {
      const found = this._list.find((pair) => pair[0] === String(name));
      return found ? found[1] : null;
    }

    getAll (name) {
      return this._list.filter((pair) => pair[0] === String(name)).map((pair) => pair[1]);
    }

    has (name, value) {
      return this._list.some((pair) => pair[0] === String(name) && (value === undefined || pair[1] === String(value)));
    }

    set (name, value) {
      const key = String(name);
      let found = false;
      this._list = this._list.filter((pair) => {
        if (pair[0] !== key) {
          return true;
        }
        if (found) {
          return false;
        }
        found = true;
        pair[1] = String(value);
        return true;
      });
      if (!found) {
        this._list.push([key, String(value)]);
      }
      this._update();
    }

    sort () {
      this._list = this._list.map((pair, index) => [pair, index]).sort((a, b) => (a[0][0] < b[0][0] ? -1 : a[0][0] > b[0][0] ? 1 : a[1] - b[1])).map((entry) => entry[0]);
      this._update();
    }

    forEach (callback, thisArg) {
      this._list.slice().forEach((pair) => callback.call(thisArg, pair[1], pair[0], this));
    }

    keys () {
      return this._list.map((pair) => pair[0])[Symbol.iterator]();
    }

    values () {
      return this._list.map((pair) => pair[1])[Symbol.iterator]();
    }

    entries () {
      return this._list.map((pair) => pair.slice())[Symbol.iterator]();
    }

    [Symbol.iterator] () {
      return this.entries();
    }

    toString () {
      return this._list.map((pair) => searchParamsEncode(pair[0]) + '=' + searchParamsEncode(pair[1])).join('&');
    }
  }

  const C0_SPACE = /^[\u0000- ]+|[\u0000- ]+$/g;
  const TAB_NEWLINE = /[\t\n\r]/g;

  function percentEncode (text, encodeSet) {
    let result = '';
    for (const character of text) {
      const code = character.codePointAt(0);
      if (code < 0x20 || code > 0x7E || encodeSet.indexOf(character) !== -1) {
        new Uint8Array(host.encodeUtf8(character)).forEach((byte) => {
          result += '%' + byte.toString(16).toUpperCase().padStart(2, '0');
        });
      }
      else {
        result += character;
      }
    }
    return result;
  }

  const FRAGMENT_SET = ' "<>`';
  const QUERY_SET = ' "#<>';
  const SPECIAL_QUERY_SET = ' "#<>\'';
  const PATH_SET = ' "#<>?`{}';
  const USERINFO_SET = ' "#<>?`{}/:;=@[\\]^|';

  function parseHost (input, special) {
    if (input === '') {
      return special ? null : '';
    }
    if (input.charAt(0) === '[') {
      return input.charAt(input.length - 1) === ']' ? input.toLowerCase() : null;
    }
    if (!special) {
      return percentEncode(input, ' "<>`#?{}');
    }
    let decoded;
    try {
      decoded = decodeURIComponent(input);
    }
    catch (error) {
      decoded = input;
    }
    let ascii = decoded.toLowerCase();
    if (/[^\x00-\x7F]/.test(ascii)) {
      ascii = host.domainToASCII(ascii);
      if (ascii === null) {
        return null;
      }
    }
    if (/[\u0000\t\n\r #/:<>?@[\\\]^|%]/.test(ascii)) {
      return null;
    }
    const ipv4 = parseIPv4(ascii);
    return ipv4 === undefined ? ascii : ipv4;
  }

  function parseIPv4 (text) {
    const parts = text.split('.');
    if (parts[parts.length - 1] === '' && parts.length > 1) {
      parts.pop();
    }
    if (parts.length > 4 || !parts.every((part) => /^(0x[0-9a-f]*|[0-9]+)$/i.test(part))) {
      return undefined;
    }
    const numbers = parts.map((part) => (/^0x/i.test(part) ? parseInt(part.slice(2) || '0', 16) : (part.length > 1 && part.charAt(0) === '0' ? parseInt(part, 8) : parseInt(part, 10))));
    let address = numbers[numbers.length - 1];
    for (let index = 0; index < numbers.length - 1; index++) {
      address += numbers[index] * Math.pow(256, 3 - index);
    }
    const octets = [];
    for (let index = 0; index < 4; index++) {
      octets.unshift(address % 256);
      address = Math.floor(address / 256);
    }
    return octets.join('.');
  }

  function normalizePath (segments, special) {
    const output = [];
    segments.forEach((segment, index) => {
      const lower = segment.toLowerCase();
      const isLast = index === segments.length - 1;
      if (lower === '..' || lower === '.%2e' || lower === '%2e.' || lower === '%2e%2e') {
        if (output.length) {
          output.pop();
        }
        if (isLast) {
          output.push('');
        }
        return;
      }
      if (lower === '.' || lower === '%2e') {
        if (isLast) {
          output.push('');
        }
        return;
      }
      output.push(percentEncode(segment, PATH_SET));
    });
    return output;
  }

  function parseURL (input, base) {
    let text = String(input).replace(C0_SPACE, '').replace(TAB_NEWLINE, '');
    const result = { scheme: '', username: '', password: '', host: null, port: null, path: [], opaque: null, query: null, fragment: null };
    const schemeMatch = /^([a-zA-Z][a-zA-Z0-9+.-]*):/.exec(text);
    let hashIndex = text.indexOf('#');
    if (hashIndex !== -1) {
      result.fragment = percentEncode(text.slice(hashIndex + 1), FRAGMENT_SET);
      text = text.slice(0, hashIndex);
    }
    if (!schemeMatch) {
      if (!base) {
        return null;
      }
      return resolveRelative(text, result, base);
    }
    result.scheme = schemeMatch[1].toLowerCase() + ':';
    let rest = text.slice(schemeMatch[0].length);
    const special = Object.prototype.hasOwnProperty.call(SPECIAL_PORTS, result.scheme);
    if (special && base && base.scheme === result.scheme && !/^[/\\]/.test(rest)) {
      return resolveRelative(rest, result, base);
    }
    const queryIndex = rest.indexOf('?');
    if (queryIndex !== -1) {
      result.query = percentEncode(rest.slice(queryIndex + 1), special ? SPECIAL_QUERY_SET : QUERY_SET);
      rest = rest.slice(0, queryIndex);
    }
    if (special) {
      rest = rest.replace(/\\/g, '/');
      rest = rest.replace(/^\/+/, '');
      return parseAuthorityAndPath(rest, result, true);
    }
    if (rest.indexOf('//') === 0) {
      return parseAuthorityAndPath(rest.slice(2), result, false);
    }
    if (rest.charAt(0) === '/') {
      result.path = normalizePath(rest.slice(1).split('/'), false);
      return result;
    }
    result.opaque = percentEncode(rest, '');
    return result;
  }

  function parseAuthorityAndPath (text, result, special) {
    const pathIndex = text.search(special ? /[/]/ : /\//);
    let authority = pathIndex === -1 ? text : text.slice(0, pathIndex);
    const path = pathIndex === -1 ? '' : text.slice(pathIndex + 1);
    const atIndex = authority.lastIndexOf('@');
    if (atIndex !== -1) {
      const userinfo = authority.slice(0, atIndex);
      authority = authority.slice(atIndex + 1);
      const colon = userinfo.indexOf(':');
      result.username = percentEncode(colon === -1 ? userinfo : userinfo.slice(0, colon), USERINFO_SET);
      result.password = colon === -1 ? '' : percentEncode(userinfo.slice(colon + 1), USERINFO_SET);
    }
    const portMatch = /:(\d*)$/.exec(authority);
    let hostText = authority;
    if (portMatch && authority.charAt(authority.length - 1) !== ']') {
      hostText = authority.slice(0, portMatch.index);
      if (portMatch[1] !== '') {
        const port = parseInt(portMatch[1], 10);
        if (port > 65535) {
          return null;
        }
        result.port = String(port) === SPECIAL_PORTS[result.scheme] ? null : String(port);
      }
    }
    else if (/:/.test(authority) && authority.charAt(0) !== '[') {
      return null;
    }
    const parsedHost = parseHost(hostText, special && result.scheme !== 'file:');
    if (parsedHost === null) {
      return null;
    }
    result.host = parsedHost;
    result.path = special || pathIndex !== -1 ? normalizePath(path.split('/'), special) : [];
    return result;
  }

  function resolveRelative (text, result, base) {
    const special = Object.prototype.hasOwnProperty.call(SPECIAL_PORTS, base.scheme);
    result.scheme = base.scheme;
    if (base.opaque !== null) {
      return text === '' ? Object.assign(result, { opaque: base.opaque, query: base.query }) : null;
    }
    let rest = special ? text.replace(/\\/g, '/') : text;
    if (rest.indexOf('//') === 0) {
      const queryIndex = rest.indexOf('?');
      if (queryIndex !== -1) {
        result.query = percentEncode(rest.slice(queryIndex + 1), special ? SPECIAL_QUERY_SET : QUERY_SET);
        rest = rest.slice(0, queryIndex);
      }
      return parseAuthorityAndPath(rest.slice(2), result, special);
    }
    result.username = base.username;
    result.password = base.password;
    result.host = base.host;
    result.port = base.port;
    const queryIndex = rest.indexOf('?');
    if (queryIndex !== -1) {
      result.query = percentEncode(rest.slice(queryIndex + 1), special ? SPECIAL_QUERY_SET : QUERY_SET);
      rest = rest.slice(0, queryIndex);
    }
    if (rest === '') {
      result.path = base.path.slice();
      if (queryIndex === -1) {
        result.query = base.query;
      }
      return result;
    }
    if (rest.charAt(0) === '/') {
      result.path = normalizePath(rest.slice(1).split('/'), special);
      return result;
    }
    const merged = base.path.slice(0, -1).concat(rest.split('/'));
    result.path = normalizePath(merged, special);
    return result;
  }

  function serializePath (record) {
    if (record.opaque !== null) {
      return record.opaque;
    }
    return record.path.length || record.host !== null ? '/' + record.path.join('/') : '';
  }

  class URL {
    constructor (url, base) {
      let baseRecord = null;
      if (base !== undefined) {
        baseRecord = parseURL(String(base), null);
        if (!baseRecord) {
          throw new TypeError('Invalid base URL');
        }
      }
      const record = parseURL(String(url), baseRecord);
      if (!record) {
        throw new TypeError('Invalid URL');
      }
      setHidden(this, '_record', record);
      setHidden(this, '_params', new URLSearchParams(record.query || ''));
      const self = this;
      this._params._url = {
        set _query (value) {
          self._record.query = value;
        }
      };
    }

    static canParse (url, base) {
      try {
        new URL(url, base);
        return true;
      }
      catch (error) {
        return false;
      }
    }

    get href () {
      const record = this._record;
      let output = record.scheme;
      if (record.host !== null) {
        output += '//';
        if (record.username || record.password) {
          output += record.username + (record.password ? ':' + record.password : '') + '@';
        }
        output += record.host + (record.port !== null ? ':' + record.port : '');
      }
      output += serializePath(record);
      if (record.query !== null) {
        output += '?' + record.query;
      }
      if (record.fragment !== null) {
        output += '#' + record.fragment;
      }
      return output;
    }

    set href (value) {
      const record = parseURL(String(value), null);
      if (!record) {
        throw new TypeError('Invalid URL');
      }
      this._record = record;
      this._params._parse(record.query || '');
    }

    get origin () {
      const record = this._record;
      if (['http:', 'https:', 'ws:', 'wss:', 'ftp:'].includes(record.scheme)) {
        return record.scheme + '//' + record.host + (record.port !== null ? ':' + record.port : '');
      }
      return 'null';
    }

    get protocol () {
      return this._record.scheme;
    }

    get username () {
      return this._record.username;
    }

    set username (value) {
      this._record.username = percentEncode(String(value), USERINFO_SET);
    }

    get password () {
      return this._record.password;
    }

    set password (value) {
      this._record.password = percentEncode(String(value), USERINFO_SET);
    }

    get host () {
      const record = this._record;
      return record.host === null ? '' : record.host + (record.port !== null ? ':' + record.port : '');
    }

    get hostname () {
      return this._record.host === null ? '' : this._record.host;
    }

    set hostname (value) {
      const parsed = parseHost(String(value), true);
      if (parsed !== null) {
        this._record.host = parsed;
      }
    }

    get port () {
      return this._record.port === null ? '' : this._record.port;
    }

    set port (value) {
      const text = String(value);
      if (text === '') {
        this._record.port = null;
        return;
      }
      const match = /^\d+/.exec(text);
      if (match && parseInt(match[0], 10) <= 65535) {
        const port = String(parseInt(match[0], 10));
        this._record.port = port === SPECIAL_PORTS[this._record.scheme] ? null : port;
      }
    }

    get pathname () {
      return serializePath(this._record);
    }

    set pathname (value) {
      const text = String(value);
      this._record.path = normalizePath((text.charAt(0) === '/' ? text.slice(1) : text).split('/'), true);
    }

    get search () {
      const query = this._record.query;
      return query === null || query === '' ? '' : '?' + query;
    }

    set search (value) {
      const text = String(value);
      const query = text.charAt(0) === '?' ? text.slice(1) : text;
      this._record.query = text === '' ? null : percentEncode(query, SPECIAL_QUERY_SET);
      this._params._parse(query);
    }

    get searchParams () {
      return this._params;
    }

    get hash () {
      const fragment = this._record.fragment;
      return fragment === null || fragment === '' ? '' : '#' + fragment;
    }

    set hash (value) {
      const text = String(value);
      this._record.fragment = text === '' ? null : percentEncode(text.charAt(0) === '#' ? text.slice(1) : text, FRAGMENT_SET);
    }

    toString () {
      return this.href;
    }

    toJSON () {
      return this.href;
    }
  }

  setHidden(Error.prototype, 'toJSON', function () {
    return { type: 'Error', name: this.name, message: this.message };
  });

  return {
    DOMException, Event, EventTarget, AbortSignal, AbortController, TextEncoder, TextDecoder, atob, btoa, Blob,
    structuredClone, queueMicrotask, crypto, URL, URLSearchParams, SharedArrayBuffer, Atomics
  };
})
