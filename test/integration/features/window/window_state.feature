#language: en
#utf-8

@all @xen_001 @window @window_state
Feature: Remember the window size and position
  The window's bounds and maximised state are kept in the local app.json (never in a workspace
  folder) and restored on the next start, fitted to the screens that exist then: a window never
  opens larger than its screen, smaller than the 960 x 600 minimum, or off every screen.

  @defaults
  Scenario: Without saved bounds the window opens at its default size, centred on the primary screen
    When I call "WindowService.Restore" with
      """
      { "screens": [{ "x": 0, "y": 0, "width": 1440, "height": 900, "primary": false }, { "x": 1440, "y": 0, "width": 1920, "height": 1080, "primary": true }] }
      """
    Then the call should succeed
    And the response should contain the field "width" equal to "1280"
    And the response should contain the field "height" equal to "800"
    And the response should contain the field "x" equal to "1760"
    And the response should contain the field "y" equal to "140"
    And the response should contain the field "maximized" equal to "false"
    And the app data file "app.json" should not exist

  @success
  Scenario: Saved bounds are written to app.json and restored on the next start
    Given the workspace "Acme" exists
    When I call "WindowService.Save" with
      """
      { "x": 120, "y": 80, "width": 1500, "height": 900, "maximized": false }
      """
    Then the call should succeed
    And the app data file "app.json" should contain the field "window.x" equal to "120"
    And the app data file "app.json" should contain the field "window.y" equal to "80"
    And the app data file "app.json" should contain the field "window.width" equal to "1500"
    And the app data file "app.json" should contain the field "window.height" equal to "900"
    And the app data file "app.json" should contain the field "window.maximized" equal to "false"
    And the app data file "app.json" should contain the field "workspaces" array length equal to 1
    And the app data path "app.json" should have the permissions "0600"
    Given the workspace "Globex" exists
    When I call "WindowService.Restore" with
      """
      { "screens": [{ "x": 0, "y": 0, "width": 1920, "height": 1080, "primary": true }] }
      """
    Then the call should succeed
    And the response should contain the field "x" equal to "120"
    And the response should contain the field "y" equal to "80"
    And the response should contain the field "width" equal to "1500"
    And the response should contain the field "height" equal to "900"
    And the app data file "app.json" should contain the field "workspaces" array length equal to 2
    And the app data file "app.json" should contain the field "window.width" equal to "1500"

  @maximized
  Scenario: A maximised window keeps the bounds it had before it was maximised
    Given I call "WindowService.Save" with
      """
      { "x": 100, "y": 60, "width": 1300, "height": 850, "maximized": false }
      """
    Then the app data file "app.json" should contain the field "version" equal to "1"
    And the app data file "app.json" should contain the field "workspaces" array length equal to 0
    When I call "WindowService.Save" with
      """
      { "x": 0, "y": 0, "width": 1920, "height": 1080, "maximized": true }
      """
    And I call "WindowService.Restore" with
      """
      { "screens": [{ "x": 0, "y": 0, "width": 1920, "height": 1080, "primary": true }] }
      """
    Then the response should contain the field "maximized" equal to "true"
    And the response should contain the field "x" equal to "100"
    And the response should contain the field "width" equal to "1300"
    When I call "WindowService.Save" with
      """
      { "x": 10, "y": 20, "width": 1000, "height": 700, "maximized": false }
      """
    And I call "WindowService.Restore" with
      """
      { "screens": [{ "x": 0, "y": 0, "width": 1920, "height": 1080, "primary": true }] }
      """
    Then the response should contain the field "maximized" equal to "false"
    And the response should contain the field "x" equal to "10"
    And the response should contain the field "width" equal to "1000"

  @clamp
  Scenario Outline: Restored bounds are fitted to the screens - <case>
    Given I call "WindowService.Save" with
      """
      <saved>
      """
    When I call "WindowService.Restore" with
      """
      { "screens": <screens> }
      """
    Then the call should succeed
    And the response should contain the field "x" equal to "<x>"
    And the response should contain the field "y" equal to "<y>"
    And the response should contain the field "width" equal to "<width>"
    And the response should contain the field "height" equal to "<height>"

    Examples:
      | case                         | saved                                                                     | screens                                                                                                                                         | x    | y   | width | height |
      | larger than the screen       | { "x": 3000, "y": -50, "width": 2600, "height": 1500, "maximized": false } | [{ "x": 0, "y": 0, "width": 1920, "height": 1080, "primary": true }]                                                                                                                                     | 0    | 0   | 1920  | 1080   |
      | below the minimum size       | { "x": 100, "y": 100, "width": 500, "height": 300, "maximized": false }    | [{ "x": 0, "y": 0, "width": 1920, "height": 1080, "primary": true }]                                                                                                                                     | 100  | 100 | 960   | 600    |
      | screen smaller than minimum  | { "x": 0, "y": 0, "width": 1280, "height": 800, "maximized": false }       | [{ "x": 0, "y": 0, "width": 800, "height": 500, "primary": true }]                                                                              | 0    | 0   | 800   | 500    |
      | partly off the right edge    | { "x": 1500, "y": 900, "width": 1000, "height": 700, "maximized": false }  | [{ "x": 0, "y": 0, "width": 1920, "height": 1080, "primary": true }]                                                                                                                                     | 920  | 380 | 1000  | 700    |
      | on a secondary screen        | { "x": 2000, "y": 100, "width": 1200, "height": 800, "maximized": false }  | [{ "x": 0, "y": 0, "width": 1440, "height": 900, "primary": true }, { "x": 1440, "y": 0, "width": 2560, "height": 1440, "primary": false }] | 2000 | 100 | 1200  | 800    |
      | screen left of the primary   | { "x": -1500, "y": 50, "width": 1200, "height": 800, "maximized": false }  | [{ "x": -1920, "y": 0, "width": 1920, "height": 1080, "primary": false }, { "x": 0, "y": 0, "width": 1440, "height": 900, "primary": true }] | -1500 | 50 | 1200  | 800    |
      | screen that is gone          | { "x": 5000, "y": 100, "width": 1200, "height": 800, "maximized": false }  | [{ "x": 0, "y": 0, "width": 1440, "height": 900, "primary": true }]                                                                             | 240  | 100 | 1200  | 800    |
      | no primary screen reported   | { "x": 5000, "y": 100, "width": 1200, "height": 800, "maximized": false }  | [{ "x": 100, "y": 0, "width": 1440, "height": 900, "primary": false }]                                                                          | 340  | 100 | 1200  | 800    |

  @validation
  Scenario Outline: Invalid window requests are rejected - <case>
    When I call "WindowService.<method>" with
      """
      <payload>
      """
    Then the call should fail with code "<code>"

    Examples:
      | case             | method  | payload                                                             | code      |
      | no screens       | Restore | { "screens": [] }                                                   | WIN-00400 |
      | empty screen     | Restore | { "screens": [{ "x": 0, "y": 0, "width": 0, "height": 900 }] }      | WIN-00400 |
      | zero width       | Save    | { "x": 0, "y": 0, "width": 0, "height": 700, "maximized": false }   | WIN-01400 |
      | negative height  | Save    | { "x": 0, "y": 0, "width": 900, "height": -1, "maximized": false }  | WIN-01400 |

  @corrupted
  Scenario: An unreadable app.json is reported so the window opens with its defaults
    Given the app data file "app.json" contains
      """
      { "version": 1, "window": broken
      """
    When I call "WindowService.Restore" with
      """
      { "screens": [{ "x": 0, "y": 0, "width": 1920, "height": 1080, "primary": true }] }
      """
    Then the call should fail with code "WKS-12500"
