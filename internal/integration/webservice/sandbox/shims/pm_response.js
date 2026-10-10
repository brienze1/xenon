(function (lists) {
  'use strict';

  const { PropertyList } = lists;

  class Response {
    constructor(options) {
      const source = options || {};
      this.code = source.code;
      this.status = source.status;
      this.headers = new PropertyList(': ', source.header, true);
      this.responseTime = source.responseTime;
      this.responseSize = source.responseSize;
      Object.defineProperty(this, 'stream', { value: source.body === undefined ? '' : String(source.body), enumerable: false });
    }

    text() {
      return this.stream;
    }

    json() {
      return JSON.parse(this.stream);
    }

    reason() {
      return this.status;
    }

    toJSON() {
      return { code: this.code, status: this.status, header: this.headers.toJSON(), body: this.stream, responseTime: this.responseTime, responseSize: this.responseSize };
    }
  }

  return { Response };
})
