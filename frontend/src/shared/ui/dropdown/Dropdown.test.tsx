import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { useState } from 'react'
import { describe, expect, it, vi } from 'vitest'
import { Dropdown } from './Dropdown'

interface ExampleProps {
  onSelect: (value: string) => void
}

function Example({ onSelect }: ExampleProps) {
  return (
    <div>
      <Dropdown.Root>
        <Dropdown.Trigger label="Method">GET</Dropdown.Trigger>
        <Dropdown.Menu label="Methods">
          <Dropdown.RadioItem checked onSelect={() => onSelect('GET')}>
            GET
          </Dropdown.RadioItem>
          <Dropdown.RadioItem checked={false} onSelect={() => onSelect('POST')}>
            POST
          </Dropdown.RadioItem>
          <Dropdown.Item onSelect={() => onSelect('custom')}>Custom</Dropdown.Item>
        </Dropdown.Menu>
      </Dropdown.Root>
      <p>outside</p>
    </div>
  )
}

interface ControlledProps {
  onOpenChange: (open: boolean) => void
}

function Controlled({ onOpenChange }: ControlledProps) {
  const [open, setOpen] = useState(false)
  return (
    <div>
      <button type="button" onClick={() => setOpen(true)}>
        open from outside
      </button>
      <Dropdown.Root
        open={open}
        onOpenChange={(next) => {
          setOpen(next)
          onOpenChange(next)
        }}
      >
        <Dropdown.Trigger label="Env">dev</Dropdown.Trigger>
        <Dropdown.Menu label="Envs">
          <Dropdown.RadioItem checked onSelect={vi.fn()}>
            dev
          </Dropdown.RadioItem>
          <Dropdown.Item disabled onSelect={vi.fn()}>
            broken
          </Dropdown.Item>
          <Dropdown.RadioItem checked={false} onSelect={vi.fn()}>
            hml
          </Dropdown.RadioItem>
          <Dropdown.RadioItem checked={false} onSelect={vi.fn()}>
            prod
          </Dropdown.RadioItem>
          <Dropdown.RadioItem checked={false} onSelect={vi.fn()}>
            preview
          </Dropdown.RadioItem>
        </Dropdown.Menu>
      </Dropdown.Root>
    </div>
  )
}

describe('Dropdown', () => {
  it('opens, focuses the first item and selects with the keyboard', async () => {
    const onSelect = vi.fn()
    render(<Example onSelect={onSelect} />)
    const trigger = screen.getByRole('button', { name: 'Method' })
    expect(trigger).toHaveAttribute('aria-expanded', 'false')
    await userEvent.click(trigger)
    expect(trigger).toHaveAttribute('aria-expanded', 'true')
    expect(screen.getByRole('menu', { name: 'Methods' })).toBeInTheDocument()
    expect(screen.getByRole('menuitemradio', { name: 'GET' })).toHaveFocus()
    expect(screen.getByRole('menuitemradio', { name: 'GET' })).toBeChecked()
    expect(screen.getByRole('menuitemradio', { name: 'POST' })).not.toBeChecked()
    expect(screen.getByRole('menuitem', { name: 'Custom' })).not.toHaveAttribute('aria-checked')
    await userEvent.keyboard('{ArrowDown}{Enter}')
    expect(onSelect).toHaveBeenCalledWith('POST')
    expect(screen.queryByRole('menu')).not.toBeInTheDocument()
    expect(trigger).toHaveFocus()
  })

  it('closes on Escape and on outside clicks', async () => {
    render(<Example onSelect={vi.fn()} />)
    await userEvent.click(screen.getByRole('button', { name: 'Method' }))
    await userEvent.keyboard('{Escape}')
    expect(screen.queryByRole('menu')).not.toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Method' })).toHaveFocus()
    await userEvent.click(screen.getByRole('button', { name: 'Method' }))
    await userEvent.click(screen.getByText('outside'))
    expect(screen.queryByRole('menu')).not.toBeInTheDocument()
  })

  it('can be opened from outside, reports open changes and skips disabled items', async () => {
    const onOpenChange = vi.fn()
    render(<Controlled onOpenChange={onOpenChange} />)
    await userEvent.click(screen.getByRole('button', { name: 'open from outside' }))
    expect(screen.getByRole('menu', { name: 'Envs' })).toBeInTheDocument()
    expect(screen.getByRole('menuitemradio', { name: 'dev' })).toHaveFocus()
    expect(screen.getByRole('menuitem', { name: 'broken' })).toBeDisabled()
    await userEvent.keyboard('{ArrowDown}')
    expect(screen.getByRole('menuitemradio', { name: 'hml' })).toHaveFocus()
    await userEvent.keyboard('{Escape}')
    expect(onOpenChange).toHaveBeenLastCalledWith(false)
    expect(screen.queryByRole('menu')).not.toBeInTheDocument()
    await userEvent.click(screen.getByRole('button', { name: 'Env' }))
    expect(onOpenChange).toHaveBeenLastCalledWith(true)
    await userEvent.click(screen.getByRole('button', { name: 'Env' }))
    expect(onOpenChange).toHaveBeenLastCalledWith(false)
  })

  it('moves focus to the item whose name starts with the typed text', async () => {
    render(<Controlled onOpenChange={vi.fn()} />)
    await userEvent.click(screen.getByRole('button', { name: 'Env' }))
    await userEvent.keyboard('p')
    expect(screen.getByRole('menuitemradio', { name: 'prod' })).toHaveFocus()
    await userEvent.keyboard('p')
    expect(screen.getByRole('menuitemradio', { name: 'preview' })).toHaveFocus()
    await userEvent.keyboard('x')
    expect(screen.getByRole('menuitemradio', { name: 'preview' })).toHaveFocus()
  })
})
