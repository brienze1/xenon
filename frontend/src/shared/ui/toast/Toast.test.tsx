import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import { Toast } from './Toast'

describe('Toast', () => {
  it('announces its message politely, with its actions and a close button', async () => {
    const onClose = vi.fn()
    const onAction = vi.fn()
    render(
      <Toast.Root closeLabel="Dismiss" onClose={onClose}>
        <Toast.Message>Saved.</Toast.Message>
        <Toast.Actions>
          <button type="button" onClick={onAction}>
            Undo
          </button>
        </Toast.Actions>
      </Toast.Root>,
    )
    const toast = screen.getByRole('status')
    expect(toast).toHaveAttribute('aria-live', 'polite')
    expect(toast).toHaveTextContent('Saved.')
    await userEvent.click(screen.getByRole('button', { name: 'Undo' }))
    expect(onAction).toHaveBeenCalledTimes(1)
    await userEvent.click(screen.getByRole('button', { name: 'Dismiss' }))
    expect(onClose).toHaveBeenCalledTimes(1)
  })

  it('names the close button "Close" by default', () => {
    render(
      <Toast.Root onClose={vi.fn()}>
        <Toast.Message>Hi</Toast.Message>
      </Toast.Root>,
    )
    expect(screen.getByRole('button', { name: 'Close' })).toBeInTheDocument()
  })
})
