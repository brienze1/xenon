import { describe, expect, it } from 'vitest'
import { suspendTextSelection } from './text-selection'

describe('suspendTextSelection', () => {
  it('marks the document as dragging until the returned function restores it', () => {
    const restore = suspendTextSelection()
    expect(document.documentElement).toHaveAttribute('data-dragging')
    restore()
    expect(document.documentElement).not.toHaveAttribute('data-dragging')
  })
})
