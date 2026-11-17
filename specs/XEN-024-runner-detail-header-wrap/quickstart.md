# Quickstart: validating XEN-024

## Prerequisites

- Fresh worktree: `cd frontend && npm ci`, then `make bindings` from the repo root before Vitest.

## Automated

```bash
cd frontend
npx vitest run src/features/runner/components/RunRowDetail.test.tsx \
               src/features/runner/components/RunResults.test.tsx \
               src/features/documentation/components/ApiDocView.test.tsx
cd .. && make validate
```

Expected: the new cases (contract C1 and C2 in [contracts/layout-contract.md](contracts/layout-contract.md))
are red on `origin/main` and green after the change; every pre-existing case stays green.

## By eye (macOS)

Use the scratchpad Vite harness (see the by-eye harness notes: `serve.mjs` with `configFile: false`,
`@` alias → `frontend/src`, `node_modules` symlink, global styles imported) rendering:

1. `RunRowDetail` for a row that sent `Content-Type: application/json`, `User-Agent: Xenon/0.1.14`,
   `Accept: */*`, `Accept-Encoding: gzip, deflate, br`, `Connection: keep-alive`, URL
   `https://api.test/campaign-manager-api/v1/campaigns/510876/briefings/7/guidelines`, a JSON body —
   in a 280 px (default layout) and a 420 px (right panel hidden) wide panel, inside a 1280×800
   page for the 280 px case.
   - Expect (both widths): 5 header lines, every name on one line, values aligned; no horizontal
     scroll; no single-letter lines.
   - Expect (280 px, 800 px tall): the start of the request body is visible without scrolling the
     panel (SC-002).
2. The same row plus `X-Very-Long-Custom-Correlation-Header-Name-For-Tracing-Abcdef` (61 chars)
   and a 400-character unbroken value. Expect: name column ≤ max(40% of the list, 16 characters), long name wraps at
   hyphens, value wraps in its cell, `scrollWidth === clientWidth` on the panel.
3. `ApiDocView` with an opened operation whose parameters are `Accept-Encoding` (required, 400-char
   example) and a 51-character name. Expect: `Accept-Encoding required` on one line; the 51-char
   name wraps within ≈ 32 characters; no horizontal scroll.

Take before (`origin/main` copies in `harness/before/`) and after screenshots; store them in
`specs/XEN-024-runner-detail-header-wrap/screenshots/` and attach them to the PR.
