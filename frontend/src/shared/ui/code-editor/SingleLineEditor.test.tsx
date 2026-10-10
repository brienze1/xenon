import { act, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { EditorView } from '@codemirror/view'
import { describe, expect, it, vi } from 'vitest'
import type { CompletionSource, EditorDecoration } from './editor-extensions'
import { SingleLineEditor } from './SingleLineEditor'

/** CodeMirror ignores keys for a moment after the list opens (autocompletion `interactionDelay`). */
const COMPLETION_INTERACTION_DELAY_MS = 100

function editorView(label: string): EditorView {
  const view = EditorView.findFromDOM(screen.getByRole('textbox', { name: label }))
  if (!view) {
    throw new Error(`no editor labelled ${label}`)
  }
  return view
}

describe('SingleLineEditor', () => {
  it('is a labelled one-line textbox that reports typing and Enter', async () => {
    const onChange = vi.fn()
    const onEnter = vi.fn()
    render(<SingleLineEditor value="" label="Request URL" onChange={onChange} onEnter={onEnter} />)
    const textbox = screen.getByRole('textbox', { name: 'Request URL' })
    expect(textbox).toHaveAttribute('aria-multiline', 'false')
    await userEvent.type(textbox, 'ab{Enter}')
    expect(onChange).toHaveBeenLastCalledWith('ab')
    expect(onEnter).toHaveBeenCalledTimes(1)
    expect(textbox).toHaveTextContent('ab')
  })

  it('turns line breaks into spaces', () => {
    const onChange = vi.fn()
    render(<SingleLineEditor value="" label="Value" onChange={onChange} />)
    act(() => editorView('Value').dispatch({ changes: { from: 0, insert: 'a\nb\r\nc' } }))
    expect(onChange).toHaveBeenLastCalledWith('a b c')
  })

  it('follows the value prop without echoing it through onChange', async () => {
    const onChange = vi.fn()
    const { rerender } = render(<SingleLineEditor value="one" label="Value" onChange={onChange} />)
    rerender(<SingleLineEditor value="two" label="Value" onChange={onChange} />)
    await waitFor(() => expect(screen.getByRole('textbox', { name: 'Value' })).toHaveTextContent('two'))
    expect(onChange).not.toHaveBeenCalled()
  })

  it('marks decorated ranges with their tone and tooltip', async () => {
    const decorations: EditorDecoration[] = [
      { from: 0, to: 6, tone: 'success', tooltip: 'resolved' },
      { from: 7, to: 13, tone: 'danger', tooltip: 'missing' },
      { from: 20, to: 30, tone: 'info', tooltip: 'out of range' },
    ]
    const { container, rerender } = render(<SingleLineEditor value="{{ok}}/{{no}}" label="Value" onChange={vi.fn()} decorations={decorations} />)
    await waitFor(() => expect(container.querySelectorAll('[data-tone]')).toHaveLength(2))
    expect(container.querySelector('[data-tone="success"]')).toHaveTextContent('{{ok}}')
    expect(container.querySelector('[data-tone="success"]')).toHaveAttribute('title', 'resolved')
    expect(container.querySelector('[data-tone="danger"]')).toHaveAttribute('title', 'missing')
    rerender(<SingleLineEditor value="{{ok}}/{{no}}" label="Value" onChange={vi.fn()} decorations={[]} />)
    await waitFor(() => expect(container.querySelectorAll('[data-tone]')).toHaveLength(0))
  })

  it('offers the completion source options and applies the chosen one', async () => {
    const source: CompletionSource = vi.fn(async (text: string, position: number) =>
      text.endsWith('{{ba') ? { from: position - 2, options: [{ label: 'base_url', detail: 'Environment', apply: 'base_url}}' }] } : null,
    )
    const onChange = vi.fn()
    render(<SingleLineEditor value="" label="Request URL" onChange={onChange} completionSource={source} />)
    await userEvent.type(screen.getByRole('textbox', { name: 'Request URL' }), '{{{{ba')
    const option = await screen.findByRole('option', { name: /base_url/ })
    expect(option).toHaveTextContent('Environment')
    await new Promise((resolve) => setTimeout(resolve, COMPLETION_INTERACTION_DELAY_MS))
    await userEvent.keyboard('{Enter}')
    await waitFor(() => expect(onChange).toHaveBeenLastCalledWith('{{base_url}}'))
    expect(source).toHaveBeenCalledWith('{{ba', 4)
  })

  it('reports the selection', async () => {
    const onSelectionChange = vi.fn()
    render(<SingleLineEditor value="Bearer abc" label="Value" onChange={vi.fn()} onSelectionChange={onSelectionChange} />)
    act(() => editorView('Value').dispatch({ selection: { anchor: 7, head: 10 } }))
    expect(onSelectionChange).toHaveBeenLastCalledWith({ from: 7, to: 10, text: 'abc' })
  })
})
