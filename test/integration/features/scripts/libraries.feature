#language: en
#utf-8

@all @xen_004 @scripts @libraries
Feature: Scripts use the libraries and globals Postman's sandbox ships
  require() loads the libraries of postman-sandbox 6.7.4 at the same versions (ajv, atob, btoa,
  chai, cheerio, crypto-js, csv-parse/lib/sync, lodash, moment, postman-collection, tv4, uuid,
  xml2js, backbone) and the Node-style modules (assert, buffer, events, path, punycode,
  querystring, string_decoder, stream, timers, url, util), and scripts see the same globals
  (atob, btoa, Buffer, URL, URLSearchParams, TextEncoder, TextDecoder, structuredClone,
  queueMicrotask, crypto, AbortController, Event, EventTarget, DOMException, Blob, Intl and the
  timers). Each gives the result postman-sandbox gives. Any other module name fails with
  Postman's "Cannot find module" error. Nothing else of the machine is reachable.

  Background:
    Given the workspace "Acme" exists
    And the "api" variable is set to the mocked api url
    And the "billing" variable is set to "01JK8000000000000000000001"
    And the "create" variable is set to "01JK8000000000000000000004"
    And the workspace file "acme/workspace.yaml" contains
      """
      xenon: 1
      kind: workspace
      id: 01JK8000000000000000000000
      name: Acme
      order:
        - billing
      """
    And the workspace file "acme/collections/billing/collection.yaml" contains
      """
      xenon: 1
      kind: collection
      id: {{billing}}
      name: Billing
      """
    And the 1 "GET" request to "/invoices" returns status 200 with the following response
      """
      {"id":"inv-1"}
      """
    And I call "ConsoleService.Clear"

  @require
  Scenario Outline: require('<module>') gives Postman's result
    Given the workspace file "acme/collections/billing/create.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create
      method: GET
      url: "{{api}}/invoices"
      scripts:
        preRequest: "<script>"
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "name": "Create", "method": "GET", "url": "{{api}}/invoices" } }
      """
    Then the call should succeed
    And the response should contain the field "outcome" equal to "sent"
    And the response should contain the field "script_errors" array length equal to 0
    When I call "ConsoleService.List" with
      """
      { "limit": 100 }
      """
    Then the response should contain the field "" array length equal to 1
    And the response should contain the field "0.summary" equal to "<result>"

    Examples:
      | module             | script                                                                                                                                         | result                                       |
      | ajv                | const Ajv = require('ajv'); const ajv = new Ajv(); console.log(ajv.validate({ type: 'object', required: ['id'] }, { id: 1 }), ajv.validate({ type: 'string' }, 1)) | true false                                   |
      | atob               | console.log(require('atob')('eGVub24='))                                                                                                       | xenon                                        |
      | btoa               | console.log(require('btoa')('xenon'))                                                                                                          | eGVub24=                                     |
      | chai               | const { expect } = require('chai'); try { expect(1).to.equal(2) } catch (e) { console.log(e.message) }                                         | expected 1 to equal 2                        |
      | cheerio            | const $ = require('cheerio').load('<ul><li>a</li><li>b</li></ul>'); console.log($('li').length, $('li').last().text())                         | 2 b                                          |
      | crypto-js          | const CryptoJS = require('crypto-js'); console.log(CryptoJS.HmacSHA256('message', 'key').toString(CryptoJS.enc.Base64))                         | bp7ym3X//Ft6uuUn1Y/a2y/kLnIZARl2kXNDBl9Y7Uo= |
      | csv-parse/lib/sync | const parse = require('csv-parse/lib/sync'); console.log(parse('a,b\\\\n1,2', { columns: true })[0].b)                                            | 2                                            |
      | lodash             | const _ = require('lodash'); console.log(_.VERSION, _.groupBy([1.2, 1.4, 2.1], Math.floor)[1].length)                                           | 4.18.1 2                                     |
      | moment             | const moment = require('moment'); console.log(moment.utc('2024-01-31').add(1, 'month').format('YYYY-MM-DD'), moment.version)                   | 2024-02-29 2.30.1                            |
      | postman-collection | const { Url, Header } = require('postman-collection'); console.log(new Url('https://a.io/p?x=1').getHost(), Header.parse('A: 1\\\\nB: 2').length) | a.io 2                                       |
      | tv4                | const tv4 = require('tv4'); console.log(tv4.validate({ a: 1 }, { type: 'object', properties: { a: { type: 'string' } } }), tv4.error.message)  | false Invalid type: number (expected string) |
      | uuid               | const uuid = require('uuid'); console.log(/^[0-9a-f-]{36}$/.test(uuid.v4()), typeof uuid)                                                      | true function                                |
      | xml2js             | require('xml2js').parseString('<a><b>1</b></a>', (err, result) => console.log(err, result.a.b[0]))                                             | null 1                                       |
      | backbone           | const Backbone = require('backbone'); console.log(new Backbone.Model({ a: 1 }).get('a'), Backbone.VERSION)                                     | 1 1.6.0                                      |
      | assert             | const assert = require('assert'); try { assert.strictEqual(1, 2) } catch (e) { console.log(e.name, e.code) }                                   | AssertionError ERR_ASSERTION                 |
      | buffer             | const { Buffer } = require('buffer'); console.log(Buffer.from('xenon').toString('hex'))                                                        | 78656e6f6e                                   |
      | events             | const EventEmitter = require('events'); const e = new EventEmitter(); e.on('x', (v) => console.log('got', v)); e.emit('x', 1)                   | got 1                                        |
      | path               | const path = require('path'); console.log(path.join('/a', 'b', '../c'), path.extname('f.json'))                                                | /a/c .json                                   |
      | punycode           | console.log(require('punycode').toASCII('mañana.com'))                                                                                         | xn--maana-pta.com                            |
      | querystring        | const qs = require('querystring'); console.log(qs.stringify({ a: 1, b: [2, 3] }), qs.parse('x=1&y=2').y)                                        | a=1&b=2&b=3 2                                |
      | string_decoder     | const { StringDecoder } = require('string_decoder'); console.log(new StringDecoder('utf8').write(Buffer.from([0xe2, 0x82, 0xac])))              | €                                            |
      | stream             | const { Readable } = require('stream'); console.log(typeof Readable, typeof require('stream').Transform)                                        | function function                            |
      | timers             | console.log(typeof require('timers').setTimeout, typeof require('timers').clearInterval)                                                       | function function                            |
      | url                | const url = require('url'); console.log(url.parse('https://a.io:8080/p?q=1').port, url.resolve('https://a.io/a/b', '../c'))                    | 8080 https://a.io/c                          |
      | util               | const util = require('util'); console.log(util.format('%s=%d', 'a', 1), util.inspect({ a: [1] }))                                              | a=1 { a: [ 1 ] }                             |
      | unknown module     | try { require('child_process') } catch (e) { console.log(e.message) }                                                                          | Cannot find module 'child_process'           |

  @globals
  Scenario Outline: The <global> global gives Postman's result
    Given the workspace file "acme/collections/billing/create.yaml" contains
      """
      xenon: 1
      kind: request
      id: {{create}}
      name: Create
      method: GET
      url: "{{api}}/invoices"
      scripts:
        preRequest: "<script>"
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "id": "{{create}}", "name": "Create", "method": "GET", "url": "{{api}}/invoices" } }
      """
    Then the call should succeed
    And the response should contain the field "script_errors" array length equal to 0
    When I call "ConsoleService.List" with
      """
      { "limit": 100 }
      """
    Then the response should contain the field "<index>.summary" equal to "<result>"

    Examples:
      | global          | script                                                                                                                                                       | index | result                                             |
      | atob and btoa   | console.log(btoa('xenon'), atob('eGVub24='))                                                                                                                 | 0     | eGVub24= xenon                                     |
      | Buffer          | console.log(Buffer.from('hi').toString('base64'))                                                                                                            | 0     | aGk=                                               |
      | URL             | const u = new URL('https://a.io/p?x=1&y=2'); u.searchParams.append('z', '3'); console.log(u.host, u.searchParams.get('y'), u.toString())                     | 0     | a.io 2 https://a.io/p?x=1&y=2&z=3                  |
      | URLSearchParams | console.log(new URLSearchParams({ a: '1', b: 'x y' }).toString())                                                                                            | 0     | a=1&b=x+y                                          |
      | TextEncoder     | console.log(new TextEncoder().encode('é').length, new TextDecoder().decode(new Uint8Array([104, 105])))                                                      | 0     | 2 hi                                               |
      | structuredClone | const c = structuredClone({ d: new Date(0), m: new Map([[1, 2]]) }); console.log(c.d.toISOString(), c.m.get(1))                                              | 0     | 1970-01-01T00:00:00.000Z 2                         |
      | queueMicrotask  | queueMicrotask(() => console.log('micro')); console.log('sync')                                                                                              | 1     | micro                                              |
      | crypto          | console.log(crypto.getRandomValues(new Uint8Array(4)).length, /^[0-9a-f-]{36}$/.test(crypto.randomUUID()))                                                   | 0     | 4 true                                             |
      | crypto.subtle   | const d = await crypto.subtle.digest('SHA-256', new TextEncoder().encode('abc')); console.log(Buffer.from(d).toString('hex'))                                 | 0     | ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad |
      | HMAC            | const k = await crypto.subtle.importKey('raw', new TextEncoder().encode('key'), { name: 'HMAC', hash: 'SHA-256' }, false, ['sign', 'verify']); const s = await crypto.subtle.sign('HMAC', k, new TextEncoder().encode('message')); console.log(Buffer.from(s).toString('base64'), await crypto.subtle.verify('HMAC', k, s, new TextEncoder().encode('message'))) | 0 | bp7ym3X//Ft6uuUn1Y/a2y/kLnIZARl2kXNDBl9Y7Uo= true |
      | AbortController | const c = new AbortController(); c.signal.addEventListener('abort', () => console.log('aborted', c.signal.aborted)); c.abort()                                | 0     | aborted true                                       |
      | EventTarget     | const t = new EventTarget(); t.addEventListener('ping', (e) => console.log(e.type)); t.dispatchEvent(new Event('ping'))                                       | 0     | ping                                               |
      | DOMException    | console.log(new DOMException('m', 'AbortError').name, new Blob(['abc']).size)                                                                                | 0     | AbortError 3                                       |
      | Intl            | console.log(new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(1234.5), (1234567.891).toLocaleString('en-US'))                    | 0     | $1,234.50 1,234,567.891                            |
      | Intl dates      | console.log(new Intl.DateTimeFormat('en-US', { timeZone: 'UTC', dateStyle: 'medium' }).format(new Date(0)))                                                  | 0     | Jan 1, 1970                                        |
      | timers          | setTimeout(() => console.log('timeout'), 5); setImmediate(() => console.log('immediate')); const i = setInterval(() => { console.log('interval'); clearInterval(i) }, 1) | 2 | timeout                                  |
      | lazy legacy _   | console.log(typeof _, typeof CryptoJS, typeof tv4, typeof cheerio, typeof Backbone)                                                                          | 0     | function object object function object             |
