import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it } from 'vitest'
import { generatedDocument } from '@/features/documentation/test-fixtures'
import { injectStylesheets } from '@/test/stylesheets'
import { ApiDocView } from './ApiDocView'

const LONG_NAME = 'a-very-long-parameter-name-that-exceeds-the-cap-xyz'

const document = {
  ...generatedDocument,
  operations: [
    {
      ...generatedDocument.operations[0]!,
      parameters: [
        { name: 'Accept-Encoding', in: 'header' as const, description: 'Encodings', example: 'x'.repeat(400), required: true, variable: false },
        { name: LONG_NAME, in: 'query' as const, description: '', example: '1', required: false, variable: false },
      ],
    },
  ],
}

describe('ApiDocView', () => {
  it('keeps parameter names on one line inside a block capped at 32 characters', async () => {
    injectStylesheets('src/features/documentation/components/ApiDocView.module.css')
    render(<ApiDocView kind="collection" document={document} />)
    await userEvent.click(screen.getByText('/invoices'))
    const parameters = await screen.findByRole('region', { name: 'List invoices parameters' })
    const [, first, second] = within(parameters).getAllByRole('row')
    const [nameCell, , , exampleCell] = within(first!).getAllByRole('cell')
    const name = nameCell!.querySelector('span')!
    expect(name.textContent).toBe('Accept-Encoding')
    expect(name.nextElementSibling).toHaveTextContent('required')
    const cell = getComputedStyle(nameCell!)
    expect(cell.whiteSpace).toBe('nowrap')
    expect(cell.overflowWrap).not.toBe('anywhere')
    const block = getComputedStyle(name)
    expect(block.display).toBe('inline-block')
    expect(block.width).toBe('max-content')
    expect(block.maxWidth).toBe('256px')
    expect(block.whiteSpace).toBe('normal')
    expect(block.overflowWrap).toBe('anywhere')
    expect(getComputedStyle(exampleCell!).overflowWrap).toBe('anywhere')
    expect(within(second!).getAllByRole('cell')[0]!.querySelector('span')!.textContent).toBe(LONG_NAME)
  })
})
