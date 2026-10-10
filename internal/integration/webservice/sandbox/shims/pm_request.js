(function (lists, urls) {
  'use strict';

  const { PropertyList } = lists;
  const { Url, parseQuery } = urls;

  class RequestBody {
    constructor(options) {
      this.update(options);
    }

    update(options) {
      const source = typeof options === 'string' ? { mode: 'raw', raw: options } : options || {};
      this.mode = source.mode || undefined;
      this.raw = source.raw;
      this.urlencoded = new PropertyList('=', source.urlencoded, false);
      this.formdata = new PropertyList('=', source.formdata, false);
      this.graphql = source.graphql;
      this.file = source.file;
    }

    isEmpty() {
      switch (this.mode) {
        case 'raw':
          return !this.raw;
        case 'urlencoded':
          return this.urlencoded.count() === 0;
        case 'formdata':
          return this.formdata.count() === 0;
        case 'graphql':
          return !this.graphql;
        case 'file':
          return !this.file;
        default:
          return true;
      }
    }

    toString() {
      switch (this.mode) {
        case 'raw':
          return this.raw === undefined ? '' : String(this.raw);
        case 'urlencoded':
          return this.urlencoded.toString();
        default:
          return '';
      }
    }

    toJSON() {
      const json = { mode: this.mode };
      if (this.mode === 'raw') {
        json.raw = this.raw === undefined ? '' : String(this.raw);
      }
      if (this.mode === 'urlencoded') {
        json.urlencoded = this.urlencoded.toJSON();
      }
      if (this.mode === 'formdata') {
        json.formdata = this.formdata.toJSON();
      }
      return json;
    }
  }

  class Request {
    constructor(options) {
      const source = typeof options === 'string' ? { url: options } : options || {};
      this.method = source.method ? String(source.method).toUpperCase() : 'GET';
      this.url = source.url instanceof Url ? source.url : new Url(typeof source.url === 'object' && source.url !== null ? source.url.raw : source.url);
      this.headers = new PropertyList(': ', Request.headerItems(source.header), true);
      this.body = new RequestBody(source.body);
    }

    static headerItems(header) {
      if (header === undefined || header === null) {
        return [];
      }
      if (Array.isArray(header) || typeof header === 'string') {
        return header;
      }
      return Object.keys(header).map((key) => ({ key, value: header[key] }));
    }

    addHeader(header) {
      this.headers.add(header);
    }

    upsertHeader(header) {
      this.headers.upsert(header);
    }

    removeHeader(header) {
      this.headers.remove(typeof header === 'object' && header !== null ? header.key : header);
    }

    addQueryParams(params) {
      this.url.addQueryParams(params);
    }

    removeQueryParams(params) {
      this.url.removeQueryParams(params);
    }

    getHeaders(options) {
      return this.headers.toObject(Boolean(options && options.enabled));
    }

    toJSON() {
      return { method: this.method, url: this.url.toString(), header: this.headers.toJSON(), body: this.body.toJSON() };
    }
  }

  return { Request, RequestBody, parseQuery };
})
