#language: en
#utf-8

@all @xen_002 @variables @dynamic_variables
Feature: Postman dynamic variables resolve to fresh values when sending
  All 118 of Postman's dynamic variables (test/integration/resources/postman-dynamic-variables.txt,
  vendored from postman-collection) resolve at send time to a value of Postman's format. Every
  occurrence gets its own value, they are looked up after every user scope (a user variable with
  the same name wins), unknown "{{$...}}" names stay literal, and time-based ones follow the app
  clock. The generator is seeded per scenario, so runs are reproducible. Each name is generated
  1,000 times in one request and every value must match its format (SC-005).

  Background:
    Given the workspace "Acme" exists
    And the "api" variable is set to the mocked api url

  @formats
  Scenario Outline: <name> produces 1,000 values of its format
    Given the "payload" variable is set to "{{<name>}}" repeated 1000 times separated by "␞"
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "POST", "url": "{{api}}/echo", "body": { "mode": "raw", "raw": { "language": "text", "content": "{{payload}}" } } } }
      """
    Then the call should succeed
    And the 1 "POST" request to "/echo" should have the body split by "␞" into 1000 parts all matching "<format>"

    Examples:
      | name | format |
      | $guid                        | ^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$ |
      | $timestamp                   | ^\d{10}$ |
      | $isoTimestamp                | ^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{3}Z$ |
      | $randomInt                   | ^(\d{1,3}\|1000)$ |
      | $randomPhoneNumber           | ^\d{3}-\d{3}-\d{4}$ |
      | $randomPhoneNumberExt        | ^\d{2}-\d{3}-\d{3}-\d{4}$ |
      | $randomLocale                | ^[a-z]{2}$ |
      | $randomWords                 | ^[a-z]+( [a-z]+){1,4}$ |
      | $randomFilePath              | ^(/[a-z]+)+/[a-z]+(_[a-z]+)*\.[a-z0-9]+$ |
      | $randomDirectoryPath         | ^(/[a-z]+)+$ |
      | $randomCity                  | ^\p{Lu}[\p{L} .'-]*$ |
      | $randomStreetName            | ^\p{Lu}[\p{L}' -]* [A-Z][a-z]+$ |
      | $randomStreetAddress         | ^\d{3,5} \p{Lu}[\p{L}' -]* [A-Z][a-z]+$ |
      | $randomCountry               | ^\p{Lu}[\p{L} ,.'()&-]+$ |
      | $randomCountryCode           | ^[A-Z]{2}$ |
      | $randomLatitude              | ^-?\d{1,2}\.\d{4}$ |
      | $randomLongitude             | ^-?\d{1,3}\.\d{4}$ |
      | $randomColor                 | ^[a-z]+$ |
      | $randomDepartment            | ^[A-Z][a-z]+$ |
      | $randomProductName           | ^[A-Z][a-z]+ [A-Z][a-z]+ [A-Z][a-z]+$ |
      | $randomProductAdjective      | ^[A-Z][a-z]+$ |
      | $randomProductMaterial       | ^[A-Z][a-z]+$ |
      | $randomProduct               | ^[A-Z][a-z]+$ |
      | $randomCompanyName           | ^\p{Lu}[\p{L}' ,&-]+$ |
      | $randomCompanySuffix         | ^(Inc\|and Sons\|LLC\|Group)$ |
      | $randomCatchPhrase           | ^[A-Z][a-z0-9-]+ [a-z0-9-]+ [a-z0-9-]+$ |
      | $randomBs                    | ^[a-z-]+ [a-z0-9-]+ [a-z0-9-]+$ |
      | $randomCatchPhraseAdjective  | ^[A-Z][a-z-]+$ |
      | $randomCatchPhraseDescriptor | ^[a-z0-9-]+$ |
      | $randomCatchPhraseNoun       | ^[a-z-]+$ |
      | $randomBsAdjective           | ^[a-z0-9-]+$ |
      | $randomBsBuzz                | ^[a-z-]+$ |
      | $randomBsNoun                | ^[a-z-]+$ |
      | $randomDatabaseColumn        | ^[a-zA-Z]+$ |
      | $randomDatabaseType          | ^[a-z]+$ |
      | $randomDatabaseCollation     | ^[a-z0-9_]+$ |
      | $randomDatabaseEngine        | ^[A-Za-z]+$ |
      | $randomDatePast              | ^(Mon\|Tue\|Wed\|Thu\|Fri\|Sat\|Sun) (Jan\|Feb\|Mar\|Apr\|May\|Jun\|Jul\|Aug\|Sep\|Oct\|Nov\|Dec) \d{2} \d{4} \d{2}:\d{2}:\d{2} GMT\+0000 \(Coordinated Universal Time\)$ |
      | $randomDateFuture            | ^(Mon\|Tue\|Wed\|Thu\|Fri\|Sat\|Sun) (Jan\|Feb\|Mar\|Apr\|May\|Jun\|Jul\|Aug\|Sep\|Oct\|Nov\|Dec) \d{2} \d{4} \d{2}:\d{2}:\d{2} GMT\+0000 \(Coordinated Universal Time\)$ |
      | $randomDateRecent            | ^(Mon\|Tue\|Wed\|Thu\|Fri\|Sat\|Sun) (Jan\|Feb\|Mar\|Apr\|May\|Jun\|Jul\|Aug\|Sep\|Oct\|Nov\|Dec) \d{2} \d{4} \d{2}:\d{2}:\d{2} GMT\+0000 \(Coordinated Universal Time\)$ |
      | $randomMonth                 | ^(January\|February\|March\|April\|May\|June\|July\|August\|September\|October\|November\|December)$ |
      | $randomWeekday               | ^(Monday\|Tuesday\|Wednesday\|Thursday\|Friday\|Saturday\|Sunday)$ |
      | $randomBankAccount           | ^\d{8}$ |
      | $randomBankAccountName       | ^[A-Z][a-z]+( [A-Z][a-z]+)* Account$ |
      | $randomCreditCardMask        | ^\d{4}$ |
      | $randomPrice                 | ^\d{1,4}\.\d{2}$ |
      | $randomTransactionType       | ^(deposit\|withdrawal\|payment\|invoice)$ |
      | $randomCurrencyCode          | ^[A-Z]{3}$ |
      | $randomCurrencyName          | ^\p{Lu}[\p{L} ().'-]+$ |
      | $randomCurrencySymbol        | ^\S{1,4}$ |
      | $randomBitcoin               | ^[13][a-km-zA-HJ-NP-Z1-9]{25,34}$ |
      | $randomBankAccountIban       | ^[A-Z]{2}\d{2}[A-Z0-9]{11,30}$ |
      | $randomBankAccountBic        | ^[A-Z]{6}[A-Z0-9]{2}([A-Z0-9]{3})?$ |
      | $randomAbbreviation          | ^[A-Z]{2,5}$ |
      | $randomAdjective             | ^[a-z]+$ |
      | $randomNoun                  | ^[a-z]+$ |
      | $randomVerb                  | ^[a-z]+$ |
      | $randomIngverb               | ^[a-z]+ing$ |
      | $randomPhrase                | ^[A-Z].+[.!?]$ |
      | $randomAvatarImage           | ^https://\S+/avatar/\d+\.jpg$ |
      | $randomImageUrl              | ^http://lorempixel\.com/640/480$ |
      | $randomAbstractImage         | ^http://lorempixel\.com/640/480/abstract$ |
      | $randomAnimalsImage          | ^http://lorempixel\.com/640/480/animals$ |
      | $randomBusinessImage         | ^http://lorempixel\.com/640/480/business$ |
      | $randomCatsImage             | ^http://lorempixel\.com/640/480/cats$ |
      | $randomCityImage             | ^http://lorempixel\.com/640/480/city$ |
      | $randomFoodImage             | ^http://lorempixel\.com/640/480/food$ |
      | $randomNightlifeImage        | ^http://lorempixel\.com/640/480/nightlife$ |
      | $randomFashionImage          | ^http://lorempixel\.com/640/480/fashion$ |
      | $randomPeopleImage           | ^http://lorempixel\.com/640/480/people$ |
      | $randomNatureImage           | ^http://lorempixel\.com/640/480/nature$ |
      | $randomSportsImage           | ^http://lorempixel\.com/640/480/sports$ |
      | $randomTransportImage        | ^http://lorempixel\.com/640/480/transport$ |
      | $randomImageDataUri          | ^data:image/svg\+xml;charset=UTF-8,%3Csvg\S+%3C%2Fsvg%3E$ |
      | $randomEmail                 | ^[A-Za-z0-9._+-]+@[a-z0-9-]+(\.[a-z0-9-]+)*\.[a-z]{2,}$ |
      | $randomExampleEmail          | ^[A-Za-z0-9._+-]+@example\.(com\|org\|net)$ |
      | $randomUserName              | ^[A-Za-z0-9._]+$ |
      | $randomProtocol              | ^https?$ |
      | $randomUrl                   | ^https?://[a-z0-9-]+(\.[a-z0-9-]+)*\.[a-z]{2,}$ |
      | $randomDomainName            | ^[a-z0-9-]+\.[a-z]{2,}$ |
      | $randomDomainSuffix          | ^[a-z]{2,}$ |
      | $randomDomainWord            | ^[a-z0-9-]+$ |
      | $randomIP                    | ^((25[0-5]\|2[0-4]\d\|1\d\d\|[1-9]?\d)\.){3}(25[0-5]\|2[0-4]\d\|1\d\d\|[1-9]?\d)$ |
      | $randomIPV6                  | ^([0-9a-f]{4}:){7}[0-9a-f]{4}$ |
      | $randomUserAgent             | ^(Mozilla\|Opera)/\d+\.\d+ \S.*$ |
      | $randomHexColor              | ^#[0-9a-f]{6}$ |
      | $randomMACAddress            | ^([0-9a-f]{2}:){5}[0-9a-f]{2}$ |
      | $randomPassword              | ^[A-Za-z0-9]{15}$ |
      | $randomLoremWord             | ^[a-z]+$ |
      | $randomLoremWords            | ^[a-z]+ [a-z]+ [a-z]+$ |
      | $randomLoremSentence         | ^[A-Z][a-z]*( [a-z]+)+\.$ |
      | $randomLoremSlug             | ^[a-z]+-[a-z]+-[a-z]+$ |
      | $randomLoremSentences        | ^[A-Z][a-z]*( [a-z]+)+\.( [A-Z][a-z]*( [a-z]+)+\.){1,5}$ |
      | $randomLoremParagraph        | ^[A-Z][a-z]*( [a-z]+)+\.( [A-Z][a-z]*( [a-z]+)+\.){2,}$ |
      | $randomLoremParagraphs       | ^[A-Z][a-z .A-Z]*\.\n[A-Z][a-z .A-Z]*\.\n[A-Z][a-z .A-Z]*\.$ |
      | $randomLoremText             | ^[A-Z][a-z .A-Z]*\.$ |
      | $randomLoremLines            | ^[A-Z][a-z]*( [a-z]+)+\.(\n[A-Z][a-z]*( [a-z]+)+\.){0,4}$ |
      | $randomFirstName             | ^\p{Lu}[\p{L}'-]+$ |
      | $randomLastName              | ^\p{Lu}[\p{L}'-]+$ |
      | $randomFullName              | ^\p{Lu}[\p{L}'-]+ \p{Lu}[\p{L}'-]+$ |
      | $randomJobTitle              | ^[A-Z][a-z]+ [A-Z][a-z]+ [A-Z][a-z]+$ |
      | $randomNamePrefix            | ^(Mr\.\|Mrs\.\|Ms\.\|Miss\|Dr\.)$ |
      | $randomNameSuffix            | ^(Jr\.\|Sr\.\|I\|II\|III\|IV\|V\|MD\|DDS\|PhD\|DVM)$ |
      | $randomJobDescriptor         | ^[A-Z][a-z]+$ |
      | $randomJobArea               | ^[A-Z][a-z]+$ |
      | $randomJobType               | ^[A-Z][a-z]+$ |
      | $randomUUID                  | ^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$ |
      | $randomBoolean               | ^(true\|false)$ |
      | $randomWord                  | ^[a-z]+$ |
      | $randomAlphaNumeric          | ^[a-z0-9]$ |
      | $randomFileName              | ^[a-z]+(_[a-z]+)*\.[a-z0-9]+$ |
      | $randomCommonFileName        | ^[a-z]+(_[a-z]+)*\.(pdf\|mpeg\|wav\|png\|jpeg\|gif\|mp4\|html)$ |
      | $randomMimeType              | ^[a-z]+/[a-z0-9.+-]+$ |
      | $randomCommonFileType        | ^(video\|audio\|image\|text\|application)$ |
      | $randomCommonFileExt         | ^(pdf\|mpeg\|wav\|png\|jpeg\|gif\|mp4\|html)$ |
      | $randomFileType              | ^(application\|audio\|font\|image\|message\|model\|multipart\|text\|video)$ |
      | $randomFileExt               | ^[a-z0-9]+$ |
      | $randomSemver                | ^\d+\.\d+\.\d+$ |

  @fresh
  Scenario: Every occurrence gets its own value
    Given the "payload" variable is set to "{{$guid}}" repeated 50 times separated by " "
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "POST", "url": "{{api}}/echo", "headers": [ { "key": "X-Id", "value": "{{$randomUUID}}", "enabled": true } ], "body": { "mode": "raw", "raw": { "language": "text", "content": "{{payload}}" } } } }
      """
    Then the call should succeed
    And the 1 "POST" request to "/echo" should have the body split by " " into 50 distinct parts
    And the 1 "POST" request to "/echo" should have the header "X-Id" matching "^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$"

  @fresh
  Scenario: Values are generated again for every send
    When I call "RequestService.Send" 2 times cycling through the payloads
      """
      [ { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "POST", "url": "{{api}}/echo", "body": { "mode": "raw", "raw": { "language": "text", "content": "{{$guid}}" } } } } ]
      """
    Then the 1 "POST" request to "/echo" should have the body matching "^[0-9a-f-]{36}$"
    And the 2 "POST" requests to "/echo" should have different bodies

  @json
  Scenario: Dynamic values fill a JSON body, the URL query and headers
    When I call "RequestService.Send" with
      """
      {
        "local_id": "{{Acme.local_id}}",
        "tab_id": "tab-1",
        "request": {
          "method": "POST",
          "url": "{{api}}/echo?n={{$randomInt}}",
          "headers": [ { "key": "X-Request-Id", "value": "{{$guid}}", "enabled": true } ],
          "body": { "mode": "raw", "raw": { "language": "json", "content": "{\"id\": \"{{$guid}}\", \"at\": {{$timestamp}}, \"iso\": \"{{$isoTimestamp}}\", \"n\": {{$randomInt}}, \"email\": \"{{$randomEmail}}\"}" } }
        }
      }
      """
    Then the call should succeed
    And the 1 "POST" request to "/echo" should have the body field "id" matching "^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$"
    And the 1 "POST" request to "/echo" should have the body field "at" matching "^\d{10}$"
    And the 1 "POST" request to "/echo" should have the body field "iso" matching "^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{3}Z$"
    And the 1 "POST" request to "/echo" should have the body field "n" matching "^(\d{1,3}|1000)$"
    And the 1 "POST" request to "/echo" should have the body field "email" matching "^[^@\s]+@[^@\s]+\.[a-z]{2,}$"
    And the 1 "POST" request to "/echo" should have the header "X-Request-Id" matching "^[0-9a-f-]{36}$"
    And the 1 "POST" request to "/echo" should have the query "n" matching "^(\d{1,3}|1000)$"
    When I call "HistoryService.List" with
      """
      { "local_id": "{{Acme.local_id}}" }
      """
    Then the response should contain the field "0.url" equal to "{{api}}/echo?n={{$randomInt}}"

  @clock
  Scenario: Time-based values follow the app clock
    Given the current time is "2026-10-09T12:34:56Z"
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "POST", "url": "{{api}}/echo", "body": { "mode": "raw", "raw": { "language": "text", "content": "{{$timestamp}} {{$isoTimestamp}}" } } } }
      """
    Then the 1 "POST" request to "/echo" should have the body matching "^1791549296 2026-10-09T12:34:56\.\d{3}Z$"

  @precedence
  Scenario Outline: A user variable named <name> in <scope> wins over the dynamic one
    Given I call "VariableService.AddVariable" with
      """
      { "local_id": "{{Acme.local_id}}", "scope": "globals", "key": "<name>", "value": "user-defined", "secret": false }
      """
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "POST", "url": "{{api}}/echo", "body": { "mode": "raw", "raw": { "language": "text", "content": "{{<name>}}" } } } }
      """
    Then the 1 "POST" request to "/echo" should have the body equal to
      """
      user-defined
      """

    Examples:
      | name       | scope   |
      | $guid      | globals |
      | $timestamp | globals |

  @unknown
  Scenario Outline: An unknown name <name> is sent literally
    When I call "RequestService.Send" with
      """
      { "local_id": "{{Acme.local_id}}", "tab_id": "tab-1", "request": { "method": "POST", "url": "{{api}}/echo", "body": { "mode": "raw", "raw": { "language": "text", "content": "{{<name>}}" } } } }
      """
    Then the 1 "POST" request to "/echo" should have the body equal to
      """
      {{<name>}}
      """

    Examples:
      | name         |
      | $notARealOne |
      | $GUID        |
      | $randomint   |
      | $ guid       |
