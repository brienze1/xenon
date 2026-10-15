import { fireEvent, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { useState } from 'react'
import { describe, expect, it, vi } from 'vitest'
import { EditableName } from './EditableName'

interface HarnessProps {
  initial?: string
  error?: string | null
  maxLength?: number
  onChange?: (value: string) => void
  onCancel?: () => void
  onParentKeyDown?: (key: string) => void
}

/** Owns the value like a tab does: typing updates it, Escape puts the saved name back. */
function Harness({ initial = 'Create invoice', error = null, maxLength, onChange, onCancel, onParentKeyDown }: HarnessProps) {
  const [value, setValue] = useState(initial)
  return (
    <div onKeyDown={(event) => onParentKeyDown?.(event.key)}>
      <EditableName
        value={value}
        label="Rename request"
        maxLength={maxLength}
        error={error}
        onChange={(next) => {
          setValue(next)
          onChange?.(next)
        }}
        onCancel={() => {
          setValue(initial)
          onCancel?.()
        }}
      />
      <button type="button">Elsewhere</button>
    </div>
  )
}

function nameButton() {
  return screen.getByRole('button', { name: /^Rename request/ })
}

function nameInput() {
  return screen.getByRole('textbox', { name: 'Rename request' }) as HTMLInputElement
}

describe('EditableName', () => {
  it('shows the name as a button and no input until it is edited', () => {
    render(<Harness />)
    expect(nameButton()).toHaveTextContent('Create invoice')
    expect(nameButton()).toHaveAttribute('title', expect.stringContaining('Click to rename'))
    expect(screen.queryByRole('textbox')).not.toBeInTheDocument()
  })

  it('turns into an input with the whole name selected when clicked', async () => {
    render(<Harness />)
    await userEvent.click(nameButton())
    const input = nameInput()
    expect(input).toHaveFocus()
    expect(input).toHaveValue('Create invoice')
    expect([input.selectionStart, input.selectionEnd]).toEqual([0, 'Create invoice'.length])
  })

  it('turns into an input from the keyboard with Enter or Space', async () => {
    render(<Harness />)
    nameButton().focus()
    await userEvent.keyboard('{Enter}')
    expect(nameInput()).toHaveFocus()
    await userEvent.keyboard('{Escape}')
    nameButton().focus()
    await userEvent.keyboard(' ')
    expect(nameInput()).toHaveFocus()
  })

  it('reports every keystroke without trimming', async () => {
    const onChange = vi.fn()
    render(<Harness onChange={onChange} />)
    await userEvent.click(nameButton())
    await userEvent.clear(nameInput())
    await userEvent.type(nameInput(), 'Ping ')
    expect(onChange).toHaveBeenLastCalledWith('Ping ')
    expect(onChange).toHaveBeenCalledTimes(6)
  })

  it('closes on Enter keeping the draft, and gives focus back to the name', async () => {
    const onCancel = vi.fn()
    render(<Harness onCancel={onCancel} />)
    await userEvent.click(nameButton())
    await userEvent.clear(nameInput())
    await userEvent.type(nameInput(), 'Draft{Enter}')
    expect(screen.queryByRole('textbox')).not.toBeInTheDocument()
    expect(nameButton()).toHaveTextContent('Draft')
    expect(nameButton()).toHaveFocus()
    expect(onCancel).not.toHaveBeenCalled()
  })

  it('cancels on Escape, shows the saved name again, and gives focus back to the name', async () => {
    const onCancel = vi.fn()
    render(<Harness onCancel={onCancel} />)
    await userEvent.click(nameButton())
    await userEvent.type(nameInput(), ' edited')
    await userEvent.keyboard('{Escape}')
    expect(onCancel).toHaveBeenCalledTimes(1)
    expect(screen.queryByRole('textbox')).not.toBeInTheDocument()
    expect(nameButton()).toHaveTextContent('Create invoice')
    expect(nameButton()).toHaveFocus()
  })

  it('closes on blur keeping the draft and leaves the focus where it went', async () => {
    render(<Harness />)
    await userEvent.click(nameButton())
    await userEvent.type(nameInput(), ' v2')
    await userEvent.click(screen.getByRole('button', { name: 'Elsewhere' }))
    expect(screen.queryByRole('textbox')).not.toBeInTheDocument()
    expect(nameButton()).toHaveTextContent('Create invoice v2')
    expect(screen.getByRole('button', { name: 'Elsewhere' })).toHaveFocus()
  })

  it('accepts at most 200 characters by default, also when pasting', async () => {
    render(<Harness initial="" />)
    await userEvent.click(nameButton())
    expect(nameInput()).toHaveAttribute('maxLength', '200')
    await userEvent.paste('x'.repeat(300))
    expect(nameInput().value).toHaveLength(200)
  })

  it('takes another maximum length', async () => {
    render(<Harness maxLength={10} />)
    await userEvent.click(nameButton())
    expect(nameInput()).toHaveAttribute('maxLength', '10')
  })

  it('shows an error as an alert linked to the name', async () => {
    render(<Harness error="Type a name for the request." />)
    const alert = screen.getByRole('alert')
    expect(alert).toHaveTextContent('Type a name for the request.')
    expect(nameButton()).toHaveAttribute('aria-invalid', 'true')
    expect(nameButton()).toHaveAccessibleDescription('Type a name for the request.')
    await userEvent.click(nameButton())
    expect(nameInput()).toHaveAttribute('aria-invalid', 'true')
    expect(nameInput()).toHaveAccessibleDescription('Type a name for the request.')
  })

  it('keeps Enter and Escape away from the tab around it but lets the save hotkey through', async () => {
    const onParentKeyDown = vi.fn()
    render(<Harness onParentKeyDown={onParentKeyDown} />)
    await userEvent.click(nameButton())
    fireEvent.keyDown(nameInput(), { key: 's', metaKey: true })
    expect(onParentKeyDown).toHaveBeenCalledWith('s')
    fireEvent.keyDown(nameInput(), { key: 'Enter' })
    await userEvent.click(nameButton())
    fireEvent.keyDown(nameInput(), { key: 'Escape' })
    expect(onParentKeyDown).not.toHaveBeenCalledWith('Enter')
    expect(onParentKeyDown).not.toHaveBeenCalledWith('Escape')
  })

  it('stays clickable when the name is empty', async () => {
    render(<Harness initial="" />)
    await userEvent.click(nameButton())
    expect(nameInput()).toHaveValue('')
  })
})
