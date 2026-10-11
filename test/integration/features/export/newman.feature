#language: en
#utf-8

@newman @xen_006_newman
Feature: Postman runs Xenon's Postman exports and sends the same requests (SC-002)
  The Xenon-created corpus (resources/export/shop: every body mode, auth type, params, path
  variables, headers, collection/folder scripts and headers, disabled scripts, settings) is sent
  from Xenon, then exported as a Postman collection and environment and run with Newman (pinned in
  tools/newman) against the same mock server. Every request Newman sends matches what Xenon sent:
  method, path, query, body and the headers the request, its auth and its scripts set. AWS
  Signature v4 is time-dependent, so only its shape is compared. Run with make newman (CI job
  "newman"); it needs Node.js and is not part of make bdd.

  Background:
    Given the workspace "Acme" exists
    And the "api" variable is set to the mocked api url
    And the resource folder "export/shop" is copied to the workspace folder "acme"
    And the workspace file "acme/environments/mock.yaml" contains
      """
      xenon: 1
      kind: environment
      id: 01JX0000000000000000000301
      name: mock
      variables:
        - key: base_url
          value: "{{api}}"
        - key: api_key
          value: k-456
        - key: admin_password
          value: s3cret
        - key: aws_secret
          value: wJalrXUtnFEMI/K7MDENG+bPxRfiCYEXAMPLEKEY
      """
    And I call "EnvironmentService.SetActive" with
      """
      { "local_id": "{{Acme.local_id}}", "id": "01JX0000000000000000000301" }
      """

  Scenario: Newman sends every exported request the way Xenon sends it
    When I send the saved request "01JX0000000000000000000011" in the workspace "Acme"
    And I send the saved request "01JX0000000000000000000012" in the workspace "Acme"
    And I send the saved request "01JX0000000000000000000013" in the workspace "Acme"
    And I send the saved request "01JX0000000000000000000021" in the workspace "Acme"
    And I send the saved request "01JX0000000000000000000022" in the workspace "Acme"
    And I send the saved request "01JX0000000000000000000023" in the workspace "Acme"
    And I send the saved request "01JX0000000000000000000024" in the workspace "Acme"
    And I send the saved request "01JX0000000000000000000025" in the workspace "Acme"
    And I send the saved request "01JX0000000000000000000031" in the workspace "Acme"
    And I call "ExportService.Export" with
      """
      {
        "local_id": "{{Acme.local_id}}", "item_kind": "collection", "item_id": "01JX0000000000000000000001",
        "format": "postman", "shape": "file", "target_path": "{{temp_root}}/shop.postman_collection.json"
      }
      """
    And I call "ExportService.Export" with
      """
      {
        "local_id": "{{Acme.local_id}}", "item_kind": "environment", "item_id": "01JX0000000000000000000301",
        "format": "postman", "shape": "file", "target_path": "{{temp_root}}/mock.postman_environment.json"
      }
      """
    And I run Newman on the file "shop.postman_collection.json" with the environment file "mock.postman_environment.json" in the workspace folder "acme"
    Then the "GET" request to "/products" should have been received 2 times
    And the 1 and 2 "GET" requests to "/products" should have the same query, body and headers "Accept, Authorization, X-Tenant, X-From-Collection, X-From-Folder"
    And the 2 "GET" request to "/products" should not have the header "X-Trace"
    And the 1 and 2 "GET" requests to "/products/42" should have the same query, body and headers "X-Api-Key, X-Tenant, X-From-Collection, X-From-Folder"
    And the 2 "GET" request to "/products/42" should not have the header "Authorization"
    And the 1 and 2 "POST" requests to "/products" should have the same query, body and headers "Authorization, Content-Type, X-Tenant, X-From-Collection, X-From-Folder"
    And the 1 and 2 "POST" requests to "/invoices" should have the same query, body and headers "X-Tenant, X-Area, X-From-Collection"
    And the 2 "POST" request to "/invoices" should have the form field "kind" equal to "invoice"
    And the 2 "POST" request to "/invoices" should not have the form field "note"
    And the 2 "POST" request to "/invoices" should not have the header "Authorization"
    And the 1 and 2 "POST" requests to "/login" should have the same query, body and headers "X-Tenant, X-Area"
    And the 1 and 2 "POST" requests to "/graphql" should have the same query, body and headers "X-Tenant, X-Area"
    And the 1 and 2 "PUT" requests to "/blobs/7" should have the same query, body and headers "X-Tenant, X-Area"
    And the 1 "GET" request to "/signed" should have the header "Authorization" matching "^AWS4-HMAC-SHA256 Credential=AKIDEXAMPLE/[0-9]{8}/us-east-1/execute-api/aws4_request, SignedHeaders=.+, Signature=[0-9a-f]{64}$"
    And the 2 "GET" request to "/signed" should have the header "Authorization" matching "^AWS4-HMAC-SHA256 Credential=AKIDEXAMPLE/[0-9]{8}/us-east-1/execute-api/aws4_request, SignedHeaders=.+, Signature=[0-9a-f]{64}$"
    And the 1 and 2 "HEAD" requests to "/ping" should have the same query, body and headers "X-Tenant, X-From-Collection"
    And the 2 "HEAD" request to "/ping" should not have the header "X-Disabled-Script"
    And the 2 "HEAD" request to "/ping" should not have the header "Authorization"
