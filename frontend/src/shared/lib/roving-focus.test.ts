import { afterEach, describe, expect, it } from 'vitest'
import { moveFocus, rovingIndex } from './roving-focus'

function menu(count: number): HTMLElement {
  const container = document.createElement('div')
  for (let index = 0; index < count; index++) {
    const item = document.createElement('button')
    item.setAttribute('role', 'menuitem')
    item.textContent = `item ${index}`
    container.append(item)
  }
  document.body.append(container)
  return container
}

describe('moveFocus', () => {
  afterEach(() => {
    document.body.innerHTML = ''
  })

  it('moves through items and wraps around', () => {
    const container = menu(3)
    const items = Array.from(container.querySelectorAll('button'))
    expect(moveFocus(container, 'ArrowDown', '[role="menuitem"]')).toBe(true)
    expect(items[0]).toHaveFocus()
    moveFocus(container, 'ArrowUp', '[role="menuitem"]')
    expect(items[2]).toHaveFocus()
    moveFocus(container, 'ArrowDown', '[role="menuitem"]')
    expect(items[0]).toHaveFocus()
  })

  it('starts from the last item when moving up from outside', () => {
    const container = menu(2)
    moveFocus(container, 'ArrowUp', '[role="menuitem"]')
    expect(container.querySelectorAll('button')[1]).toHaveFocus()
  })

  it('ignores other keys and handles empty menus', () => {
    expect(moveFocus(menu(1), 'Enter', '[role="menuitem"]')).toBe(false)
    expect(moveFocus(menu(0), 'ArrowDown', '[role="menuitem"]')).toBe(true)
  })

  it('jumps to the ends with Home and End', () => {
    const container = menu(3)
    const items = Array.from(container.querySelectorAll('button'))
    moveFocus(container, 'End', '[role="menuitem"]')
    expect(items[2]).toHaveFocus()
    moveFocus(container, 'Home', '[role="menuitem"]')
    expect(items[0]).toHaveFocus()
  })

  it('follows the horizontal arrows for horizontal lists', () => {
    const container = menu(2)
    const items = Array.from(container.querySelectorAll('button'))
    expect(moveFocus(container, 'ArrowDown', '[role="menuitem"]', 'horizontal')).toBe(false)
    moveFocus(container, 'ArrowRight', '[role="menuitem"]', 'horizontal')
    expect(items[0]).toHaveFocus()
    moveFocus(container, 'ArrowLeft', '[role="menuitem"]', 'horizontal')
    expect(items[1]).toHaveFocus()
  })
})

describe('rovingIndex', () => {
  it('stops at the ends when wrapping is off', () => {
    expect(rovingIndex(3, 2, 'ArrowDown', { wrap: false })).toBe(2)
    expect(rovingIndex(3, 0, 'ArrowUp', { wrap: false })).toBe(0)
    expect(rovingIndex(3, 1, 'ArrowDown', { wrap: false })).toBe(2)
    expect(rovingIndex(3, 1, 'End', { wrap: false })).toBe(2)
    expect(rovingIndex(0, 0, 'ArrowDown')).toBeNull()
    expect(rovingIndex(3, 0, 'Enter')).toBeNull()
  })
})
