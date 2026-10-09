import { afterEach, describe, expect, it } from 'vitest'
import { moveFocus } from './roving-focus'

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
})
