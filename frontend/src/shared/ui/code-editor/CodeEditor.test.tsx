import { render, screen, waitFor } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { CodeEditor } from './CodeEditor'
import { loadLanguage } from './languages'

describe('loadLanguage', () => {
  it('loads a language extension per supported language and none for text', async () => {
    for (const language of ['json', 'xml', 'html', 'javascript'] as const) {
      expect(await loadLanguage(language)).not.toEqual([])
    }
    expect(await loadLanguage('text')).toEqual([])
  })
})

describe('CodeEditor', () => {
  it('shows the document in a labelled editor and follows value changes', async () => {
    const { rerender } = render(<CodeEditor value={'{"a":1}'} language="json" label="Response body" />)
    const content = screen.getByLabelText('Response body')
    expect(content).toHaveTextContent('{"a":1}')
    expect(content).toHaveAttribute('aria-readonly', 'true')
    rerender(<CodeEditor value={'{"b":2}'} language="json" label="Response body" />)
    await waitFor(() => expect(screen.getByLabelText('Response body')).toHaveTextContent('{"b":2}'))
  })

  it('does not echo a new value from its props back through onChange', async () => {
    const onChange = vi.fn()
    const { rerender } = render(<CodeEditor value={'{"user":"luis"}'} language="json" label="Raw body" onChange={onChange} />)
    rerender(<CodeEditor value={'{"user":"mx"}'} language="json" label="Raw body" onChange={onChange} />)
    await waitFor(() => expect(screen.getByLabelText('Raw body')).toHaveTextContent('{"user":"mx"}'))
    expect(onChange).not.toHaveBeenCalled()
  })

  it('stays focusable when read-only, so select-all and copy reach the text', () => {
    render(<CodeEditor value={'{"a":1}'} language="json" label="Response body" />)
    const content = screen.getByLabelText('Response body')
    expect(content).toHaveAttribute('contenteditable', 'true')
    content.focus()
    expect(content).toHaveFocus()
  })

  it('is editable when it has an onChange handler', () => {
    render(<CodeEditor value="" language="text" label="Body" onChange={vi.fn()} />)
    expect(screen.getByLabelText('Body')).toHaveAttribute('contenteditable', 'true')
  })
})

describe('CodeEditor decorations and completion', () => {
  it('decorates ranges and offers completions in a multi-line body', async () => {
    const source = vi.fn(async (_text: string, position: number) => ({ from: position, options: [{ label: 'user', detail: 'Globals', apply: 'user}}' }] }))
    render(
      <CodeEditor
        value={'{"who": "{{user}}"}'}
        language="text"
        label="Raw body"
        onChange={vi.fn()}
        decorations={[{ from: 9, to: 17, tone: 'success', tooltip: 'luis — Globals' }]}
        completionSource={source}
      />,
    )
    expect(await screen.findByTitle('luis — Globals')).toHaveTextContent('{{user}}')
  })
})
