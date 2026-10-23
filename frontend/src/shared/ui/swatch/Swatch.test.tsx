import { render } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { Swatch } from './Swatch'

describe('Swatch', () => {
  it('is a decorative dot carrying its colour', () => {
    const { container } = render(<Swatch color="green" />)
    const dot = container.querySelector('span')
    expect(dot).toHaveAttribute('data-color', 'green')
    expect(dot).toHaveAttribute('aria-hidden', 'true')
  })
})
