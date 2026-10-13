(function (sdk) {
  'use strict';

  const _ = sdk.util;
  const { Property, PropertyBase, PropertyList, HeaderList, Request } = sdk;
  const COOKIE_ATTRIBUTES = { httponly: 'httpOnly', secure: 'secure', domain: 'domain', path: 'path', 'max-age': 'maxAge', session: 'session', expires: 'expires' };
  const JSONP_LEFT = /^[^{(].*\(/;
  const JSONP_RIGHT = /\)[^}].*$|\)$/;
  const FILE_NAME = /;[ \t]*(?:filename)[ \t]*=[ \t]*("(?:[\x20!\x23-\x5b\x5d-\x7e\x80-\xff]|\\[\x20-\x7e])*"|[!#$%&'*+.0-9A-Z^_`a-z|~-]+)[ \t]*/;
  const ENCODED_FILE_NAME = /;[ \t]*(?:filename\*)[ \t]*=[ \t]*([A-Za-z0-9!#$%&+\-^_`{}~]+)'.*'((?:%[0-9A-Fa-f]{2}|[A-Za-z0-9!#$&+.^_`|~-])+)[ \t]*/;

  let reasons;
  function httpReason (code) {
    if (!reasons) {
      reasons = sdk.host.load('http-reasons');
    }
    return reasons.lookup(code);
  }

  const Cookie = function PostmanCookie (options) {
    PropertyBase.call(this, options);
    if (_.isString(options)) {
      options = Cookie.parse(options);
    }
    if (options) {
      this.update(options);
    }
  }

  _.inherit(Cookie, PropertyBase);

  Object.assign(Cookie.prototype, {
    update (options) {
      _.mergeDefined(this, {
        name: _.choose(options.name, options.key),
        expires: _.isString(options.expires) ? new Date(options.expires) : options.expires,
        maxAge: _.has(options, 'maxAge') ? Number(options.maxAge) : undefined,
        domain: options.domain,
        path: options.path,
        secure: _.has(options, 'secure') ? Boolean(options.secure) : undefined,
        httpOnly: _.has(options, 'httpOnly') ? Boolean(options.httpOnly) : undefined,
        hostOnly: _.has(options, 'hostOnly') ? Boolean(options.hostOnly) : undefined,
        session: _.has(options, 'session') ? Boolean(options.session) : undefined,
        value: options.value ? _.ensureEncoded(options.value) : undefined,
        extensions: options.extensions || undefined
      });
    },
    valueOf () {
      try {
        return decodeURIComponent(this.value);
      }
      catch (error) {
        return this.value;
      }
    },
    toString () {
      let text = Cookie.unparseSingle(this);
      if (this.expires && this.expires instanceof Date) {
        if (!Number.isNaN(this.expires.getTime())) {
          text += '; Expires=' + this.expires.toUTCString();
        }
      }
      else if (this.expires) {
        text += '; Expires=' + this.expires;
      }
      if (this.maxAge && this.maxAge !== Infinity) {
        text += '; Max-Age=' + this.maxAge;
      }
      if (this.domain && !this.hostOnly) {
        text += '; Domain=' + this.domain;
      }
      if (this.path) {
        text += '; Path=' + this.path;
      }
      if (this.secure) {
        text += '; Secure';
      }
      if (this.httpOnly) {
        text += '; HttpOnly';
      }
      if (this.extensions) {
        this.extensions.forEach(({ key, value }) => {
          text += '; ' + key + (value === true ? '' : '=' + value);
        });
      }
      return text;
    }
  });

  Object.assign(Cookie, {
    _postman_propertyName: 'Cookie',
    _postman_propertyIndexKey: 'name',
    _postman_propertyIndexCaseInsensitive: true,
    _postman_propertyAllowsMultipleValues: true,
    isCookie (object) {
      return Boolean(object) && (object instanceof Cookie || _.inSuperChain(object.constructor, '_postman_propertyName', 'Cookie'));
    },
    unparse (cookies) {
      if (!Array.isArray(cookies) && !PropertyList.isPropertyList(cookies)) {
        return '';
      }
      return cookies.map(Cookie.unparseSingle).join('; ');
    },
    unparseSingle (cookie) {
      if (!_.isObject(cookie)) {
        return '';
      }
      const value = _.isNil(cookie.value) ? '' : cookie.value;
      return cookie.name ? cookie.name + '=' + value : value;
    },
    parse (text) {
      if (!_.isString(text)) {
        return text;
      }
      const cookie = {};
      const pairs = text.split(/; */);
      const first = Cookie.splitParam(pairs.shift());
      cookie.key = first.key;
      cookie.value = first.value;
      pairs.forEach((pair) => {
        const parameter = Cookie.splitParam(pair);
        const attribute = COOKIE_ATTRIBUTES[parameter.key.toLowerCase()];
        if (attribute) {
          cookie[attribute] = parameter.value;
          return;
        }
        cookie.extensions = cookie.extensions || [];
        cookie.extensions.push(parameter);
      });
      if (!cookie.domain) {
        cookie.hostOnly = true;
      }
      return cookie;
    },
    stringify (cookie) {
      return Cookie.prototype.toString.call(cookie);
    },
    splitParam (parameter) {
      const split = parameter.split('=');
      const key = split[0].trim();
      let value = _.isString(split[1]) ? split[1].trim() : true;
      if (_.isString(value) && value[0] === '"') {
        value = value.slice(1, -1);
      }
      return { key, value };
    }
  });

  function CookieList (parent, cookies) {
    PropertyList.call(this, Cookie, parent, cookies);
  }

  _.inherit(CookieList, PropertyList);

  Object.assign(CookieList, {
    _postman_propertyName: 'CookieList',
    isCookieList (object) {
      return Boolean(object) && (object instanceof CookieList || _.inSuperChain(object.constructor, '_postman_propertyName', 'CookieList'));
    }
  });

  function parseJSON (text, reviver, strict) {
    return sdk.host.load('liquid-json').parse(text, reviver, strict);
  }

  function extensionOf (type) {
    const match = /^\s*([^;\s]*)/.exec(type);
    return (match && match[1] && sdk.host.load('mime')._extensions[match[1].toLowerCase()]) || '';
  }

  function mimeInfo (contentType) {
    const type = contentType || 'text/plain';
    const normalized = sdk.host.load('mime-format').lookup(type);
    return {
      contentType: normalized.source,
      mimeType: normalized.type,
      mimeFormat: normalized.format,
      charset: normalized.charset || 'utf8',
      extension: extensionOf(normalized.source)
    };
  }

  function fileNameOf (disposition) {
    if (!disposition) {
      return undefined;
    }
    let fileName;
    const encoded = ENCODED_FILE_NAME.exec(disposition);
    if (encoded) {
      const decoded = encoded[2].replace(/%([0-9A-Fa-f]{2})/g, (match, hex) => String.fromCharCode(parseInt(hex, 16)));
      const charset = encoded[1].toLowerCase();
      if (charset === 'iso-8859-1') {
        fileName = decoded.replace(/[^\x20-\x7e\xa0-\xff]/g, '?');
      }
      else if (charset === 'utf-8') {
        fileName = sdk.host.decodeBinary(decoded);
      }
    }
    if (!fileName) {
      const plain = FILE_NAME.exec(disposition);
      fileName = plain && plain[1];
      if (fileName && fileName[0] === '"') {
        fileName = fileName.substr(1, fileName.length - 2).replace(/\\([ -~])/g, '$1');
      }
    }
    return fileName;
  }

  const Response = function PostmanResponse (options) {
    Property.apply(this, arguments);
    this.update(options || {});
  }

  _.inherit(Response, Property);

  Object.assign(Response.prototype, {
    _postman_propertyRequiresId: true,
    update (options) {
      const details = Object.assign({}, httpReason(options.code));
      this._details = _.mergeDefined(details, { name: _.choose(options.reason, options.status), code: options.code, standardName: details.name });
      _.mergeDefined(this, {
        originalRequest: options.originalRequest ? new Request(options.originalRequest) : undefined,
        status: this._details.name,
        code: options.code,
        headers: new HeaderList(this, options.header),
        body: options.body,
        stream: options.body && _.isObject(options.body) ? options.body : options.stream,
        cookies: new CookieList(this, options.cookie),
        responseTime: options.responseTime,
        responseSize: options.stream ? options.stream.byteLength : options.responseSize,
        downloadedBytes: options.downloadedBytes
      });
    },
    toJSON () {
      const json = PropertyBase.toJSON(this);
      delete json._details;
      return json;
    },
    reason () {
      return this.status || httpReason(this.code).name;
    },
    details () {
      if (!this._details || this._details.code !== this.code) {
        this._details = Object.assign({}, httpReason(this.code));
        this._details.code = this.code;
        this._details.standardName = this._details.name;
      }
      return Object.assign({}, this._details);
    },
    text () {
      if (this.stream) {
        return _.isString(this.stream) ? this.stream : sdk.host.decodeBytes(this.stream);
      }
      return this.body;
    },
    json (reviver, strict) {
      return parseJSON(this.text(), reviver, strict);
    },
    jsonp (reviver, strict) {
      return parseJSON((this.text() || '').replace(JSONP_LEFT, '').replace(JSONP_RIGHT, ''), reviver, strict);
    },
    contentInfo () {
      const info = mimeInfo(this.headers.get('Content-Type'));
      let fileName = fileNameOf(this.headers.get('Content-Disposition'));
      const extension = info.extension;
      const result = {};
      if (!fileName) {
        fileName = 'response' + (extension ? '.' + extension : '');
      }
      if (info.contentType) {
        result.contentType = info.contentType;
      }
      if (info.mimeType) {
        result.mimeType = info.mimeType;
      }
      if (info.mimeFormat) {
        result.mimeFormat = info.mimeFormat;
      }
      if (info.charset) {
        result.charset = info.charset;
      }
      if (extension) {
        result.fileExtension = extension;
      }
      if (fileName) {
        result.fileName = fileName;
      }
      return result;
    },
    mime () {
      throw new Error('`Response#mime` has been discontinued, use `Response#contentInfo` instead.');
    },
    dataURI () {
      const contentType = this.contentInfo().contentType;
      if (!contentType) {
        return '';
      }
      let encoded = '';
      if (!_.isNil(this.stream)) {
        encoded = sdk.host.base64(_.isString(this.stream) ? this.stream : this.stream);
      }
      else if (!_.isNil(this.body)) {
        encoded = sdk.host.base64(String(this.body));
      }
      return 'data:' + contentType + ';base64, ' + encoded;
    },
    size () {
      const size = { body: 0, header: 0, total: 0 };
      const contentLength = this.headers.get('Content-Length');
      if (_.isNumeric(this.downloadedBytes)) {
        size.body = this.downloadedBytes;
      }
      else if (contentLength && _.isNumeric(contentLength)) {
        size.body = parseInt(contentLength, 10);
      }
      else if (this.stream && _.isNumeric(this.stream.byteLength)) {
        size.body = this.stream.byteLength;
      }
      else if (!_.isNil(this.body)) {
        size.body = sdk.host.byteLength(this.body.toString());
      }
      size.header = ('HTTP/X.X ' + this.code + ' ' + this.reason() + '\r\n\r\n').length + this.headers.contentSize();
      size.total = (size.body || 0) + size.header;
      return size;
    },
    encoding () {
      let format = this.headers.get('Content-Encoding');
      const body = this.stream || this.body;
      let source;
      if (format) {
        source = 'header';
      }
      else if (body && body[0] === 0x1F && body[1] === 0x8B && body[2] === 0x8) {
        format = 'gzip';
        source = 'body';
      }
      return { format, source };
    }
  });

  Object.assign(Response, {
    _postman_propertyName: 'Response',
    isResponse (object) {
      return Boolean(object) && (object instanceof Response || _.inSuperChain(object.constructor, '_postman_propertyName', 'Response'));
    },
    mimeInfo () {
      throw new Error('`Response.mimeInfo` has been discontinued, use `Response#contentInfo` instead.');
    }
  });

  return { Cookie, CookieList, Response };
})
