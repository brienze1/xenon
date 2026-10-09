import { defaultKeymap, history, historyKeymap } from '@codemirror/commands'
import { defaultHighlightStyle, foldGutter, foldKeymap, syntaxHighlighting } from '@codemirror/language'
import { highlightSelectionMatches, searchKeymap } from '@codemirror/search'
import { Compartment, EditorState } from '@codemirror/state'
import { EditorView, drawSelection, keymap, lineNumbers } from '@codemirror/view'
import { useEffect, useRef } from 'react'
import { cn } from '@/shared/lib/cn'
import styles from './CodeEditor.module.css'
import { loadLanguage, type CodeLanguage } from './languages'

interface CodeEditorProps {
  value: string
  language: CodeLanguage
  label: string
  onChange?: (value: string) => void
  className?: string
}

const theme = EditorView.theme(
  {
    '&': { backgroundColor: 'var(--color-surface)', color: 'var(--color-text)', fontSize: 'var(--font-size-sm)' },
    '.cm-content': { fontFamily: 'var(--font-mono)', caretColor: 'var(--neon-cyan)' },
    '.cm-gutters': { backgroundColor: 'var(--color-surface)', color: 'var(--color-text-muted)', border: 'none' },
    '.cm-activeLine': { backgroundColor: 'var(--color-surface-raised)' },
    '&.cm-focused .cm-selectionBackground, .cm-selectionBackground': { backgroundColor: 'var(--color-selection)' },
  },
  { dark: true },
)

export function CodeEditor({ value, language, label, onChange, className }: CodeEditorProps) {
  const containerRef = useRef<HTMLDivElement>(null)
  const viewRef = useRef<EditorView | null>(null)
  const languageRef = useRef(new Compartment())
  const onChangeRef = useRef(onChange)
  const readOnly = onChange === undefined

  useEffect(() => {
    onChangeRef.current = onChange
  }, [onChange])

  useEffect(() => {
    const view = new EditorView({
      parent: containerRef.current as HTMLDivElement,
      state: EditorState.create({
        doc: value,
        extensions: [
          lineNumbers(),
          foldGutter(),
          history(),
          drawSelection(),
          highlightSelectionMatches(),
          syntaxHighlighting(defaultHighlightStyle, { fallback: true }),
          keymap.of([...defaultKeymap, ...historyKeymap, ...searchKeymap, ...foldKeymap]),
          languageRef.current.of([]),
          EditorState.readOnly.of(readOnly),
          EditorView.editable.of(!readOnly),
          EditorView.contentAttributes.of({ 'aria-label': label }),
          EditorView.updateListener.of((update) => {
            if (update.docChanged) {
              onChangeRef.current?.(update.state.doc.toString())
            }
          }),
          theme,
        ],
      }),
    })
    viewRef.current = view
    return () => {
      view.destroy()
      viewRef.current = null
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps -- the view is created once; value and language sync below
  }, [readOnly, label])

  useEffect(() => {
    const view = viewRef.current
    if (view && view.state.doc.toString() !== value) {
      view.dispatch({ changes: { from: 0, to: view.state.doc.length, insert: value } })
    }
  }, [value])

  useEffect(() => {
    let active = true
    void loadLanguage(language).then((extension) => {
      if (active) {
        viewRef.current?.dispatch({ effects: languageRef.current.reconfigure(extension) })
      }
    })
    return () => {
      active = false
    }
  }, [language, readOnly, label])

  return <div ref={containerRef} className={cn(styles.root, className)} />
}
