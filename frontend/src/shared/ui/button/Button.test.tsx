import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import { IconButton } from '@/shared/ui/icon-button/IconButton'
import { Button } from './Button'

describe('Button', () => {
  it('renders a non-submitting button', async () => {
    const onClick = vi.fn()
    render(
      <Button variant="accent" onClick={onClick}>
        Send
      </Button>,
    )
    const button = screen.getByRole('button', { name: 'Send' })
    expect(button).toHaveAttribute('type', 'button')
    await userEvent.click(button)
    expect(onClick).toHaveBeenCalledOnce()
  })

  it('submits a form only when asked to', async () => {
    const onSubmit = vi.fn((event: SubmitEvent) => event.preventDefault())
    render(
      <form aria-label="Example" onSubmit={(event) => onSubmit(event.nativeEvent as SubmitEvent)}>
        <Button>Plain</Button>
        <Button type="submit">Submit</Button>
      </form>,
    )
    await userEvent.click(screen.getByRole('button', { name: 'Plain' }))
    expect(onSubmit).not.toHaveBeenCalled()
    await userEvent.click(screen.getByRole('button', { name: 'Submit' }))
    expect(onSubmit).toHaveBeenCalledOnce()
  })

  it('labels icon buttons for assistive technology', () => {
    render(<IconButton label="Close Tab">×</IconButton>)
    expect(screen.getByRole('button', { name: 'Close Tab' })).toHaveAttribute('title', 'Close Tab')
  })
})
