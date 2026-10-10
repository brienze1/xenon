#language: en
#utf-8

@all @xen_003 @import @bruno @bruno_scripts
Feature: Translate Bruno scripts, request variables and assertions to pm.*
  Bruno scripts use bru, req, res, test() and expect(); Xenon scripts use pm.*. Imported scripts are
  rewritten with the published bru→pm table (contracts/bruno-mapping.md §3). Request variables
  become pm.variables.set lines (pre-request values as strings at the head of the pre-request
  script, post-response expressions translated and appended to the post-response script) and
  every assertion becomes a pm.test with the matching chai check (§4). A line that cannot be
  translated is kept, commented out with a "// XENON: untranslated" prefix, and listed in the
  summary with its file and line. Disabled variables and assertions are kept as comments.
  On the realistic script corpus (84 script lines) at least 90% of the lines translate without a
  flag (SC-006): exactly 5 lines are flagged.

  Background:
    Given the workspace "Acme" exists
    And I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/bruno/scripts-collection"] }
      """
    And I save the response field "source_id" as "source"
    And I call "ImportService.Import" with
      """
      { "source_id": "{{source}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}" } ] }
      """

  @translation
  Scenario: Every row of the bru→pm table is translated in pre-request scripts
    Then the call should succeed
    And the workspace file "acme/collections/script-rows/translations.yaml" should contain the field "scripts.preRequest" equal to
      """
      const env = pm.environment.get("env");
      pm.environment.set("env", "dev");
      const v = pm.variables.get("v");
      pm.variables.set("v", 1);
      const rv = pm.variables.get("rv");
      const cv = pm.collectionVariables.get("cv");
      const g = pm.globals.get("g");
      pm.globals.set("g", 2);
      const name = pm.environment.name;
      const text = pm.variables.replaceIn("{{base_url}}/x");
      const url1 = pm.request.url.toString();
      const url2 = pm.request.url.toString();
      const m1 = pm.request.method;
      const m2 = pm.request.method;
      const h = pm.request.headers.get("Accept");
      pm.request.headers.upsert({ key: "X-A", value: "1" });
      const hs1 = pm.request.headers.toObject();
      const hs2 = pm.request.headers.toObject();
      const b1 = JSON.parse(pm.request.body.raw);
      const b2 = JSON.parse(pm.request.body.raw);
      pm.request.body.update(JSON.stringify({ a: 1 }));
      """

  @translation
  Scenario: Response calls are translated and tests are appended to the post-response script
    Then the call should succeed
    And the workspace file "acme/collections/script-rows/translations.yaml" should contain the field "scripts.postResponse" equal to
      """
      const s1 = pm.response.code;
      const s2 = pm.response.code;
      const body1 = pm.response.json();
      const body2 = pm.response.json();
      const ct = pm.response.headers.get("content-type");
      const all1 = pm.response.headers.toObject();
      const all2 = pm.response.headers.toObject();
      const t1 = pm.response.responseTime;
      const t2 = pm.response.responseTime;
      pm.test("ok", function () {
        pm.expect(pm.response.code).to.equal(200);
      });
      """

  @assertions
  Scenario: Every assertion operator becomes a pm.test with the matching chai check
    Then the call should succeed
    And the workspace file "acme/collections/script-rows/assertions.yaml" should contain the field "scripts.postResponse" equal to
      """
      pm.test("res.status eq 200", () => pm.expect(pm.response.code).to.eql(200));
      pm.test("res.status neq 500", () => pm.expect(pm.response.code).to.not.eql(500));
      pm.test("res.body.count gt 1", () => pm.expect(pm.response.json().count).to.be.above(1));
      pm.test("res.body.count gte 1", () => pm.expect(pm.response.json().count).to.be.at.least(1));
      pm.test("res.body.count lt 100", () => pm.expect(pm.response.json().count).to.be.below(100));
      pm.test("res.body.count lte 100", () => pm.expect(pm.response.json().count).to.be.at.most(100));
      pm.test("res.status in 200,201", () => pm.expect(pm.response.code).to.be.oneOf([200, 201]));
      pm.test("res.status notIn 500,502", () => pm.expect(pm.response.code).to.not.be.oneOf([500, 502]));
      pm.test("res.body.name contains Ch", () => pm.expect(pm.response.json().name).to.include("Ch"));
      pm.test("res.body.name notContains xx", () => pm.expect(pm.response.json().name).to.not.include("xx"));
      pm.test("res.body.items length 3", () => pm.expect(pm.response.json().items).to.have.lengthOf(3));
      pm.test("res.body.name matches ^Ch.*r$", () => pm.expect(pm.response.json().name).to.match(/^Ch.*r$/));
      pm.test("res.body.name notMatches ^x", () => pm.expect(pm.response.json().name).to.not.match(/^x/));
      pm.test("res.body.name startsWith Ch.", () => pm.expect(pm.response.json().name).to.match(/^Ch\./));
      pm.test("res.body.name endsWith (r)", () => pm.expect(pm.response.json().name).to.match(/\(r\)$/));
      pm.test("res.body.price between 1, 10", () => pm.expect(pm.response.json().price).to.be.within(1, 10));
      pm.test("res.body.items isEmpty", () => pm.expect(pm.response.json().items).to.be.empty);
      pm.test("res.body.tags isNotEmpty", () => pm.expect(pm.response.json().tags).to.not.be.empty);
      pm.test("res.body.deleted isNull", () => pm.expect(pm.response.json().deleted).to.be.null);
      pm.test("res.body.missing isUndefined", () => pm.expect(pm.response.json().missing).to.be.undefined);
      pm.test("res.body.id isDefined", () => pm.expect(pm.response.json().id).to.not.be.undefined);
      pm.test("res.body.active isTruthy", () => pm.expect(pm.response.json().active).to.be.ok);
      pm.test("res.body.archived isFalsy", () => pm.expect(pm.response.json().archived).to.not.be.ok);
      pm.test("res.body.count isNumber", () => pm.expect(pm.response.json().count).to.be.a("number"));
      pm.test("res.body.name isString", () => pm.expect(pm.response.json().name).to.be.a("string"));
      pm.test("res.body.active isBoolean", () => pm.expect(pm.response.json().active).to.be.a("boolean"));
      pm.test("res.body.items isArray", () => pm.expect(pm.response.json().items).to.be.a("array"));
      pm.test("res.body isJson", () => pm.expect(pm.response.json()).to.be.an("object"));
      pm.test("res.body.mode eq \"fast\"", () => pm.expect(pm.response.json().mode).to.eql("fast"));
      // pm.test("res.body.skip eq 1", () => pm.expect(pm.response.json().skip).to.eql(1));
      // XENON: untranslated res.body.size: isHuge
      """
    And the response field "warnings" should contain an item with "code" equal to "untranslated_script" and "location" equal to "Assertions.bru:44"

  @vars
  Scenario: Request variables become pm.variables.set lines
    Then the call should succeed
    And the workspace file "acme/collections/script-rows/vars.yaml" should contain the field "scripts.preRequest" equal to
      """
      pm.variables.set("host", "{{base_url}}");
      pm.variables.set("count", "3");
      // pm.variables.set("off", "x");
      """
    And the workspace file "acme/collections/script-rows/vars.yaml" should contain the field "scripts.postResponse" equal to
      """
      pm.variables.set("id", pm.response.json().id);
      pm.variables.set("code", pm.response.code);
      pm.variables.set("token", pm.response.headers.toObject()["x-token"]);
      pm.variables.set("name", pm.environment.get("name"));
      // XENON: untranslated pm.variables.set("raw", res("data.id"));
      """
    And the response field "warnings" should contain an item with "code" equal to "untranslated_script" and "location" equal to "Vars.bru:24"

  @untranslated
  Scenario: Lines without a pm equivalent are commented out and listed with file and line
    Then the call should succeed
    And the workspace file "acme/collections/script-rows/untranslated.yaml" should contain the field "scripts.postResponse" equal to
      """
      // XENON: untranslated bru.setNextRequest("Next");
      // XENON: untranslated bru.runner.skipRequest();
      // XENON: untranslated bru.sleep(500);
      // XENON: untranslated bru.sendRequest({ url: "https://x.test" }, function (err, response) {});
      // XENON: untranslated bru.cookies.jar().clear();
      // XENON: untranslated req.setUrl("https://other.test");
      // XENON: untranslated req.setMethod("POST");
      // XENON: untranslated req.setTimeout(1000);
      // XENON: untranslated req.setMaxRedirects(2);
      // XENON: untranslated res.setBody({});
      // XENON: untranslated const _ = require("lodash");
      // XENON: untranslated const p = bru.getProcessEnv("HOME");
      // bru.setVar("x", 1) in a comment stays as it is
      const s = "res.status in a string";
      """
    And the response field "warnings" should contain 14 items with "code" equal to "untranslated_script"
    And the response field "warnings" should contain an item with "code" equal to "untranslated_script" and "path" equal to "Script Rows/Untranslated"
    And the response field "warnings" should contain an item with "code" equal to "untranslated_script" and "message" equal to "This line has no pm.* equivalent and was commented out: bru.runner.skipRequest();"

  @untranslated @locations
  Scenario Outline: Each untranslated line is listed with its file and line - <line>
    Then the call should succeed
    And the response field "warnings" should contain an item with "code" equal to "untranslated_script" and "location" equal to "Untranslated.bru:<line>"

    Examples:
      | line |
      | 14   |
      | 15   |
      | 16   |
      | 17   |
      | 18   |
      | 19   |
      | 20   |
      | 21   |
      | 22   |
      | 23   |
      | 24   |
      | 25   |

  @corpus @rate
  Scenario: At least 90% of the lines of a realistic script corpus translate without a flag
    When I call "ImportService.Detect" with
      """
      { "paths": ["{{resources}}/import/bruno/scripts-corpus"] }
      """
    Then the call should succeed
    And the response field "warnings" should contain 5 items with "code" equal to "untranslated_script"
    And the response field "warnings" should contain an item with "code" equal to "untranslated_script" and "location" equal to "collection.bru:6"
    And the response field "warnings" should contain an item with "code" equal to "untranslated_script" and "location" equal to "Signed.bru:14"
    And the response field "warnings" should contain an item with "code" equal to "untranslated_script" and "location" equal to "Orders/List orders.bru:18"
    And the response field "warnings" should contain an item with "code" equal to "untranslated_script" and "location" equal to "Orders/Create order.bru:30"
    And the response field "warnings" should contain an item with "code" equal to "untranslated_script" and "location" equal to "Orders/Cancel order.bru:15"
    And I save the response field "source_id" as "corpus"
    When I call "ImportService.Import" with
      """
      { "source_id": "{{corpus}}", "items": [ { "key": "0", "include": true, "target_local_id": "{{Acme.local_id}}" } ] }
      """
    Then the call should succeed
    And the workspace file "acme/collections/script-corpus/orders/create-order.yaml" should contain the field "scripts.postResponse" equal to
      """
      if (pm.response.code === 429) {
        const wait = Number(pm.response.headers.get("retry-after") || 1) * 1000;
        // XENON: untranslated await bru.sleep(wait);
      }
      pm.variables.set("lastStatus", pm.response.code);
      console.log(`status ${pm.response.code} in ${pm.response.responseTime} ms`);
      pm.test("order created", function () {
        pm.expect(pm.response.code).to.equal(201);
        const data = pm.response.json();
        pm.expect(data.id).to.be.a("string");
        pm.expect(data.reference).to.equal(pm.variables.get("orderRef"));
        pm.variables.set("orderId", data.id);
      });

      pm.test("location header", function () {
        pm.expect(pm.response.headers.get("location")).to.include(pm.response.json().id);
      });
      """
