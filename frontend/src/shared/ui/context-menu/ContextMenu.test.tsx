import { fireEvent, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import { ContextMenu } from './ContextMenu'

function Example({ onRename, onDelete }: { onRename: () => void; onDelete: () => void }) {
  return (
    <div>
      <ContextMenu.Root>
        <ContextMenu.Trigger>Billing API</ContextMenu.Trigger>
        <ContextMenu.Menu label="Collection actions">
          <ContextMenu.Item onSelect={onRename}>Rename</ContextMenu.Item>
          <ContextMenu.Item tone="danger" onSelect={onDelete}>
            Delete
          </ContextMenu.Item>
        </ContextMenu.Menu>
      </ContextMenu.Root>
      <p>elsewhere</p>
    </div>
  )
}

describe('ContextMenu', () => {
  it('opens at the pointer and runs the chosen action', async () => {
    const onDelete = vi.fn()
    render(<Example onRename={vi.fn()} onDelete={onDelete} />)
    fireEvent.contextMenu(screen.getByText('Billing API'), { clientX: 40, clientY: 60 })
    const menu = screen.getByRole('menu', { name: 'Collection actions' })
    expect(menu).toHaveStyle({ left: '40px', top: '60px' })
    expect(screen.getByRole('menuitem', { name: 'Rename' })).toHaveFocus()
    await userEvent.keyboard('{ArrowDown}{Enter}')
    expect(onDelete).toHaveBeenCalledOnce()
    expect(screen.queryByRole('menu')).not.toBeInTheDocument()
  })

  it('closes on Escape and outside clicks', async () => {
    render(<Example onRename={vi.fn()} onDelete={vi.fn()} />)
    fireEvent.contextMenu(screen.getByText('Billing API'))
    await userEvent.keyboard('{Escape}')
    expect(screen.queryByRole('menu')).not.toBeInTheDocument()
    fireEvent.contextMenu(screen.getByText('Billing API'))
    await userEvent.click(screen.getByText('elsewhere'))
    expect(screen.queryByRole('menu')).not.toBeInTheDocument()
  })
})
