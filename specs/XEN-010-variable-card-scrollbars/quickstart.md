# Quickstart: Verifying XEN-010

## Prerequisites

- Branch `xen-010`, `npm ci` run in `frontend/`.
- macOS for the by-eye check (Wails toolchain as for `make package`).

## 1. Unit tests (red before the fix, green after)

```bash
cd frontend
npx vitest run src/shared/hooks/use-fit-height.test.tsx src/features/request-builder/components/VariableCard.test.tsx
```

- Before the fix: the new sizing cases fail (the hook file does not exist; the card sets no
  inline height/overflow).
- After: all pass, including the pre-existing card cases (SC-005).

```bash
npx vitest run src/features/environments/components/VariableTable.test.tsx
```

- Passes unchanged before and after (SC-006).

## 2. Full gate

```bash
make validate
```

Green (includes `frontend-check`: eslint, typecheck, all Vitest).

## 3. Packaged macOS app, scrollbars "Always" (FR-011)

```bash
defaults read -g AppleShowScrollBars 2>/dev/null || echo "(default: Automatic)"   # note it
defaults write -g AppleShowScrollBars Always
make package           # → bin/xenon.app
open bin/xenon.app
```

In a workspace with an active environment:

1. New `HEAD` request, URL `{{authPrvateUrl}}/health`; hover `{{authPrvateUrl}}`. The card says
   "is not defined"; the box is one line tall, "Enter value…" fully visible, **no scrollbars**.
2. Type `https://auth.dev` → still one line, no scrollbars.
3. Paste a ~200-character URL → wraps, box grows to fit (≤ 6 lines), no scrollbars; delete text →
   shrinks back to one line.
4. Paste ~500 characters → box stops at 6 lines, scrolls vertically only.
5. Add it to the environment; hover again → defined card, same behaviour. Mark a variable secret
   and hover it → "Secret — type to replace…" fully visible on one line, no scrollbars.
6. View → Zoom in to 200% and out to 67%; repeat steps 1–2.
7. Open the environment's variable table: long values still wrap and grow, no scrollbars; resize the
   window narrower and wider — values rewrap and the rows refit (the shared hook's
   resize refit, FR-012).

Take a screenshot of steps 1 and 5 for the implementation report, then restore the setting:

```bash
defaults write -g AppleShowScrollBars <previous value>   # or: defaults delete -g AppleShowScrollBars
```
