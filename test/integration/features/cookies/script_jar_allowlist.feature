#language: en
#utf-8

@all @xen_005 @cookies @script_jar_allowlist
Feature: Scripts use the workspace jar through pm.cookies.jar() for allowlisted domains
  pm.cookies.jar() works on the workspace cookie jar with Postman's API: get, getAll, set, unset
  and clear take a URL, and either a callback (error first) or, without one, return a promise.
  Scripts may only touch the cookies of domains on the workspace's allowlist (managed in the
  Cookies dialog, subdomains included); any other domain fails with Postman's "programmatic
  access is denied" error. Changes made by scripts reach the request and the Cookies dialog.

  Background:
    Given the workspace "Acme" exists
    And the "api" variable is set to the mocked api url
    And the requests to "/login" respond with the header "Set-Cookie" equal to "session=abc; Path=/"
    And I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "POST", "url": "{{api}}/login" } }
      """
    And I call "ConsoleService.Clear"

  @allowlist
  Scenario: The allowlist is kept per workspace, normalised
    When I call "CookieService.SaveAllowlist" with
      """
      { "local_id": "{{Acme.local_id}}", "domains": [ " Example.TEST ", "127.0.0.1", "example.test", "" ] }
      """
    Then the call should succeed
    And the response should contain the field "domains" with the items "example.test, 127.0.0.1"
    When I call "CookieService.GetAllowlist" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "domains" with the items "example.test, 127.0.0.1"
    And the app data file "workspaces/{{Acme.local_id}}/state.json" should contain the field "cookie_allowlist" array length equal to 2

  @callbacks
  Scenario: Every jar method works with callbacks on an allowlisted domain
    Given I call "CookieService.SaveAllowlist" with
      """
      { "local_id": "{{Acme.local_id}}", "domains": [ "127.0.0.1" ] }
      """
    When I call "RequestService.Send" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "tab_id": "tab-1",
        "request": {
          "method": "GET",
          "url": "{{api}}/profile",
          "scripts": {
            "pre_request": "const jar = pm.cookies.jar(); const url = pm.variables.replaceIn('{{api}}/profile');\njar.get(url, 'session', (error, value) => console.log('get', error, value));\njar.set(url, 'theme', 'dark', (error, cookie) => console.log('set', error, cookie.name, cookie.value));\njar.set(url, { name: 'lang', value: 'en', path: '/', httpOnly: true }, (error, cookie) => console.log('set object', error, cookie.name, cookie.httpOnly));\njar.set(url, 'raw=1; Path=/', (error) => console.log('set raw', error));\njar.getAll(url, (error, cookies) => console.log('getAll', error, cookies.count(), cookies.has('theme')));\njar.unset(url, 'raw', (error) => console.log('unset', error));\njar.get(url, 'missing', (error, value) => console.log('get missing', error, value));"
          }
        }
      }
      """
    Then the call should succeed
    And the response should contain the field "script_errors" array length equal to 0
    And the 1 "GET" request to "/profile" should have the header "Cookie" equal to "session=abc; theme=dark; lang=en"
    When I call "ConsoleService.List" with
      """
      { "limit": 100, "filter": "logs" }
      """
    Then the response should contain the field "0.summary" equal to "get null abc"
    And the response should contain the field "1.summary" equal to "set null theme dark"
    And the response should contain the field "2.summary" equal to "set object null lang true"
    And the response should contain the field "3.summary" equal to "set raw null"
    And the response should contain the field "4.summary" equal to "getAll null 4 true"
    And the response should contain the field "5.summary" equal to "unset null"
    And the response should contain the field "6.summary" equal to "get missing null null"
    When I call "CookieService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "0.cookies" array length equal to 3
    And the event "cookies:changed" should have been emitted 2 times

  @promises
  Scenario: Without a callback the jar methods return promises
    Given I call "CookieService.SaveAllowlist" with
      """
      { "local_id": "{{Acme.local_id}}", "domains": [ "127.0.0.1" ] }
      """
    When I call "RequestService.Send" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "tab_id": "tab-1",
        "request": {
          "method": "GET",
          "url": "{{api}}/profile",
          "scripts": {
            "post_response": "const jar = pm.cookies.jar(); const url = pm.request.url.toString();\nconsole.log(await jar.get(url, 'session'));\nawait jar.set(url, 'later', 'yes');\nconst all = await jar.getAll(url);\nconsole.log(all.map((cookie) => cookie.name).sort().join(','));\nawait jar.clear(url);\nconsole.log((await jar.getAll(url)).count());"
          }
        }
      }
      """
    Then the call should succeed
    And the response should contain the field "script_errors" array length equal to 0
    When I call "ConsoleService.List" with
      """
      { "limit": 100, "filter": "logs" }
      """
    Then the response should contain the field "0.summary" equal to "abc"
    And the response should contain the field "1.summary" equal to "later,session"
    And the response should contain the field "2.summary" equal to "0"
    When I call "CookieService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "" array length equal to 0

  @subdomain
  Scenario: An allowlisted domain covers its subdomains
    Given I call "CookieService.SaveAllowlist" with
      """
      { "local_id": "{{Acme.local_id}}", "domains": [ "example.test" ] }
      """
    When I call "RequestService.Send" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "tab_id": "tab-1",
        "request": {
          "method": "GET",
          "url": "{{api}}/profile",
          "scripts": { "pre_request": "await pm.cookies.jar().set('https://api.example.test/v1', 'sub', '1');\nconsole.log(await pm.cookies.jar().get('https://api.example.test/v1/users', 'sub'));" }
        }
      }
      """
    Then the response should contain the field "script_errors" array length equal to 0
    When I call "ConsoleService.List" with
      """
      { "limit": 100, "filter": "logs" }
      """
    Then the response should contain the field "0.summary" equal to "1"
    When I call "CookieService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response field "" should contain an item with "domain" equal to "api.example.test" and "cookies.0.name" equal to "sub"

  @denied
  Scenario: Domains that are not allowlisted are denied with Postman's error
    Given I call "CookieService.SaveAllowlist" with
      """
      { "local_id": "{{Acme.local_id}}", "domains": [ "example.test" ] }
      """
    When I call "RequestService.Send" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "tab_id": "tab-1",
        "request": {
          "method": "GET",
          "url": "{{api}}/profile",
          "scripts": { "pre_request": "const jar = pm.cookies.jar(); const url = pm.variables.replaceIn('{{api}}/profile');\nawait new Promise((done) => jar.get(url, 'session', (error, value) => { console.log(error.message, value); done() }));\nawait new Promise((done) => jar.set('https://notexample.test', 'a', 'b', (error) => { console.log(error.message); done() }));\ntry { await jar.clear(url) } catch (error) { console.log('rejected', error.message) }" }
        }
      }
      """
    Then the call should succeed
    And the response should contain the field "script_errors" array length equal to 0
    And the 1 "GET" request to "/profile" should have the header "Cookie" equal to "session=abc"
    When I call "ConsoleService.List" with
      """
      { "limit": 100, "filter": "logs" }
      """
    Then the response should contain the field "0.summary" equal to
      """
      CookieStore: programmatic access to "127.0.0.1" is denied undefined
      """
    And the response should contain the field "1.summary" equal to
      """
      CookieStore: programmatic access to "notexample.test" is denied
      """
    And the response should contain the field "2.summary" equal to
      """
      rejected CookieStore: programmatic access to "127.0.0.1" is denied
      """

  @errors
  Scenario: Invalid arguments throw Postman's type errors
    Given I call "CookieService.SaveAllowlist" with
      """
      { "local_id": "{{Acme.local_id}}", "domains": [ "127.0.0.1" ] }
      """
    When I call "RequestService.Send" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "tab_id": "tab-1",
        "request": {
          "method": "GET",
          "url": "{{api}}/profile",
          "scripts": { "pre_request": "const jar = pm.cookies.jar();\ntry { await jar.get('', 'a') } catch (error) { console.log(error.name, error.message) }\ntry { await jar.get('http://127.0.0.1', 1) } catch (error) { console.log(error.message) }\ntry { jar.get('http://127.0.0.1', 'a', 'nope') } catch (error) { console.log(error.message) }\ntry { await jar.set('http://127.0.0.1', 1) } catch (error) { console.log(error.message) }\ntry { await jar.unset('http://127.0.0.1') } catch (error) { console.log(error.message) }\ntry { await jar.getAll('http://127.0.0.1', 'x') } catch (error) { console.log(error.message) }" }
        }
      }
      """
    Then the response should contain the field "script_errors" array length equal to 0
    When I call "ConsoleService.List" with
      """
      { "limit": 100, "filter": "logs" }
      """
    Then the response should contain the field "0.summary" equal to "TypeError CookieJar.get() requires a valid url"
    And the response should contain the field "1.summary" equal to "CookieJar.get() requires cookie name to be a string"
    And the response should contain the field "2.summary" equal to "callback needs to be a function"
    And the response should contain the field "3.summary" equal to "CookieJar.set() requires a valid set cookie arguments"
    And the response should contain the field "4.summary" equal to "CookieJar.unset() requires cookie name to be a string"
    And the response should contain the field "5.summary" equal to "CookieJar.getAll() requires options to be an object"
