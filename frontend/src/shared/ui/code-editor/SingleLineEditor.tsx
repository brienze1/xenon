import { history, historyKeymap, defaultKeymap } from '@codemirror/commands'
import { Annotation, EditorState, Prec } from '@codemirror/state'
import { EditorView, keymap, placeholder as placeholderExtension } from '@codemirror/view'
import { useEffect, useRef } from 'react'
import { cn } from '@/shared/lib/cn'
import {
  completionExtension,
  decorationsEffect,
  decorationsExtension,
  selectionListener,
  toneTheme,
  type CompletionSource,
  type EditorDecoration,
  type EditorSelection,
} from './editor-extensions'
import styles from './SingleLineEditor.module.css'

interface SingleLineEditorProps {
  value: string
  label: string
  onChange: (value: string) => void
  /** Enter without a modifier (and with no suggestion open). */
  onEnter?: () => void
  placeholder?: string
  decorations?: readonly EditorDecoration[]
  completionSource?: CompletionSource
  onSelectionChange?: (selection: EditorSelection) => void
  className?: string
}

const NO_DECORATIONS: readonly EditorDecoration[] = []

/** Marks value updates coming from props, so they are not echoed back through onChange. */
const fromProps = Annotation.define<boolean>()

const LINE_BREAKS = /[\r\n]+/g

const theme = EditorView.theme(
  {
    '&': { backgroundColor: 'transparent', color: 'var(--color-text)', fontSize: 'var(--font-size-sm)' },
    '.cm-content': { fontFamily: 'var(--font-mono)', caretColor: 'var(--neon-cyan)', padding: '0' },
    '.cm-line': { padding: '0' },
    '.cm-scroller': { overflowX: 'auto', overflowY: 'hidden', scrollbarWidth: 'none' },
    '.cm-placeholder': { color: 'var(--color-text-muted)' },
    '&.cm-focused .cm-selectionBackground, .cm-selectionBackground': { backgroundColor: 'var(--color-selection)' },
  },
  { dark: true },
)

/** Keeps the document on one line: pasted or typed line breaks become spaces. */
const singleLine = EditorState.transactionFilter.of((transaction) => {
  if (!transaction.docChanged || transaction.newDoc.lines === 1) {
    return transaction
  }
  const changes: Array<{ from: number; to: number; insert: string }> = []
  transaction.changes.iterChanges((fromA, toA, _fromB, _toB, inserted) => {
    changes.push({ from: fromA, to: toA, insert: inserted.toString().replace(LINE_BREAKS, ' ') })
  })
  return [{ changes, selection: transaction.selection, scrollIntoView: transaction.scrollIntoView }]
})

/** A one-line code input (URL bar, table cells) with optional decorations and completion. */
export function SingleLineEditor({
  value,
  label,
  onChange,
  onEnter,
  placeholder = '',
  decorations = NO_DECORATIONS,
  completionSource,
  onSelectionChange,
  className,
}: SingleLineEditorProps) {
  const containerRef = useRef<HTMLDivElement>(null)
  const viewRef = useRef<EditorView | null>(null)
  const callbacks = useRef({ onChange, onEnter, completionSource, onSelectionChange })

  useEffect(() => {
    callbacks.current = { onChange, onEnter, completionSource, onSelectionChange }
  }, [onChange, onEnter, completionSource, onSelectionChange])

  useEffect(() => {
    const view = new EditorView({
      parent: containerRef.current as HTMLDivElement,
      state: EditorState.create({
        doc: value,
        extensions: [
          history(),
          singleLine,
          Prec.high(
            keymap.of([
              {
                key: 'Enter',
                run: () => {
                  callbacks.current.onEnter?.()
                  return true
                },
              },
            ]),
          ),
          keymap.of([...defaultKeymap, ...historyKeymap]),
          decorationsExtension(),
          completionExtension(() => callbacks.current.completionSource),
          selectionListener(() => callbacks.current.onSelectionChange),
          placeholderExtension(placeholder),
          EditorView.contentAttributes.of({ 'aria-label': label, 'aria-multiline': 'false', translate: 'no', spellcheck: 'false' }),
          EditorView.updateListener.of((update) => {
            if (update.docChanged && !update.transactions.some((transaction) => transaction.annotation(fromProps))) {
              callbacks.current.onChange(update.state.doc.toString())
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
    // eslint-disable-next-line react-hooks/exhaustive-deps -- the view is created once per label; value and decorations sync below
  }, [label, placeholder])

  useEffect(() => {
    const view = viewRef.current
    if (view && view.state.doc.toString() !== value) {
      view.dispatch({ changes: { from: 0, to: view.state.doc.length, insert: value }, annotations: fromProps.of(true) })
    }
  }, [value])

  useEffect(() => {
    viewRef.current?.dispatch({ effects: decorationsEffect(decorations) })
  }, [decorations, value])

  return <div ref={containerRef} className={cn(styles.root, className)} />
}
