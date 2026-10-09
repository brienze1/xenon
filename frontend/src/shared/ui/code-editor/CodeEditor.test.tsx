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
    expect(content).toHaveAttribute('contenteditable', 'false')
    rerender(<CodeEditor value={'{"b":2}'} language="json" label="Response body" />)
    await waitFor(() => expect(screen.getByLabelText('Response body')).toHaveTextContent('{"b":2}'))
  })

  it('is editable when it has an onChange handler', () => {
    render(<CodeEditor value="" language="text" label="Body" onChange={vi.fn()} />)
    expect(screen.getByLabelText('Body')).toHaveAttribute('contenteditable', 'true')
  })
})
