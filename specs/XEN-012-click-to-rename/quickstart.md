# Quickstart: validating click-to-rename

## Automated

```bash
cd frontend && npm test -- editable-name name-draft name-drafts TabBreadcrumb ItemBreadcrumb InfoPanel ItemTab EnvironmentTab TabStrip
make frontend-check          # lint + typecheck + every Vitest spec
make bdd TAG=@rename         # tree rename scenarios, unchanged and green
make bdd                     # the whole Godog suite, unchanged and green
make validate                # final gate
```

## Manual (in `wails3 dev`)

For each of: a saved request, a never-saved request, an example, a collection, a folder and an
environment:

1. Open its tab. The name appears once, as text: the breadcrumb's last crumb for requests and
   examples, the heading for the others. There is no Name input or Rename button anywhere in the
   tab, including the collection Overview. The request Info panel shows the name as click-to-edit
   text.
2. Click the name. An input appears with the whole name selected. Tab away, then focus the name
   and press Enter: the input appears again.
3. Type a new name. The yellow dot appears on the tab. The tab label and the sidebar still show
   the old name.
4. Press Escape. The old name is back, and the dot is gone if nothing else was edited.
5. Rename again, then edit one more field (URL, auth, a variable). Press Cmd/Ctrl+S. Both are
   saved; the label, sidebar and breadcrumb show the new name; the dot is gone.
6. Clear the name and press Cmd/Ctrl+S. The inline "Type a name for the …" message appears,
   nothing is saved (the other edited field is still dirty), and the dot stays.
7. Environment only: type another environment's name in different case and save. The inline
   "already exists" message appears and nothing is saved.
8. Paste 300 characters into the name. Only 200 are kept.
9. Rename without saving and close the tab. The discard prompt appears. Discard keeps the old
   name; Save renames and closes.
10. Rename in the tab without saving, then rename the same item from the sidebar with F2. The tab
    keeps its draft and dot, and its label shows the sidebar's new name.
11. Sidebar F2 rename and the workspace rename dialog still save straight away.
