import { act, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { CopyButton, COPIED_FOR_MS } from './CopyButton'

describe('CopyButton', () => {
  afterEach(() => vi.useRealTimers())

  it('is an icon button named by its label, never the word Copy', () => {
    render(<CopyButton label="Copy body" onCopy={vi.fn().mockResolvedValue(true)} />)
    const button = screen.getByRole('button', { name: 'Copy body' })
    expect(button.textContent).toBe('')
    expect(button.querySelector('svg')).not.toBeNull()
    expect(button).toHaveAttribute('data-state', 'idle')
  })

  it('swaps to a check and announces the copy, then goes back to the copy icon', async () => {
    vi.useFakeTimers({ shouldAdvanceTime: true })
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime })
    const onCopy = vi.fn().mockResolvedValue(true)
    render(<CopyButton label="Copy body" onCopy={onCopy} />)
    await user.click(screen.getByRole('button', { name: 'Copy body' }))
    expect(onCopy).toHaveBeenCalledOnce()
    const button = screen.getByRole('button', { name: 'Copy body' })
    expect(button).toHaveAttribute('data-state', 'copied')
    expect(button).toHaveAttribute('title', 'Copied')
    expect(screen.getByText('Copied to the clipboard.')).toHaveAttribute('aria-live', 'polite')
    await act(() => vi.advanceTimersByTimeAsync(COPIED_FOR_MS))
    expect(button).toHaveAttribute('data-state', 'idle')
    expect(button).toHaveAttribute('title', 'Copy body')
    expect(screen.queryByText('Copied to the clipboard.')).toBeNull()
  })

  it('tells the user when the copy fails, with a custom message when given', async () => {
    const user = userEvent.setup()
    render(<CopyButton label="Copy token" onCopy={vi.fn().mockResolvedValue(false)} failedMessage="Select the token and copy it by hand." />)
    await user.click(screen.getByRole('button', { name: 'Copy token' }))
    const button = screen.getByRole('button', { name: 'Copy token' })
    expect(button).toHaveAttribute('data-state', 'failed')
    expect(button).toHaveAttribute('title', 'Copy failed')
    expect(screen.getByText('Select the token and copy it by hand.')).toHaveAttribute('aria-live', 'polite')
  })

  it('treats a rejected copy as a failure and stays disabled when asked', async () => {
    const user = userEvent.setup()
    const { rerender } = render(<CopyButton label="Copy key" onCopy={vi.fn().mockRejectedValue(new Error('denied'))} />)
    await user.click(screen.getByRole('button', { name: 'Copy key' }))
    expect(screen.getByText('Couldn’t copy to the clipboard.')).toBeInTheDocument()
    rerender(<CopyButton label="Copy key" disabled onCopy={vi.fn()} />)
    expect(screen.getByRole('button', { name: 'Copy key' })).toBeDisabled()
  })
})
