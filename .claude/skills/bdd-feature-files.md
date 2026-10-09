---
name: BDD Feature Files (Clean Architecture)
description: This skill should be used when the user asks to "create a feature file", "write BDD scenarios", "add Gherkin tests", "write test scenarios", or mentions anything about .feature files in test/integration/features/.
version: 1.0.0
---

# BDD Feature Files — Golang Clean Architecture

Feature files define test scenarios in Gherkin syntax. Each file tests a specific feature/endpoint and uses reusable step definitions for setup, execution, and assertions.

**Location:** `test/integration/features/`

---

## Core Rules

### Rule 1: One Feature Per File
Each `.feature` file tests one feature/endpoint. File name matches the feature name in kebab-case.

### Rule 2: Tags for Filtering
Every feature and scenario has tags for selective execution:
```gherkin
@all @create_user
Feature: Create user
```

### Rule 3: Background for Common Setup
Use `Background:` for setup shared by all scenarios in the file.

### Rule 4: Scenario Outlines for Variations
Use `Scenario Outline:` with `Examples:` for testing multiple inputs.

### Rule 5: JSON Docstrings for Data
Use triple-quoted JSON docstrings for complex data (DB setup, request payloads, API responses).

---

## Feature File Template

```gherkin
#language: en
#utf-8

@all @create_user
Feature: Create user functionality

  Background:
    Given the tables are empty
    And the "LOG_LEVEL" env var is set to "off"
    And the header is empty
    And the "users" exists
    """
    [
      { "id": 1, "name": "Existing User", "email": "existing@test.com", "status": "active" }
    ]
    """
    And the "EXTERNAL_API_URL" env var is set to the mocked api url
    And the 1 "GET" request to "/api/validate" returns status 200 with the following response
    """
    { "valid": true }
    """

  @success
  Scenario: Successfully creates a user
    Given the header contains the key "Authorization" with "Bearer valid-token"
    When I call "POST" "/v1/users" with the following payload
    """
    {
      "name": "New User",
      "email": "new@test.com"
    }
    """
    Then the status returned should be 201
    And the db should contain 2 objects in the "users" table

  @validation
  Scenario Outline: Validation errors for invalid payload - <field>
    Given the header contains the key "Authorization" with "Bearer valid-token"
    When I call "POST" "/v1/users" with the following payload
    """
    {
      "name": "<name>",
      "email": "<email>"
    }
    """
    Then the status returned should be 400
    And the response should contain the field "error.code" equal to "USR-01400"

    Examples:
      | field        | name       | email           |
      | missing name |            | valid@test.com  |
      | missing email| Valid Name |                 |
      | invalid email| Valid Name | not-an-email    |

  @not_found
  Scenario: Returns 404 when user not found
    Given the header contains the key "Authorization" with "Bearer valid-token"
    When I call "GET" "/v1/users/00000000-0000-0000-0000-000000000000" with the following payload
    """
    """
    Then the status returned should be 404
    And the response should contain the field "error.code" equal to "USR-02404"

  @event
  Scenario: Successfully processes event from SQS
    When the following event is received via sqs
    """
    {
      "type": "user-created",
      "payload": {
        "event_type": "USER_CREATED",
        "event_data": { "user_id": 1, "name": "Test" }
      }
    }
    """
    Then the lambda should finish without errors
    And the sqs queue "notification-queue" should have 1 messages published
    And the sqs queue "notification-queue" should have the 1 message published with "user_id" field equal to "1"
```

---

## Gherkin Patterns

### Database Setup
```gherkin
Given the "table_name" exists
"""
[
  { "id": 1, "column1": "value1", "column2": 123 },
  { "id": 2, "column1": "value2", "column2": 456 }
]
"""
```

### Mock External API
```gherkin
Given the 1 "GET" request to "/api/resource/123" returns status 200 with the following response
"""
{ "id": "123", "name": "Test Resource" }
"""
And the 1 "POST" request to "/api/process" returns status 201 with the following response
"""
{ "status": "processed" }
"""
```

### Verify Mock API Was Called Correctly
```gherkin
Then the 1 "POST" request for "/api/process" headers should have the variable "Authorization" with value "Bearer token123"
And the 1 "POST" request for "/api/process" json body should have the variable "user_id" with value "1"
```

### Verify Queue Messages
```gherkin
Then the sqs queue "order-queue" should have 1 messages published
And the sqs queue "order-queue" should have the 1 message published with "order_id" field equal to "123"
And the sqs queue "order-queue" should have the 1 message published with "type" message attribute equal to "ORDER_CREATED"
```

### Verify Database State
```gherkin
Then the db should contain 3 objects in the "orders" table
And the db should contain the "orders" with the id "1" colum "status" equal to "completed"
```

---

- [ ] Follows the comments skill — no inline comments, no unnecessary Go doc comments (see comments.md)
