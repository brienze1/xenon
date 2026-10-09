import { fireEvent, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import { ContextMenu } from './ContextMenu'

interface ExampleProps {
  onRename: () => void
  onDelete: () => void
}

function Example({ onRename, onDelete }: ExampleProps) {
  return (
    <div>
      <ContextMenu.Root>
        <ContextMenu.Trigger>
          <span tabIndex={0}>Billing API</span>
        </ContextMenu.Trigger>
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

  it('opens from the keyboard with Shift+F10 at the focused element', async () => {
    const onRename = vi.fn()
    render(<Example onRename={onRename} onDelete={vi.fn()} />)
    const trigger = screen.getByText('Billing API')
    trigger.focus()
    await userEvent.keyboard('{Shift>}{F10}{/Shift}')
    expect(screen.getByRole('menu', { name: 'Collection actions' })).toBeInTheDocument()
    expect(screen.getByRole('menuitem', { name: 'Rename' })).toHaveFocus()
    await userEvent.keyboard('{End}')
    expect(screen.getByRole('menuitem', { name: 'Delete' })).toHaveFocus()
    await userEvent.keyboard('{Home}')
    expect(screen.getByRole('menuitem', { name: 'Rename' })).toHaveFocus()
    await userEvent.keyboard('{Enter}')
    expect(onRename).toHaveBeenCalledOnce()
    expect(trigger).toHaveFocus()
  })

  it('opens from the ContextMenu key and returns focus to the trigger on Escape', async () => {
    render(<Example onRename={vi.fn()} onDelete={vi.fn()} />)
    const trigger = screen.getByText('Billing API')
    trigger.focus()
    fireEvent.keyDown(trigger, { key: 'ContextMenu' })
    expect(screen.getByRole('menu', { name: 'Collection actions' })).toBeInTheDocument()
    await userEvent.keyboard('{Escape}')
    expect(screen.queryByRole('menu')).not.toBeInTheDocument()
    expect(trigger).toHaveFocus()
  })

  it('returns focus to the focusable trigger after a pointer-opened menu closes', async () => {
    render(<Example onRename={vi.fn()} onDelete={vi.fn()} />)
    const trigger = screen.getByText('Billing API')
    fireEvent.contextMenu(trigger)
    await userEvent.keyboard('{Escape}')
    expect(trigger).toHaveFocus()
  })
})
