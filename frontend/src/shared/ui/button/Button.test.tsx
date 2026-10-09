import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import { IconButton } from '@/shared/ui/icon-button/IconButton'
import { Button } from './Button'

describe('Button', () => {
  it('renders a non-submitting button with its variant', async () => {
    const onClick = vi.fn()
    render(
      <Button variant="accent" onClick={onClick}>
        Send
      </Button>,
    )
    const button = screen.getByRole('button', { name: 'Send' })
    expect(button).toHaveAttribute('type', 'button')
    expect(button.className).toContain('accent')
    await userEvent.click(button)
    expect(onClick).toHaveBeenCalledOnce()
  })

  it('labels icon buttons for assistive technology', () => {
    render(<IconButton label="Close tab">×</IconButton>)
    expect(screen.getByRole('button', { name: 'Close tab' })).toHaveAttribute('title', 'Close tab')
  })
})
