import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import { Dialog } from './Dialog'

interface ExampleProps {
  open: boolean
  onOpenChange: (open: boolean) => void
}

function Example({ open, onOpenChange }: ExampleProps) {
  return (
    <Dialog.Root open={open} onOpenChange={onOpenChange}>
      <Dialog.Title>Unsaved changes</Dialog.Title>
      <Dialog.Body>Save before closing?</Dialog.Body>
      <Dialog.Actions>
        <button type="button">Save</button>
        <button type="button">Discard</button>
      </Dialog.Actions>
    </Dialog.Root>
  )
}

describe('Dialog', () => {
  it('renders nothing while closed', () => {
    render(<Example open={false} onOpenChange={vi.fn()} />)
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
  })

  it('is a labelled modal that focuses its first control', () => {
    render(<Example open onOpenChange={vi.fn()} />)
    expect(screen.getByRole('dialog', { name: 'Unsaved changes' })).toHaveAttribute('aria-modal', 'true')
    expect(screen.getByRole('button', { name: 'Save' })).toHaveFocus()
  })

  it('asks to close on Escape and on backdrop clicks only', async () => {
    const onOpenChange = vi.fn()
    render(<Example open onOpenChange={onOpenChange} />)
    await userEvent.click(screen.getByText('Save before closing?'))
    expect(onOpenChange).not.toHaveBeenCalled()
    await userEvent.keyboard('{Escape}')
    expect(onOpenChange).toHaveBeenCalledWith(false)
    const backdrop = screen.getByRole('dialog').parentElement as HTMLElement
    await userEvent.click(backdrop)
    expect(onOpenChange).toHaveBeenCalledTimes(2)
  })

  it('keeps Tab focus inside the dialog', async () => {
    render(
      <>
        <button type="button">Outside</button>
        <Example open onOpenChange={vi.fn()} />
      </>,
    )
    const save = screen.getByRole('button', { name: 'Save' })
    const discard = screen.getByRole('button', { name: 'Discard' })
    expect(save).toHaveFocus()
    await userEvent.tab()
    expect(discard).toHaveFocus()
    await userEvent.tab()
    expect(save).toHaveFocus()
    await userEvent.tab({ shift: true })
    expect(discard).toHaveFocus()
  })
})
