import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it } from 'vitest'
import { Tooltip } from './Tooltip'

describe('Tooltip', () => {
  it('shows on hover and focus and hides afterwards', async () => {
    render(
      <Tooltip content="DNS 4 ms">
        <button type="button">120 ms</button>
      </Tooltip>,
    )
    expect(screen.queryByRole('tooltip')).not.toBeInTheDocument()
    await userEvent.hover(screen.getByRole('button'))
    expect(screen.getByRole('tooltip')).toHaveTextContent('DNS 4 ms')
    await userEvent.unhover(screen.getByRole('button'))
    expect(screen.queryByRole('tooltip')).not.toBeInTheDocument()
    await userEvent.tab()
    const tooltip = screen.getByRole('tooltip')
    expect(screen.getByRole('button')).toHaveAttribute('aria-describedby', tooltip.id)
    expect(screen.getByRole('button').parentElement).not.toHaveAttribute('aria-describedby')
    await userEvent.tab()
    expect(screen.queryByRole('tooltip')).not.toBeInTheDocument()
  })
})
