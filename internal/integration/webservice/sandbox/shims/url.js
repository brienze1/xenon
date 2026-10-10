(function (sdk) {
  'use strict';

  const _ = sdk.util;
  const { Property, PropertyBase, PropertyList, VariableList, Substitutor } = sdk;
  const VARIABLE_TOKEN = /\{\{[^{}]*}}/g;
  const PLACEHOLDER = /\u0001(\d+)\u0002/g;
  const PATH_VARIABLE = ':';

  function encodeReservedChars (text, encodeEquals) {
    if (!text) {
      return text;
    }
    let result = text;
    if (result.indexOf('&') !== -1) {
      result = result.replace(/&/g, '%26');
    }
    if (result.indexOf('#') !== -1) {
      result = result.replace(/#/g, '%23');
    }
    if (encodeEquals && result.indexOf('=') !== -1) {
      result = result.replace(/=/g, '%3D');
    }
    return result;
  }

  function normalizeParam (text, encodeEquals) {
    if (!(text && typeof text === 'string')) {
      return text;
    }
    if (text.indexOf('&') === -1 && text.indexOf('#') === -1 && !(encodeEquals && text.indexOf('=') !== -1)) {
      return text;
    }
    const pattern = new RegExp(Substitutor.REGEX_EXTRACT_VARS.source, 'g');
    let normalized = '';
    let pointer = 0;
    let match;
    while ((match = pattern.exec(text)) !== null) {
      normalized += encodeReservedChars(text.slice(pointer, match.index), encodeEquals) + match[0];
      pointer = match.index + match[0].length;
    }
    if (pointer < text.length) {
      normalized += encodeReservedChars(text.slice(pointer), encodeEquals);
    }
    return normalized;
  }

  const QueryParam = function PostmanQueryParam (options) {
    Property.apply(this, arguments);
    this.update(options);
  }

  _.inherit(QueryParam, Property);

  Object.assign(QueryParam.prototype, {
    toString () {
      return QueryParam.unparseSingle(this);
    },
    update (param) {
      Object.assign(this, _.isString(param) ? QueryParam.parseSingle(param) : { key: _.get(param, 'key'), value: _.get(param, 'value') });
      if (_.has(param, 'system')) {
        this.system = param.system;
      }
    },
    valueOf () {
      return _.isString(this.value) ? this.value : '';
    }
  });

  Object.assign(QueryParam, {
    _postman_propertyName: 'QueryParam',
    _postman_propertyIndexKey: 'key',
    _postman_propertyAllowsMultipleValues: true,
    parse (query) {
      return _.isString(query) ? query.split('&').map(QueryParam.parseSingle) : [];
    },
    parseSingle (param, index, all) {
      if (param === '' && typeof index === 'number' && Array.isArray(all) && index !== all.length - 1) {
        return { key: null, value: null };
      }
      const position = typeof param === 'string' ? param.indexOf('=') : -1;
      if (position < 0) {
        return { key: param.substr(0, param.length), value: null };
      }
      return { key: param.substr(0, position), value: param.substr(position + 1) };
    },
    unparse (params) {
      if (!params) {
        return '';
      }
      if (!Array.isArray(params) && !PropertyList.isPropertyList(params)) {
        return Object.keys(params).reduce((result, key) => (result ? result + '&' : result) + QueryParam.unparseSingle({ key, value: params[key] }), '');
      }
      let first = true;
      return params.reduce((result, param) => {
        if (param.disabled === true) {
          return result;
        }
        if (first) {
          first = false;
        }
        else {
          result += '&';
        }
        return result + QueryParam.unparseSingle(param);
      }, '');
    },
    unparseSingle (param) {
      if (!param) {
        return '';
      }
      let result = typeof param.key === 'string' ? normalizeParam(param.key, true) : '';
      if (typeof param.value === 'string') {
        result += '=' + normalizeParam(param.value);
      }
      return result;
    }
  });

  function parsePathVariable (segment) {
    if (String(segment).startsWith(PATH_VARIABLE)) {
      const separator = segment.indexOf('.');
      return segment.slice(1, separator === -1 ? undefined : separator) || null;
    }
    return null;
  }

  function parseRaw (raw) {
    const parsed = { raw, protocol: undefined, auth: undefined, host: undefined, port: undefined, path: undefined, query: undefined, hash: undefined };
    if (!(raw && typeof raw === 'string')) {
      return parsed;
    }
    let text = raw.replace(/^[\s﻿\xA0]+/, '');
    parsed.raw = text;
    const tokens = [];
    text = text.replace(VARIABLE_TOKEN, (token) => {
      tokens.push(token);
      return '\u0001' + (tokens.length - 1) + '\u0002';
    });
    const restore = (value) => (tokens.length && typeof value === 'string' ? value.replace(PLACEHOLDER, (match, index) => tokens[Number(index)]) : value);
    let index;
    if ((index = text.indexOf('#')) !== -1) {
      parsed.hash = text.slice(index + 1);
      text = text.slice(0, index);
    }
    if ((index = text.indexOf('?')) !== -1) {
      parsed.query = text.slice(index + 1).split('&');
      text = text.slice(0, index);
    }
    text = text.replace(/\\/g, '/');
    if ((index = text.indexOf('://')) !== -1) {
      parsed.protocol = text.slice(0, index);
      text = text.slice(index + 3).replace(/^\/+/, parsed.protocol.toLowerCase() === 'file' ? '/' : '');
    }
    if ((index = text.indexOf('/')) !== -1) {
      parsed.path = text.slice(index + 1).split('/');
      text = text.slice(0, index);
    }
    if ((index = text.lastIndexOf('@')) !== -1) {
      const auth = text.slice(0, index);
      text = text.slice(index + 1);
      const separator = auth.indexOf(':');
      parsed.auth = separator !== -1 ? [auth.slice(0, separator), auth.slice(separator + 1)] : [auth];
    }
    let port;
    if ((index = text.lastIndexOf(':')) !== -1 && (port = text.slice(index + 1)).indexOf(']') === -1) {
      parsed.port = port;
      text = text.slice(0, index);
    }
    if (text) {
      parsed.host = text.split('.');
    }
    ['protocol', 'hash', 'port'].forEach((key) => { parsed[key] = restore(parsed[key]); });
    ['query', 'path', 'auth', 'host'].forEach((key) => {
      if (parsed[key]) {
        parsed[key] = parsed[key].map(restore);
      }
    });
    return parsed;
  }

  const Url = function PostmanUrl (options) {
    PropertyBase.apply(this, arguments);
    this.update(options);
  }

  _.inherit(Url, PropertyBase);

  Object.assign(Url.prototype, {
    update (url) {
      if (!url) {
        url = '';
      }
      const parsed = _.isString(url) ? Url.parse(url) : url;
      let query = parsed.query;
      let variable = parsed.variable;
      let path = parsed.path;
      let host = parsed.host;
      if (query) {
        if (_.isString(query)) {
          query = QueryParam.parse(query);
        }
        if (!Array.isArray(query) && Object.keys(query).length) {
          query = Object.keys(query).map((key) => ({ key, value: query[key] }));
        }
      }
      if (Array.isArray(variable)) {
        variable = variable.map((item) => {
          if (_.isObject(item)) {
            item.key = item.key || item.id;
          }
          return item;
        });
      }
      if (_.isString(path)) {
        if (path) {
          path = path.replace(/^\/((.+))$/, '$1');
        }
        path = path ? (path === '/' ? [''] : path.split('/')) : undefined;
      }
      if (_.isString(host)) {
        host = host.split(/\.(?![^{]*\}{2})/g);
      }
      Object.assign(this, {
        auth: parsed.auth,
        protocol: parsed.protocol,
        port: parsed.port,
        path,
        hash: parsed.hash,
        host,
        query: new PropertyList(QueryParam, this, query || []),
        variables: new VariableList(this, variable || [])
      });
    },
    addQueryParams (params) {
      this.query.populate(_.isString(params) ? QueryParam.parse(params) : params);
    },
    removeQueryParams (params) {
      const keys = Array.isArray(params) ? params.map((param) => (param.key ? param.key : param)) : [params];
      this.query.remove((param) => keys.includes(param.key));
    },
    getRaw () {
      throw new Error('`Url#getRaw` has been discontinued, use `Url#toString` instead.');
    },
    toString (forceProtocol) {
      let raw = '';
      let protocol = this.protocol;
      if (forceProtocol && !protocol) {
        protocol = 'http://';
      }
      if (protocol) {
        raw += protocol.endsWith('://') ? protocol : protocol + '://';
      }
      if (this.auth) {
        let auth;
        if (typeof this.auth.user === 'string') {
          auth = this.auth.user;
        }
        if (typeof this.auth.password === 'string') {
          auth = (auth || '') + ':' + this.auth.password;
        }
        if (typeof auth === 'string') {
          raw += auth + '@';
        }
      }
      if (this.host) {
        raw += this.getHost();
      }
      if (this.port !== undefined && this.port !== null && typeof this.port.toString === 'function') {
        raw += ':' + this.port.toString();
      }
      if (this.path) {
        raw += this.getPath();
      }
      if (this.query && this.query.count()) {
        let queryString = this.getQueryString();
        if (queryString === '') {
          queryString = this.query.find((param) => !(param && param.disabled)) && '';
        }
        if (typeof queryString === 'string') {
          raw += '?' + queryString;
        }
      }
      if (typeof this.hash === 'string') {
        raw += '#' + this.hash;
      }
      return raw;
    },
    getPath (unresolved) {
      if (unresolved) {
        return '/' + (this.path ? this.path.join('/') : '');
      }
      const segments = (this.path || []).map((segment) => {
        const key = parsePathVariable(segment);
        const value = this.variables.get(key);
        return value && typeof value === 'string' ? value + segment.slice(key.length + 1) : segment;
      });
      return '/' + segments.join('/');
    },
    getQueryString () {
      if (!this.query.count()) {
        return '';
      }
      return QueryParam.unparse(this.query.all());
    },
    getPathWithQuery () {
      const query = this.getQueryString();
      return this.getPath() + (query ? '?' + query : '');
    },
    getHost () {
      if (!this.host) {
        return '';
      }
      return Array.isArray(this.host) ? this.host.join('.') : this.host.toString();
    },
    getRemote (forcePort) {
      const host = this.getHost();
      let port = this.port && this.port.toString();
      if (forcePort && !port) {
        port = this.protocol && this.protocol === 'https' ? '443' : '80';
      }
      return port ? host + ':' + port : host;
    },
    getOAuth1BaseUrl () {
      let protocol = this.protocol || 'http';
      const port = this.port ? this.port.toString() : undefined;
      const host = ((port === '80' || port === '443' || port === undefined) && this.host.join('.')) || (this.host.join('.') + ':' + port);
      protocol = protocol.endsWith('://') ? protocol : protocol + '://';
      return protocol.toLowerCase() + host.toLowerCase() + this.getPath();
    }
  });

  Object.assign(Url, {
    _postman_propertyName: 'Url',
    parse (raw) {
      const url = parseRaw(raw);
      if (url.auth) {
        url.auth = { user: url.auth[0], password: url.auth[1] };
      }
      if (url.query) {
        url.query = url.query.map(QueryParam.parseSingle);
      }
      const seen = {};
      const variables = [];
      (url.path || []).forEach((segment) => {
        const key = parsePathVariable(segment);
        if (key && !seen[key]) {
          seen[key] = true;
          variables.push({ key });
        }
      });
      url.variable = variables.length ? variables : undefined;
      return url;
    },
    isUrl (object) {
      return Boolean(object) && (object instanceof Url || _.inSuperChain(object.constructor, '_postman_propertyName', 'Url'));
    }
  });

  return { QueryParam, Url };
})
