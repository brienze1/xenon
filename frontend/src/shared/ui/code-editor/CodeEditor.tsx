import { defaultKeymap, history, historyKeymap } from '@codemirror/commands'
import { defaultHighlightStyle, foldGutter, foldKeymap, syntaxHighlighting } from '@codemirror/language'
import { highlightSelectionMatches, searchKeymap } from '@codemirror/search'
import { Compartment, EditorState } from '@codemirror/state'
import { EditorView, drawSelection, keymap, lineNumbers } from '@codemirror/view'
import { useEffect, useImperativeHandle, useRef, type Ref } from 'react'
import { cn } from '@/shared/lib/cn'
import styles from './CodeEditor.module.css'
import {
  completionExtension,
  decorationsEffect,
  decorationsExtension,
  diagnosticDecorations,
  toneTheme,
  type CompletionSource,
  type EditorDecoration,
  type EditorDiagnostic,
} from './editor-extensions'
import { loadLanguage, type CodeLanguage } from './languages'

/** Imperative access for toolbars next to the editor (e.g. a snippets menu). */
export interface CodeEditorHandle {
  /** Inserts text at the cursor (replacing the selection) and focuses the editor. */
  insert: (text: string) => void
}

interface CodeEditorProps {
  value: string
  language: CodeLanguage
  label: string
  onChange?: (value: string) => void
  decorations?: readonly EditorDecoration[]
  /** Problems underlined in the text and listed under the editor. */
  diagnostics?: readonly EditorDiagnostic[]
  completionSource?: CompletionSource
  className?: string
  ref?: Ref<CodeEditorHandle>
}

const NO_DECORATIONS: readonly EditorDecoration[] = []
const NO_DIAGNOSTICS: readonly EditorDiagnostic[] = []

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
  diagnostics = NO_DIAGNOSTICS,
  completionSource,
  className,
  ref,
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

  useImperativeHandle(
    ref,
    () => ({
      insert: (text) => {
        const view = viewRef.current
        if (!view) {
          return
        }
        const { from, to } = view.state.selection.main
        view.dispatch({ changes: { from, to, insert: text }, selection: { anchor: from + text.length } })
        view.focus()
      },
    }),
    [],
  )

  useEffect(() => {
    const view = viewRef.current
    if (view) {
      view.dispatch({ effects: decorationsEffect([...decorations, ...diagnosticDecorations(view.state.doc, diagnostics)]) })
    }
  }, [decorations, diagnostics, value, readOnly, label])

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

  return (
    <div className={cn(styles.frame, className)}>
      <div ref={containerRef} className={styles.root} translate="no" />
      {diagnostics.length > 0 ? (
        <ul aria-label={`${label} problems`} className={styles.diagnostics}>
          {diagnostics.map((diagnostic, index) => (
            <li key={index}>
              Line {diagnostic.line}, column {diagnostic.column}: {diagnostic.message}
            </li>
          ))}
        </ul>
      ) : null}
    </div>
  )
}
