# Layout Contract: header names and parameter names

The DOM structure and computed styles below are what the Vitest cases lock (jsdom, real CSS module
injected). Visual results are verified by eye (see [quickstart.md](../quickstart.md)).

## C1 — Runner header list (`RunRowDetail`)

DOM (only when `row.sentRequest.headers.length > 0`), inside the `Request` region:

```html
<dl class="headers" aria-label="Request headers">
  <dt translate="no">Content-Type</dt><dd translate="no">application/json</dd>
  <dt translate="no">User-Agent</dt><dd translate="no">Xenon/0.1.14</dd>
  …
</dl>
```

- No `<table>` inside the Request region.
- `dt`/`dd` are direct children of the `dl`, in sent order, strictly alternating.
- Accessible: `getAllByRole('term')` (Testing Library computes no name for `term`, so match by
  text) and each term's `nextElementSibling`, a `definition`.

Computed styles:

| Element | Property | Value |
|---|---|---|
| `dl.headers` | `display` | `grid` |
| `dl.headers` | `grid-template-columns` (whitespace removed) | `fit-content(max(40%,16ch))minmax(0,1fr)` |
| `dl.headers dt` | `overflow-wrap` | `anywhere` |
| `dl.headers dt` | `min-width` | `0px` or `0` |
| `dl.headers dd` | `overflow-wrap` | `anywhere` |
| `dl.headers dd` | `margin` | `0px` / `0` |

## C2 — Documentation Parameters Name cell (`ApiDocView` `Parameters`)

DOM per parameter row:

```html
<td class="nameCell" translate="no"><span class="parameterName">status</span><span class="required"> required</span></td>
```

Computed styles:

| Element | Property | Value |
|---|---|---|
| `td.nameCell` | `white-space` | `nowrap` |
| `td.nameCell` | `overflow-wrap` | not `anywhere` |
| `span.parameterName` | `display` | `inline-block` |
| `span.parameterName` | `width` | `max-content` |
| `span.parameterName` | `max-width` | `256px` in jsdom (authored `32ch`) |
| `span.parameterName` | `white-space` | `normal` |
| `span.parameterName` | `overflow-wrap` | `anywhere` |

The Example cell keeps `class="code"` (`overflow-wrap: anywhere`).
