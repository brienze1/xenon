import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { ConsoleDrawer } from './ConsoleDrawer'

describe('ConsoleDrawer', () => {
  it('explains that the console arrives with scripts', () => {
    render(<ConsoleDrawer />)
    const empty = screen.getByRole('region', { name: 'Console is empty' })
    expect(empty).toContainElement(screen.getByRole('heading', { name: 'Console Arrives with Scripts' }))
    expect(empty.querySelector('[data-art="console"] svg')).toBeInTheDocument()
  })
})
