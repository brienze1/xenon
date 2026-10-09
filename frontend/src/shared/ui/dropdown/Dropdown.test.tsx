import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import { Dropdown } from './Dropdown'

function Example({ onSelect }: { onSelect: (value: string) => void }) {
  return (
    <div>
      <Dropdown.Root>
        <Dropdown.Trigger label="Method">GET</Dropdown.Trigger>
        <Dropdown.Menu label="Methods">
          <Dropdown.Item checked onSelect={() => onSelect('GET')}>
            GET
          </Dropdown.Item>
          <Dropdown.Item checked={false} onSelect={() => onSelect('POST')}>
            POST
          </Dropdown.Item>
        </Dropdown.Menu>
      </Dropdown.Root>
      <p>outside</p>
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
})
