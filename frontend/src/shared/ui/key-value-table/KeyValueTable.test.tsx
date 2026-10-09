import { fireEvent, render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import { KeyValueTable, type KeyValueMode, type KeyValueRow } from './KeyValueTable'

const EXAMPLES = { key: 'page', value: '2', description: 'Page to fetch' }

const rows: KeyValueRow[] = [
  { id: 'r1', key: 'page', value: '2', description: '', enabled: true },
  { id: 'r2', key: 'limit', value: '10', description: 'page size', enabled: false },
]

function renderTable(mode: KeyValueMode = 'table', bulkText = '') {
  const handlers = {
    onModeChange: vi.fn(),
    onBulkTextChange: vi.fn(),
    onRowChange: vi.fn(),
    onRowAdd: vi.fn(),
    onRowRemove: vi.fn(),
  }
  render(<KeyValueTable label="Query params" rows={rows} mode={mode} bulkText={bulkText} examples={EXAMPLES} {...handlers} />)
  return handlers
}

describe('KeyValueTable', () => {
  it('lists the rows with their enabled state', () => {
    renderTable()
    const table = screen.getByRole('table', { name: 'Query params' })
    expect(within(table).getAllByRole('checkbox')).toHaveLength(2)
    expect(screen.getByRole('checkbox', { name: 'Enable page' })).toBeChecked()
    expect(screen.getByRole('checkbox', { name: 'Enable limit' })).not.toBeChecked()
    expect(screen.getAllByRole('textbox', { name: 'Description' })[1]).toHaveValue('page size')
  })

  it('reports edits, toggles and removals per row', async () => {
    const handlers = renderTable()
    fireEvent.change(screen.getAllByRole('textbox', { name: 'Value' })[0] as HTMLElement, { target: { value: '3' } })
    expect(handlers.onRowChange).toHaveBeenCalledWith('r1', { value: '3' })
    await userEvent.click(screen.getByRole('checkbox', { name: 'Enable limit' }))
    expect(handlers.onRowChange).toHaveBeenCalledWith('r2', { enabled: true })
    await userEvent.click(screen.getByRole('button', { name: 'Remove page' }))
    expect(handlers.onRowRemove).toHaveBeenCalledWith('r1')
  })

  it('adds a row when typing into the empty last row', async () => {
    const handlers = renderTable()
    await userEvent.type(screen.getByRole('textbox', { name: 'New key' }), 'q')
    expect(handlers.onRowAdd).toHaveBeenCalledWith({ key: 'q', value: '', description: '', enabled: true })
    await userEvent.type(screen.getByRole('textbox', { name: 'New value' }), 'v')
    expect(handlers.onRowAdd).toHaveBeenLastCalledWith({ key: '', value: 'v', description: '', enabled: true })
  })

  it('switches to bulk edit through the toggle', async () => {
    const handlers = renderTable()
    const toggle = screen.getByRole('button', { name: 'Bulk Edit' })
    expect(toggle).toHaveAttribute('aria-pressed', 'false')
    await userEvent.click(toggle)
    expect(handlers.onModeChange).toHaveBeenCalledWith('bulk')
  })

  it('edits the rows as text in bulk mode', async () => {
    const handlers = renderTable('bulk', 'page:2\n//limit:10')
    expect(screen.queryByRole('table')).not.toBeInTheDocument()
    const editor = screen.getByRole('textbox', { name: 'Query params bulk edit' })
    expect(editor).toHaveValue('page:2\n//limit:10')
    fireEvent.change(editor, { target: { value: 'page:3' } })
    expect(handlers.onBulkTextChange).toHaveBeenCalledWith('page:3')
    await userEvent.click(screen.getByRole('button', { name: 'Key-Value Edit' }))
    expect(handlers.onModeChange).toHaveBeenCalledWith('table')
  })

  it('names unnamed rows generically', () => {
    render(
      <KeyValueTable
        label="Headers"
        rows={[{ id: 'x', key: '', value: '', description: '', enabled: true }]}
        mode="table"
        bulkText=""
        onModeChange={vi.fn()}
        onBulkTextChange={vi.fn()}
        onRowChange={vi.fn()}
        onRowAdd={vi.fn()}
        onRowRemove={vi.fn()}
        examples={EXAMPLES}
      />,
    )
    expect(screen.getByRole('checkbox', { name: 'Enable row' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Remove Row' })).toBeInTheDocument()
  })

  it('shows example placeholders in the new row', () => {
    renderTable()
    expect(screen.getByRole('textbox', { name: 'New key' })).toHaveAttribute('placeholder', 'page…')
    expect(screen.getByRole('textbox', { name: 'New value' })).toHaveAttribute('placeholder', '2…')
  })
})
