(function (sdk) {
  'use strict';

  const _ = sdk.util;
  const FIRST_LINE = /(.*?)\n.*/g;
  const statusClassReasons = { info: 1, success: 2, redirection: 3, clientError: 4, serverError: 5 };
  const statusCodeReasons = {
    accepted: { statusCode: 202, name: 'Accepted' },
    withoutContent: { statusCode: 204, name: 'No Content' },
    badRequest: { statusCode: 400, name: 'Bad Request' },
    unauthorised: { statusCode: 401, name: 'Unauthorised' },
    unauthorized: { statusCode: 401, name: 'Unauthorized' },
    forbidden: { statusCode: 403, name: 'Forbidden' },
    notFound: { statusCode: 404, name: 'Not Found' },
    notAcceptable: { statusCode: 406, name: 'Acceptable' },
    rateLimited: { statusCode: 429, name: 'Too Many Requests' }
  };

  function requestOrResponse (what) {
    return (sdk.Request.isRequest(what) && 'request') || (sdk.Response.isResponse(what) && 'response') || undefined;
  }

  function parseErrorOf (response) {
    try {
      response.json();
      return undefined;
    }
    catch (error) {
      if (error.name === 'JSONError' && error.message) {
        return error.message.replace(FIRST_LINE, '$1');
      }
      return error.name + ': ' + error.message;
    }
  }

  return {
    chaiPostman (chai) {
      const Assertion = chai.Assertion;

      function expectStatusCode (assertion) {
        new Assertion(assertion._obj).to.be.postmanResponse;
        new Assertion(assertion._obj).to.have.property('code');
        new Assertion(assertion._obj.code).to.be.a('number');
      }

      function expectReasonAndDetails (assertion) {
        new Assertion(assertion._obj).to.have.property('reason');
        new Assertion(assertion._obj.reason).to.be.a('function');
        new Assertion(assertion._obj).to.have.property('details');
        new Assertion(assertion._obj.details).to.be.a('function');
      }

      Assertion.addProperty('postmanRequest', function () {
        this.assert(sdk.Request.isRequest(this._obj), 'expecting a postman request object but got #{this}', 'not expecting a postman request object');
      });

      Assertion.addProperty('postmanResponse', function () {
        this.assert(sdk.Response.isResponse(this._obj), 'expecting a postman response object but got #{this}', 'not expecting a postman response object');
      });

      Assertion.addProperty('postmanRequestOrResponse', function () {
        this.assert(sdk.Response.isResponse(this._obj) || sdk.Request.isRequest(this._obj),
          'expecting a postman request or response object but got #{this}', 'not expecting a postman request or response object');
      });

      Assertion.addMethod('statusCodeClass', function (classCode) {
        expectStatusCode(this);
        const statusClass = Math.floor(this._obj.code / 100);
        this.assert(statusClass === classCode, 'expected response code to be #{exp}XX but found #{act}', 'expected response code to not be #{exp}XX', classCode, this._obj.code);
      });

      _.forOwn(statusClassReasons, (classCode, reason) => {
        Assertion.addProperty(reason, function () {
          expectStatusCode(this);
          const statusClass = Math.floor(this._obj.code / 100);
          this.assert(statusClass === classCode, 'expected response code to be #{exp}XX but found #{act}', 'expected response code to not be #{exp}XX', classCode, this._obj.code);
        });
      });

      Assertion.addProperty('error', function () {
        expectStatusCode(this);
        const statusClass = Math.floor(this._obj.code / 100);
        this.assert(statusClass === 4 || statusClass === 5, 'expected response code to be 4XX or 5XX but found #{act}', 'expected response code to not be 4XX or 5XX', null, this._obj.code);
      });

      Assertion.addMethod('statusCode', function (code) {
        expectStatusCode(this);
        this.assert(this._obj.code === Number(code), 'expected response to have status code #{exp} but got #{act}', 'expected response to not have status code #{act}', Number(code), this._obj.code);
      });

      Assertion.addMethod('statusReason', function (reason) {
        new Assertion(this._obj).to.be.postmanResponse;
        new Assertion(this._obj).to.have.property('reason');
        new Assertion(this._obj.reason).to.be.a('function');
        reason = String(reason);
        this.assert(this._obj.reason() === reason, 'expected response to have status reason #{exp} but got #{act}', 'expected response to not have status reason #{act}', reason, this._obj.reason());
      });

      _.forOwn(statusCodeReasons, (reason, assertionName) => {
        Assertion.addProperty(assertionName, function () {
          new Assertion(this._obj).to.be.postmanResponse;
          expectReasonAndDetails(this);
          const wantedReason = reason.name.toUpperCase();
          const actualReason = String(this._obj.reason()).toUpperCase();
          const actualStatusCode = Number(this._obj.details().code);
          this.assert(actualReason === wantedReason, 'expected response to have status reason #{exp} but got #{act}', 'expected response to not have status reason #{act}', wantedReason, actualReason);
          this.assert(actualStatusCode === reason.statusCode, 'expected response to have status code #{exp} but got #{act}', 'expected response to not have status code #{act}', reason.statusCode, actualStatusCode);
        });
      });

      Assertion.addProperty('ok', function () {
        if (!sdk.Response.isResponse(this._obj)) {
          this.assert(chai.util.flag(this, 'object'), 'expected #{this} to be truthy', 'expected #{this} to be falsy');
          return;
        }
        expectReasonAndDetails(this);
        const actualReason = String(this._obj.reason()).toUpperCase();
        const actualStatusCode = Number(this._obj.details().code);
        this.assert(actualReason === 'OK', 'expected response to have status reason #{exp} but got #{act}', 'expected response to not have status reason #{act}', 'OK', actualReason);
        this.assert(actualStatusCode === 200, 'expected response to have status code #{exp} but got #{act}', 'expected response to not have status code #{act}', 200, actualStatusCode);
      });

      Assertion.addMethod('status', function (codeOrReason) {
        new Assertion(this._obj).to.be.postmanResponse;
        if (_.isNumber(codeOrReason)) {
          new Assertion(this._obj).to.have.property('code');
          new Assertion(this._obj.code).to.be.a('number');
          this.assert(this._obj.code === Number(codeOrReason), 'expected response to have status code #{exp} but got #{act}', 'expected response to not have status code #{act}', Number(codeOrReason), this._obj.code);
          return;
        }
        new Assertion(this._obj).to.have.property('reason');
        new Assertion(this._obj.reason).to.be.a('function');
        this.assert(this._obj.reason() === codeOrReason, 'expected response to have status reason #{exp} but got #{act}', 'expected response to not have status reason #{act}', codeOrReason, this._obj.reason());
      });

      Assertion.addMethod('header', function (headerKey, headerValue) {
        new Assertion(this._obj).to.be.postmanRequestOrResponse;
        new Assertion(this._obj).to.have.property('headers');
        const kind = requestOrResponse(this._obj);
        this.assert(this._obj.headers.has(headerKey),
          'expected ' + kind + ' to have header with key \'' + String(headerKey) + '\'',
          'expected ' + kind + ' to not have header with key \'' + String(headerKey) + '\'',
          true, this._obj.headers.has(headerKey));
        if (arguments.length < 2) {
          return;
        }
        this.assert(this._obj.headers.one(headerKey).value === headerValue,
          'expected \'' + String(headerKey) + '\' ' + kind + ' header to be #{exp} but got #{act}',
          'expected \'' + String(headerKey) + '\' ' + kind + ' header to not be #{act}',
          headerValue, this._obj.headers.one(headerKey).value);
      });

      Assertion.addProperty('withBody', function () {
        new Assertion(this._obj).to.be.postmanResponse;
        new Assertion(this._obj).to.have.property('text');
        new Assertion(this._obj.text).to.be.a('function');
        const bodyText = this._obj.text();
        this.assert(_.isString(bodyText) && bodyText.length, 'expected response to have content in body', 'expected response to not have content in body');
      });

      Assertion.addProperty('json', function () {
        new Assertion(this._obj).to.be.postmanResponse;
        new Assertion(this._obj).to.have.property('json');
        new Assertion(this._obj.json).to.be.a('function');
        const parseError = parseErrorOf(this._obj);
        this.assert(!parseError, 'expected response body to be a valid json but got error ' + parseError, 'expected response body not to be a valid json');
      });

      Assertion.addMethod('body', function (content) {
        let bodyData;
        new Assertion(this._obj).to.be.postmanResponse;
        if (_.isPlainObject(content)) {
          new Assertion(this._obj).to.have.property('json');
          new Assertion(this._obj.json).to.be.a('function');
          try {
            bodyData = this._obj.json();
          }
          catch (error) {
            bodyData = undefined;
          }
          this.assert(_.isEqual(bodyData, content), 'expected response body json to equal #{exp} but got #{act}', 'expected response body json to not equal #{exp} but got #{act}', content, bodyData);
          return;
        }
        new Assertion(this._obj).to.have.property('text');
        new Assertion(this._obj.text).to.be.a('function');
        bodyData = this._obj.text();
        if (arguments.length === 0) {
          this.assert(_.isString(bodyData) && bodyData.length, 'expected response to have content in body', 'expected response to not have content in body');
          return;
        }
        if (content instanceof RegExp) {
          this.assert(content.exec(bodyData) !== null, 'expected response body text to match #{exp} but got #{act}', 'expected response body text #{act} to not match #{exp}', content, bodyData);
          return;
        }
        this.assert(_.isEqual(bodyData, content), 'expected response body to equal #{exp} but got #{act}', 'expected response body to not equal #{exp}', content, bodyData);
      });

      Assertion.addMethod('jsonBody', function (path, value) {
        new Assertion(this._obj).to.be.postmanResponse;
        new Assertion(this._obj).to.have.property('json');
        new Assertion(this._obj.json).to.be.a('function');
        let jsonBody;
        let parseError;
        try {
          jsonBody = this._obj.json();
        }
        catch (error) {
          parseError = error.name === 'JSONError' && error.message ? error.message.replace(FIRST_LINE, '$1') : error.name + ': ' + error.message;
        }
        if (arguments.length === 0) {
          this.assert(!parseError, 'expected response body to be a valid json but got error ' + parseError, 'expected response body not to be a valid json');
          return;
        }
        if (!_.isString(path)) {
          return;
        }
        this.assert(_.hasPath(jsonBody, path), 'expected #{act} in response to contain property #{exp}', 'expected #{act} in response to not contain property #{exp}', path, jsonBody);
        if (arguments.length > 1) {
          const actual = _.get(jsonBody, path);
          this.assert(_.isEqual(actual, value),
            'expected response body json at "' + path + '" to contain #{exp} but got #{act}',
            'expected response body json at "' + path + '" to not contain #{exp} but got #{act}', value, actual);
        }
      });

      Assertion.addChainableMethod('responseTime', function (value) {
        const time = chai.util.flag(this, 'number');
        this.assert(_.isNumber(time), 'expected response to have a valid response time but got #{act}', 'expected response to not have a valid response time but got #{act}', null, time);
        if (arguments.length) {
          this.assert(_.isEqual(time, value), 'expected response to have a valid response time but got #{act}', 'expected response to not have a valid response time but got #{act}', null, time);
        }
      }, function () {
        new Assertion(this._obj).to.be.postmanResponse;
        new Assertion(this._obj).to.have.property('responseTime');
        chai.util.flag(this, 'number', this._obj.responseTime);
        this._obj = this._obj.responseTime;
      });

      Assertion.addChainableMethod('responseSize', function (value) {
        const size = chai.util.flag(this, 'number');
        this.assert(_.isNumber(size), 'expected response to have a valid response size but got #{act}', 'expected response to not have a valid response size but got #{act}', null, size);
        if (arguments.length) {
          this.assert(_.isEqual(size, value), 'expected response size to equal #{act} but got #{exp}', 'expected response size to not equal #{act} but got #{exp}', value, size);
        }
      }, function () {
        new Assertion(this._obj).to.be.postmanResponse;
        new Assertion(this._obj).to.have.property('size');
        new Assertion(this._obj.size).to.be.a('function');
        const size = this._obj.size();
        const total = (size.header || 0) + (size.body || 0);
        chai.util.flag(this, 'number', total);
        this._obj = total;
      });

      Assertion.addMethod('jsonSchema', function (schema, options) {
        new Assertion(schema).to.be.an('object');
        const settings = Object.assign({ allErrors: true, logger: false }, options);
        let data;
        if (sdk.Response.isResponse(this._obj) || (sdk.Request.isRequest(this._obj) && typeof this._obj.json === 'function')) {
          data = this._obj.json();
        }
        else {
          data = this._obj;
        }
        const Ajv = sdk.host.load('ajv');
        const ajv = new Ajv(settings);
        const valid = ajv.validate(schema, data);
        this.assert(valid && !ajv.errors, 'expected data to satisfy schema but found following errors: \n' + ajv.errorsText(), 'expected data to not satisfy schema', true, valid);
      });
    }
  };
})
