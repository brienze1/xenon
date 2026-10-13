(function (sdk) {
  'use strict';

  const _ = sdk.util;
  const { Property, PropertyBase, PropertyList, VariableList, QueryParam, Url } = sdk;
  const MODES = { file: 'file', formdata: 'formdata', graphql: 'graphql', raw: 'raw', urlencoded: 'urlencoded' };

  const Header = function PostmanHeader (options, name) {
    if (_.isString(options)) {
      options = _.isString(name) ? { key: name, value: options } : Header.parseSingle(options);
    }
    Property.apply(this, arguments);
    this.update(options);
  }

  _.inherit(Header, Property);

  Object.assign(Header.prototype, {
    toString () {
      return this.key + ': ' + this.value;
    },
    valueOf () {
      return this.value;
    },
    update (options) {
      this.key = _.get(options, 'key') || '';
      this.value = _.get(options, 'value', '');
      if (_.has(options, 'system')) {
        this.system = options.system;
      }
      if (_.has(options, 'disabled')) {
        this.disabled = options.disabled;
      }
    }
  });

  Object.assign(Header, {
    _postman_propertyName: 'Header',
    _postman_propertyIndexKey: 'key',
    _postman_propertyIndexCaseInsensitive: true,
    _postman_propertyAllowsMultipleValues: true,
    parse (text) {
      const headers = [];
      const pattern = /^(\S+):(.*)$/gm;
      let match = pattern.exec(text);
      text = text.toString().replace(/\r\n([ \t])/g, '$1');
      while (match) {
        headers.push({ key: match[1], value: match[2].replace(/^\s*(.*\S)?\s*$/, '$1') });
        match = pattern.exec(text);
      }
      return headers;
    },
    parseSingle (header) {
      if (!_.isString(header)) {
        return { key: '', value: '' };
      }
      let index = header.indexOf(':');
      if (index < 0) {
        index = header.length;
      }
      return { key: header.substr(0, index).trim(), value: header.substr(index + 1).trim() };
    },
    unparse (headers, separator) {
      if (!Array.isArray(headers) && !PropertyList.isPropertyList(headers)) {
        return '';
      }
      const end = separator === undefined ? '\r\n' : separator;
      return headers.reduce((text, header) => (header && !header.disabled ? text + Header.unparseSingle(header) + end : text), '');
    },
    unparseSingle (header) {
      if (!_.isObject(header)) {
        return '';
      }
      return header.key + ': ' + header.value;
    },
    isHeader (object) {
      return Boolean(object) && (object instanceof Header || _.inSuperChain(object.constructor, '_postman_propertyName', 'Header'));
    },
    create (value, name) {
      return new Header(value, name);
    }
  });

  function HeaderList (parent, headers) {
    PropertyList.call(this, Header, parent, headers);
  }

  _.inherit(HeaderList, PropertyList);

  Object.assign(HeaderList.prototype, {
    contentSize () {
      if (!this.count()) {
        return 0;
      }
      return Header.unparse(this).length;
    }
  });

  Object.assign(HeaderList, {
    _postman_propertyName: 'HeaderList',
    isHeaderList (object) {
      return Boolean(object) && (object instanceof HeaderList || _.inSuperChain(object.constructor, '_postman_propertyName', 'HeaderList'));
    }
  });

  const FormParam = function PostmanFormParam (options) {
    const source = options || {};
    Property.apply(this, arguments);
    this.key = source.key || '';
    this.value = source.value || '';
    this.type = source.type;
    this.src = source.src;
    this.contentType = source.contentType;
    this.fileName = source.fileName;
  }

  _.inherit(FormParam, Property);

  Object.assign(FormParam.prototype, {
    toString () {
      return this.key + '=' + this.value;
    },
    valueOf () {
      return this.value;
    },
    toJSON () {
      const json = PropertyBase.toJSON(this);
      if (json.type === 'file' && (typeof json.value !== 'string' || !json.value)) {
        delete json.value;
      }
      return json;
    }
  });

  Object.assign(FormParam, {
    _postman_propertyName: 'FormParam',
    _postman_propertyIndexKey: 'key',
    _postman_propertyAllowsMultipleValues: true,
    parse () {
      return undefined;
    }
  });

  const RequestBody = function PostmanRequestBody (options) {
    PropertyBase.apply(this, arguments);
    if (!options) {
      return;
    }
    this.update(options);
  }

  _.inherit(RequestBody, PropertyBase);

  Object.assign(RequestBody.prototype, {
    update (options) {
      if (_.isString(options)) {
        options = { mode: 'raw', raw: options };
      }
      if (!options.mode) {
        return;
      }
      const mode = MODES[options.mode.toString().toLowerCase()] || MODES.raw;
      let urlencoded = options.urlencoded;
      let formdata = options.formdata;
      let graphql = options.graphql;
      let file = options.file;
      let raw = options.raw;
      if (options.urlencoded) {
        if (_.isString(options.urlencoded)) {
          urlencoded = QueryParam.parse(options.urlencoded);
        }
        urlencoded = new PropertyList(QueryParam, this, urlencoded);
      }
      if (options.formdata) {
        formdata = new PropertyList(FormParam, this, options.formdata);
      }
      if (options.graphql) {
        graphql = { query: graphql.query, operationName: graphql.operationName, variables: graphql.variables };
      }
      if (_.isString(options.file)) {
        file = { src: file };
      }
      if (mode === MODES.raw && !raw) {
        raw = '';
      }
      if (mode === MODES.urlencoded && !urlencoded) {
        urlencoded = new PropertyList(QueryParam, this, []);
      }
      if (mode === MODES.formdata && !formdata) {
        formdata = new PropertyList(FormParam, this, []);
      }
      if (mode === MODES.graphql && !graphql) {
        graphql = {};
      }
      Object.assign(this, {
        mode,
        raw,
        urlencoded,
        formdata,
        file,
        graphql,
        options: _.isObject(options.options) ? options.options : undefined,
        disabled: options.disabled
      });
    },
    toString () {
      if (this.mode === MODES.formdata || this.mode === MODES.file) {
        return '';
      }
      if (this.mode === MODES.urlencoded) {
        if (PropertyList.isPropertyList(this.urlencoded)) {
          return QueryParam.unparse(this.urlencoded.all());
        }
        return this.urlencoded && _.isFunction(this.urlencoded.toString) ? this.urlencoded.toString() : '';
      }
      if (this.mode === MODES.raw) {
        return this.raw && _.isFunction(this.raw.toString) ? this.raw.toString() : '';
      }
      return '';
    },
    isEmpty () {
      const data = this.mode && this[this.mode];
      if (!data) {
        return true;
      }
      if (this.mode === MODES.file) {
        return !(data.src || data.content);
      }
      if (_.isString(data)) {
        return data.length === 0;
      }
      if (_.isFunction(data.count)) {
        return data.count() === 0;
      }
      return _.isEmpty(data);
    },
    toJSON () {
      const json = PropertyBase.toJSON(this);
      if (json.file && json.file.content && typeof json.file.content !== 'string') {
        delete json.file.content;
      }
      return json;
    }
  });

  Object.assign(RequestBody, { _postman_propertyName: 'RequestBody', MODES });

  const RequestAuth = function PostmanRequestAuth (options, parent) {
    Property.call(this, options);
    if (parent) {
      this.setParent(parent);
    }
    if (_.has(options, 'type')) {
      this.use(options.type);
    }
    Object.keys(options || {}).forEach((key) => {
      if (key !== 'type') {
        this.update(options[key], key);
      }
    });
  }

  _.inherit(RequestAuth, Property);

  Object.assign(RequestAuth.prototype, {
    update (options, type) {
      if (!_.isObject(options)) {
        return;
      }
      if (!type) {
        type = this.type;
      }
      if (!RequestAuth.isValidType(type)) {
        return;
      }
      let parameters = this[type];
      if (!VariableList.isVariableList(parameters)) {
        parameters = this[type] = new VariableList(this);
        parameters._postman_requestAuthType = type;
      }
      if (Array.isArray(options) || VariableList.isVariableList(options)) {
        parameters.assimilate(options);
        return;
      }
      parameters.syncFromObject(options, false, false);
    },
    use (type, options) {
      if (!RequestAuth.isValidType(type)) {
        return;
      }
      this.type = type;
      let parameters = this[type];
      if (!VariableList.isVariableList(parameters)) {
        parameters = this[type] = new VariableList(this);
      }
      if (Array.isArray(options) || VariableList.isVariableList(options)) {
        parameters.assimilate(options);
        return;
      }
      parameters.syncFromObject(options, false, false);
    },
    current () {
      throw new Error('`Request#current` has been discontinued, use `Request#parameters` instead.');
    },
    parameters () {
      return this[this.type];
    },
    clear (type) {
      if (!(RequestAuth.isValidType(type) && VariableList.isVariableList(this[type]))) {
        return;
      }
      this[type].clear();
      if (type !== this.type) {
        delete this[type];
      }
    }
  });

  Object.assign(RequestAuth, {
    _postman_propertyName: 'RequestAuth',
    isValidType (type) {
      return _.isString(type) && type !== 'type';
    }
  });

  const Request = function PostmanRequest (options) {
    Property.apply(this, arguments);
    if (typeof options === 'string') {
      options = { url: options };
    }
    Object.assign(this, {
      url: new Url(),
      headers: new HeaderList(this, options && options.header),
      method: _.has(options, 'method') && !_.isNil(options.method) ? String(options.method).toUpperCase() : 'GET'
    });
    this.update(options);
  }

  _.inherit(Request, Property);

  Object.assign(Request.prototype, {
    update (options) {
      if (!options) {
        return;
      }
      if (_.has(options, 'url')) {
        this.url.update(options.url);
      }
      if (options.header) {
        this.headers.repopulate(options.header);
      }
      if (_.has(options, 'method')) {
        this.method = _.isNil(options.method) ? 'GET' : String(options.method).toUpperCase();
      }
      _.mergeDefined(this, {
        body: _.has(options, 'body') ? new RequestBody(options.body) : undefined,
        auth: options.auth ? new RequestAuth(options.auth) : undefined
      });
    },
    authorizeUsing (type, options) {
      if (_.isObject(type) && _.isNil(options)) {
        options = Object.assign({}, type);
        delete options.type;
        type = type.type;
      }
      if (type === null) {
        if (_.has(this, 'auth')) {
          delete this.auth;
        }
        return;
      }
      if (!RequestAuth.isValidType(type)) {
        return;
      }
      if (!this.auth) {
        this.auth = new RequestAuth(null, this);
      }
      else {
        this.auth.clear(type);
      }
      this.auth.use(type, options);
    },
    getHeaders (options) {
      options = options || {};
      return this.headers.toObject(options.enabled, !options.ignoreCase, options.multiValue, options.sanitizeKeys);
    },
    forEachHeader (callback) {
      this.headers.all().forEach((header) => callback(header, this));
    },
    addHeader (header) {
      this.headers.add(header);
    },
    removeHeader (toRemove, options) {
      toRemove = _.isString(toRemove) ? toRemove : toRemove.key;
      options = options || {};
      if (!toRemove) {
        return;
      }
      if (options.ignoreCase) {
        toRemove = toRemove.toLowerCase();
      }
      this.headers.remove((header) => (options.ignoreCase ? header.key.toLowerCase() : header.key) === toRemove);
    },
    upsertHeader (header) {
      if (!(header && header.key)) {
        return undefined;
      }
      const existing = this.headers.find({ key: header.key });
      if (!existing) {
        return this.headers.add(header);
      }
      existing.value = header.value;
      return undefined;
    },
    addQueryParams (params) {
      this.url.addQueryParams(params);
    },
    removeQueryParams (params) {
      this.url.removeQueryParams(params);
    },
    size () {
      const contentLength = this.headers.get('Content-Length');
      const size = { body: 0, header: 0, total: 0, source: 'COMPUTED' };
      if (contentLength && _.isNumeric(contentLength)) {
        size.body = parseInt(contentLength, 10);
        size.source = 'CONTENT-LENGTH';
      }
      else if (this.body) {
        size.body = sdk.host.byteLength(this.body.toString());
      }
      size.header = (this.method + ' ' + this.url.getPathWithQuery() + ' HTTP/X.X\r\n\r\n').length + this.headers.contentSize();
      size.total = (size.body || 0) + size.header;
      return size;
    },
    toJSON () {
      const json = PropertyBase.toJSON(this);
      if (Array.isArray(json.header) && !json.header.length) {
        delete json.header;
      }
      return json;
    },
    clone () {
      return new Request(this.toJSON());
    }
  });

  Object.assign(Request, {
    _postman_propertyName: 'Request',
    isRequest (object) {
      return Boolean(object) && (object instanceof Request || _.inSuperChain(object.constructor, '_postman_propertyName', 'Request'));
    }
  });

  return { Header, HeaderList, FormParam, RequestBody, RequestAuth, Request };
})
