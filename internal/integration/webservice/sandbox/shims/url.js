(function (lists) {
  'use strict';

  const { PropertyList } = lists;

  function parseQuery(text) {
    if (text === undefined || text === null || text === '') {
      return [];
    }
    return String(text).split('&').map((pair) => {
      const index = pair.indexOf('=');
      return index < 0 ? { key: pair, value: null } : { key: pair.slice(0, index), value: pair.slice(index + 1) };
    });
  }

  class Url {
    constructor(raw) {
      this.update(raw);
    }

    update(raw) {
      let text = raw === undefined || raw === null ? '' : String(raw);
      const hashIndex = text.indexOf('#');
      this.hash = hashIndex < 0 ? undefined : text.slice(hashIndex + 1);
      text = hashIndex < 0 ? text : text.slice(0, hashIndex);
      const queryIndex = text.indexOf('?');
      const base = queryIndex < 0 ? text : text.slice(0, queryIndex);
      Object.defineProperty(this, 'base', { value: base, enumerable: false, writable: true, configurable: true });
      this.query = new PropertyList('=', parseQuery(queryIndex < 0 ? '' : text.slice(queryIndex + 1)), false);
      const protocolIndex = base.indexOf('://');
      this.protocol = protocolIndex < 0 ? undefined : base.slice(0, protocolIndex);
      const rest = protocolIndex < 0 ? base : base.slice(protocolIndex + 3);
      const pathIndex = rest.indexOf('/');
      const authority = pathIndex < 0 ? rest : rest.slice(0, pathIndex);
      const portIndex = authority.lastIndexOf(':');
      const hasPort = portIndex > -1 && /^\d+$/.test(authority.slice(portIndex + 1));
      this.host = (hasPort ? authority.slice(0, portIndex) : authority).split('.').filter((part) => part !== '');
      this.port = hasPort ? authority.slice(portIndex + 1) : undefined;
      this.path = pathIndex < 0 ? [] : rest.slice(pathIndex + 1).split('/');
    }

    getHost() {
      return this.host.join('.');
    }

    getPath() {
      return '/' + this.path.join('/');
    }

    getQueryString() {
      return this.query.toString();
    }

    getPathWithQuery() {
      const query = this.getQueryString();
      return this.getPath() + (query ? '?' + query : '');
    }

    getRemote() {
      return this.getHost() + (this.port ? ':' + this.port : '');
    }

    getRaw() {
      return this.toString();
    }

    addQueryParams(params) {
      this.query.populate(typeof params === 'string' ? parseQuery(params) : params);
    }

    removeQueryParams(params) {
      const keys = (Array.isArray(params) ? params : [params]).map((param) => (typeof param === 'object' && param !== null ? param.key : param));
      keys.forEach((key) => this.query.remove(key));
    }

    toString() {
      const query = this.getQueryString();
      return this.base + (query ? '?' + query : '') + (this.hash === undefined ? '' : '#' + this.hash);
    }

    toJSON() {
      return this.toString();
    }
  }

  Url.parse = function (raw) {
    return new Url(raw);
  };

  return { Url, parseQuery };
})
