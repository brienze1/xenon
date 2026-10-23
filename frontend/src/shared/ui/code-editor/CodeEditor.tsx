import { defaultKeymap, history, historyKeymap } from '@codemirror/commands'
import { foldGutter, foldKeymap } from '@codemirror/language'
import { highlightSelectionMatches, openSearchPanel, searchKeymap } from '@codemirror/search'
import { Annotation, Compartment, EditorState } from '@codemirror/state'
import { EditorView, drawSelection, keymap, lineNumbers } from '@codemirror/view'
import { useEffect, useImperativeHandle, useMemo, useRef, useState, type Ref } from 'react'
import { cn } from '@/shared/lib/cn'
import { DecorationCards } from './DecorationCards'
import styles from './CodeEditor.module.css'
import { findSyntaxErrors } from './syntax-check'
import { syntaxColors } from './syntax-highlight'
import { wrapIndent } from './wrap-indent'
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

/** Marks value updates coming from props, so they are not echoed back through onChange. */
const fromProps = Annotation.define<boolean>()

/** Imperative access for toolbars next to the editor (e.g. a snippets menu). */
export interface CodeEditorHandle {
  /** Inserts text at the cursor (replacing the selection) and focuses the editor. */
  insert: (text: string) => void
  /** Focuses the editor and opens its search panel. */
  find: () => void
}

interface CodeEditorProps {
  value: string
  language: CodeLanguage
  label: string
  onChange?: (value: string) => void
  decorations?: readonly EditorDecoration[]
  /** Problems underlined in the text and listed under the editor. */
  diagnostics?: readonly EditorDiagnostic[]
  /** Marks where the text stops being valid for its language (bodies), like Postman. */
  checkSyntax?: boolean
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
  checkSyntax = false,
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
  const syntaxErrors = useSyntaxErrors(checkSyntax ? value : '', language)
  const problems = useMemo(() => (syntaxErrors.length > 0 ? [...diagnostics, ...syntaxErrors] : diagnostics), [diagnostics, syntaxErrors])

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
          syntaxColors,
          // Long lines wrap like Postman: no sideways scrolling, and a wrapped line keeps a single line number.
          EditorView.lineWrapping,
          wrapIndent,
          keymap.of([...defaultKeymap, ...historyKeymap, ...searchKeymap, ...foldKeymap]),
          languageRef.current.of([]),
          decorationsExtension(),
          completionExtension(() => completionRef.current),
          // Read-only still stays editable (focusable), so select-all, copy and find keep working in viewers.
          EditorState.readOnly.of(readOnly),
          EditorView.contentAttributes.of({ 'aria-label': label }),
          EditorView.updateListener.of((update) => {
            if (update.docChanged && !update.transactions.some((transaction) => transaction.annotation(fromProps))) {
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
      view.dispatch({ changes: { from: 0, to: view.state.doc.length, insert: value }, annotations: fromProps.of(true) })
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
      find: () => {
        const view = viewRef.current
        if (view) {
          view.focus()
          openSearchPanel(view)
        }
      },
    }),
    [],
  )

  useEffect(() => {
    const view = viewRef.current
    if (view) {
      view.dispatch({ effects: decorationsEffect([...decorations, ...diagnosticDecorations(view.state.doc, problems)]) })
    }
  }, [decorations, problems, value, readOnly, label])

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
      <DecorationCards host={containerRef} />
      {problems.length > 0 ? (
        <ul aria-label={`${label} problems`} className={styles.diagnostics}>
          {problems.map((diagnostic, index) => (
            <li key={index}>
              Line {diagnostic.line}, column {diagnostic.column}: {diagnostic.message}
            </li>
          ))}
        </ul>
      ) : null}
    </div>
  )
}

const SYNTAX_CHECK_DELAY_MS = 300

// Re-checks shortly after typing stops; an answer for older text is never shown.
function useSyntaxErrors(text: string, language: CodeLanguage): readonly EditorDiagnostic[] {
  const [checked, setChecked] = useState<{ text: string; language: CodeLanguage; errors: readonly EditorDiagnostic[] }>({
    text: '',
    language,
    errors: NO_DIAGNOSTICS,
  })
  useEffect(() => {
    if (text === '') {
      return undefined
    }
    let active = true
    const timer = window.setTimeout(() => {
      void findSyntaxErrors(text, language).then((errors) => {
        if (active) {
          setChecked({ text, language, errors })
        }
      })
    }, SYNTAX_CHECK_DELAY_MS)
    return () => {
      active = false
      window.clearTimeout(timer)
    }
  }, [text, language])
  return checked.text === text && checked.language === language ? checked.errors : NO_DIAGNOSTICS
}
