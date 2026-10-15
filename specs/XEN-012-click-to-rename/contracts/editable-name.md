# Contract: `EditableName` (`frontend/src/shared/ui/editable-name/EditableName.tsx`)

A domain-free shared component. It doesn't know about stores, tabs or saving.

## Props

```ts
interface EditableNameProps {
  /** The name to show: the tab's draft name (falls back to the saved name when there is no draft). */
  value: string
  /** What is renamed, e.g. "Rename request". The input's accessible name; the display button is "<label>: <value>" so the name itself is still announced. */
  label: string
  /** Called on every keystroke with the raw input value (not trimmed). */
  onChange: (value: string) => void
  /** Escape while editing: the caller drops the draft so `value` becomes the saved name again. */
  onCancel: () => void
  /** Shown instead of an empty value, e.g. "Untitled Request" for a never-saved request; never part of the value. */
  placeholder?: string
  /** Inline error under the name (already-formatted message), or null. */
  error?: string | null
  /** Upper bound on characters; default 200 (`maxItemNameLength`). */
  maxLength?: number
}
```

## Behaviour

| Situation | Result |
| --- | --- |
| Not editing | `<button type="button" aria-label={`${label}: ${value}`} title="Click to rename">` (the label alone would hide the name from screen readers and from the heading's accessible name) showing `value` (`translate="no"`), one line with `text-overflow: ellipsis`, pointer cursor, a hover background and a visible focus ring from the tokens (the look of today's `ItemBreadcrumb` `.rename` crumb). The `title` is "Click to rename" plus the full name when it is cut. No `<input>` in the DOM. |
| Click the button, or focus it and press Enter or Space | Renders `<input aria-label={label} maxLength={maxLength} value={value} spellCheck={false}>` in place. It is focused with the whole text selected. |
| Type | `onChange(next)` on every change. The input is controlled by `value`. |
| Enter | `stopPropagation`; the input closes (back to the button showing `value`) and focus returns to the button. `onChange` and `onCancel` are not called and nothing is saved. |
| Escape | `stopPropagation`; `onCancel()`; the input closes and focus returns to the button. |
| Blur (Tab, click elsewhere) | The input closes; `value` is kept; focus is not moved back. |
| Input size | Inherits the font and line height of the text it replaces (`font: inherit`) and fills the same width, so the heading or crumb doesn't jump. |
| Paste longer than `maxLength` | The browser cuts it at `maxLength` (native `maxLength`). |
| Mod+S while editing | Not handled here. It bubbles to the app-wide `save` binding. |
| `error` set | Renders `<p role="alert" id={errorId}>` under the name. The button or input gets `aria-invalid="true"` and `aria-describedby={errorId}`. |
| `value` is empty while not editing | The button shows the empty string with a visible focus ring and a min-width, so it stays clickable. |

The component keeps only the `editing` boolean (`useState`). The value comes from the caller.
