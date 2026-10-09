import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { useState } from 'react'
import { describe, expect, it } from 'vitest'
import { Tabs } from './Tabs'

function Example() {
  const [value, setValue] = useState('params')
  return (
    <Tabs.Root value={value} onValueChange={setValue}>
      <Tabs.List label="Request">
        <Tabs.Tab value="params">Params</Tabs.Tab>
        <Tabs.Tab value="auth">Authorization</Tabs.Tab>
        <Tabs.Tab value="scripts" disabled>
          Scripts
        </Tabs.Tab>
        <Tabs.Tab value="headers">Headers</Tabs.Tab>
      </Tabs.List>
      <Tabs.Panel value="params">params panel</Tabs.Panel>
      <Tabs.Panel value="auth">auth panel</Tabs.Panel>
      <Tabs.Panel value="headers">headers panel</Tabs.Panel>
    </Tabs.Root>
  )
}

describe('Tabs', () => {
  it('shows the selected panel linked to its tab', () => {
    render(<Example />)
    const tab = screen.getByRole('tab', { name: 'Params' })
    expect(tab).toHaveAttribute('aria-selected', 'true')
    const panel = screen.getByRole('tabpanel')
    expect(panel).toHaveTextContent('params panel')
    expect(panel).toHaveAttribute('aria-labelledby', tab.id)
    expect(tab).toHaveAttribute('aria-controls', panel.id)
    expect(screen.queryByText('auth panel')).not.toBeInTheDocument()
  })

  it('switches panels on click', async () => {
    render(<Example />)
    await userEvent.click(screen.getByRole('tab', { name: 'Authorization' }))
    expect(screen.getByRole('tab', { name: 'Authorization' })).toHaveAttribute('aria-selected', 'true')
    expect(screen.getByRole('tabpanel')).toHaveTextContent('auth panel')
  })

  it('moves between enabled tabs with the arrow keys, wrapping around', async () => {
    render(<Example />)
    screen.getByRole('tab', { name: 'Params' }).focus()
    await userEvent.keyboard('{ArrowRight}')
    expect(screen.getByRole('tab', { name: 'Authorization' })).toHaveFocus()
    await userEvent.keyboard('{ArrowRight}')
    expect(screen.getByRole('tab', { name: 'Headers' })).toHaveAttribute('aria-selected', 'true')
    await userEvent.keyboard('{ArrowRight}')
    expect(screen.getByRole('tabpanel')).toHaveTextContent('params panel')
    await userEvent.keyboard('{ArrowLeft}')
    expect(screen.getByRole('tabpanel')).toHaveTextContent('headers panel')
    await userEvent.keyboard('{Home}')
    expect(screen.getByRole('tabpanel')).toHaveTextContent('params panel')
    await userEvent.keyboard('{End}')
    expect(screen.getByRole('tabpanel')).toHaveTextContent('headers panel')
  })

  it('refuses parts rendered outside the root', () => {
    expect(() => render(<Tabs.Tab value="x">x</Tabs.Tab>)).toThrow('Tabs parts must be rendered inside <Tabs.Root>')
  })
})
