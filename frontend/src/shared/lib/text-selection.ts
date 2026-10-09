const DRAGGING_ATTRIBUTE = 'data-dragging'

// Turns off text selection across the whole document while something is being dragged, so
// moving the pointer over text (inputs and editors included) never starts a selection.
// Returns the function that turns it back on.
export function suspendTextSelection(): () => void {
  const root = document.documentElement
  root.setAttribute(DRAGGING_ATTRIBUTE, '')
  return () => root.removeAttribute(DRAGGING_ATTRIBUTE)
}
