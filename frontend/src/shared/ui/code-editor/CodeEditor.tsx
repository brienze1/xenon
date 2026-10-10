import { defaultKeymap, history, historyKeymap } from '@codemirror/commands'
import { defaultHighlightStyle, foldGutter, foldKeymap, syntaxHighlighting } from '@codemirror/language'
import { highlightSelectionMatches, searchKeymap } from '@codemirror/search'
import { Compartment, EditorState } from '@codemirror/state'
import { EditorView, drawSelection, keymap, lineNumbers } from '@codemirror/view'
import { useEffect, useRef } from 'react'
import { cn } from '@/shared/lib/cn'
import styles from './CodeEditor.module.css'
import {
  completionExtension,
  decorationsEffect,
  decorationsExtension,
  toneTheme,
  type CompletionSource,
  type EditorDecoration,
} from './editor-extensions'
import { loadLanguage, type CodeLanguage } from './languages'

interface CodeEditorProps {
  value: string
  language: CodeLanguage
  label: string
  onChange?: (value: string) => void
  decorations?: readonly EditorDecoration[]
  completionSource?: CompletionSource
  className?: string
}

const NO_DECORATIONS: readonly EditorDecoration[] = []

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

export function CodeEditor({
  value,
  language,
  label,
  onChange,
  decorations = NO_DECORATIONS,
  completionSource,
  className,
}: CodeEditorProps) {
  const containerRef = useRef<HTMLDivElement>(null)
  const viewRef = useRef<EditorView | null>(null)
  const languageRef = useRef(new Compartment())
  const onChangeRef = useRef(onChange)
  const completionRef = useRef(completionSource)
  const readOnly = onChange === undefined

  useEffect(() => {
    onChangeRef.current = onChange
    completionRef.current = completionSource
  }, [onChange, completionSource])

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
          decorationsExtension(),
          completionExtension(() => completionRef.current),
          EditorState.readOnly.of(readOnly),
          EditorView.editable.of(!readOnly),
          EditorView.contentAttributes.of({ 'aria-label': label }),
          EditorView.updateListener.of((update) => {
            if (update.docChanged) {
              onChangeRef.current?.(update.state.doc.toString())
            }
          }),
          theme,
          toneTheme,
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
    viewRef.current?.dispatch({ effects: decorationsEffect(decorations) })
  }, [decorations, value, readOnly, label])

  useEffect(() => {
    let active = true
    async function applyLanguage() {
      const extension = await loadLanguage(language)
      if (active) {
        viewRef.current?.dispatch({ effects: languageRef.current.reconfigure(extension) })
      }
    }
    void applyLanguage()
    return () => {
      active = false
    }
  }, [language, readOnly, label])

  return <div ref={containerRef} className={cn(styles.root, className)} translate="no" />
}
