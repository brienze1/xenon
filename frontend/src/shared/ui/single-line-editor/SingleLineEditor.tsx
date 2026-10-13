import { history, historyKeymap, defaultKeymap } from '@codemirror/commands'
import { Annotation, EditorState, Prec } from '@codemirror/state'
import { EditorView, keymap, placeholder as placeholderExtension } from '@codemirror/view'
import { useEffect, useImperativeHandle, useRef, type Ref } from 'react'
import { cn } from '@/shared/lib/cn'
import { DecorationCards } from '@/shared/ui/code-editor/DecorationCards'
import {
  completionExtension,
  decorationsEffect,
  decorationsExtension,
  selectionListener,
  toneTheme,
  type CompletionSource,
  type EditorDecoration,
  type EditorSelection,
} from '@/shared/ui/code-editor/editor-extensions'
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
  /** Sees pasted text first; returning true means it was handled and nothing is inserted. */
  onPaste?: (text: string) => boolean
  className?: string
  ref?: Ref<SingleLineEditorHandle>
}

export interface SingleLineEditorHandle {
  focus: () => void
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
  onPaste,
  className,
  ref,
}: SingleLineEditorProps) {
  const containerRef = useRef<HTMLDivElement>(null)
  const viewRef = useRef<EditorView | null>(null)
  const callbacks = useRef({ onChange, onEnter, completionSource, onSelectionChange, onPaste })

  useEffect(() => {
    callbacks.current = { onChange, onEnter, completionSource, onSelectionChange, onPaste }
  }, [onChange, onEnter, completionSource, onSelectionChange, onPaste])

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
          EditorView.domEventHandlers({
            paste: (event) => {
              const text = event.clipboardData?.getData('text/plain') ?? ''
              if (text === '' || callbacks.current.onPaste?.(text) !== true) {
                return false
              }
              event.preventDefault()
              return true
            },
          }),
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

  useImperativeHandle(ref, () => ({ focus: () => viewRef.current?.focus() }), [])

  useEffect(() => {
    const view = viewRef.current
    if (view && view.state.doc.toString() !== value) {
      view.dispatch({ changes: { from: 0, to: view.state.doc.length, insert: value }, annotations: fromProps.of(true) })
    }
  }, [value])

  useEffect(() => {
    viewRef.current?.dispatch({ effects: decorationsEffect(decorations) })
  }, [decorations, value])

  return (
    <>
      <div ref={containerRef} className={cn(styles.root, className)} />
      <DecorationCards host={containerRef} />
    </>
  )
}
