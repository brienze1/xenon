import { autocompletion, type CompletionContext, type CompletionResult } from '@codemirror/autocomplete'
import { StateEffect, StateField, type Extension, type Range, type Text } from '@codemirror/state'
import { Decoration, EditorView, type DecorationSet } from '@codemirror/view'

/** Visual tone of a decorated range; the meaning is decided by whoever passes the decorations. */
export type EditorTone = 'success' | 'danger' | 'accent' | 'info'

export interface EditorDecoration {
  from: number
  to: number
  tone: EditorTone
  tooltip: string
  /** Opens a hover card (see DecorationCardContext) instead of the plain tooltip. */
  card?: string
}

export interface EditorCompletionOption {
  label: string
  detail: string
  apply: string
}

export interface EditorCompletions {
  /** Position the chosen option replaces from (up to the cursor). */
  from: number
  options: EditorCompletionOption[]
}

/** Asked on every keystroke with the document and the cursor; `null` offers nothing. */
export type CompletionSource = (text: string, position: number) => Promise<EditorCompletions | null>

export interface EditorSelection {
  from: number
  to: number
  text: string
}

/** A problem to underline, at a 1-based line and column (lint or syntax errors). */
export interface EditorDiagnostic {
  line: number
  column: number
  message: string
}

/** Turns diagnostics into danger decorations covering the rest of the word (or the last character) at their position. */
export function diagnosticDecorations(doc: Text, diagnostics: readonly EditorDiagnostic[]): EditorDecoration[] {
  const decorations: EditorDecoration[] = []
  for (const diagnostic of diagnostics) {
    if (diagnostic.line < 1 || diagnostic.line > doc.lines) {
      continue
    }
    const line = doc.line(diagnostic.line)
    const start = Math.min(line.from + Math.max(diagnostic.column - 1, 0), line.to)
    const word = /^[\w$]+/.exec(doc.sliceString(start, line.to))?.[0].length ?? 1
    const to = Math.min(start + word, line.to)
    const from = to > start ? start : Math.max(line.from, to - 1)
    if (from < to) {
      decorations.push({ from, to, tone: 'danger', tooltip: diagnostic.message })
    }
  }
  return decorations
}

const replaceDecorations = StateEffect.define<readonly EditorDecoration[]>()

function toDecorationSet(decorations: readonly EditorDecoration[], length: number): DecorationSet {
  const ranges: Range<Decoration>[] = []
  for (const decoration of decorations) {
    if (decoration.from < decoration.to && decoration.to <= length) {
      ranges.push(
        Decoration.mark({
          class: `cm-tone-${decoration.tone}`,
          attributes:
            decoration.card === undefined
              ? { title: decoration.tooltip, 'data-tone': decoration.tone }
              : { 'aria-description': decoration.tooltip, 'data-card': decoration.card, 'data-tone': decoration.tone },
        }).range(decoration.from, decoration.to),
      )
    }
  }
  return Decoration.set(ranges, true)
}

const decorationField = StateField.define<DecorationSet>({
  create: () => Decoration.none,
  update(decorations, transaction) {
    let next = decorations.map(transaction.changes)
    for (const effect of transaction.effects) {
      if (effect.is(replaceDecorations)) {
        next = toDecorationSet(effect.value, transaction.state.doc.length)
      }
    }
    return next
  },
  provide: (field) => EditorView.decorations.from(field),
})

/** Decorations come from props: dispatch `decorationsEffect(list)` whenever they change. */
export function decorationsExtension(): Extension {
  return decorationField
}

export function decorationsEffect(decorations: readonly EditorDecoration[]): StateEffect<readonly EditorDecoration[]> {
  return replaceDecorations.of(decorations)
}

/** Completion reads the source through a getter so a new callback never rebuilds the editor. */
export function completionExtension(source: () => CompletionSource | undefined): Extension {
  return autocompletion({
    activateOnTyping: true,
    icons: false,
    override: [
      async (context: CompletionContext): Promise<CompletionResult | null> => {
        const current = source()
        if (!current) {
          return null
        }
        const completions = await current(context.state.doc.toString(), context.pos)
        if (!completions || completions.options.length === 0) {
          return null
        }
        return {
          from: completions.from,
          filter: false,
          options: completions.options.map((option) => ({ label: option.label, detail: option.detail, apply: option.apply })),
        }
      },
    ],
  })
}

export function selectionListener(onSelection: () => ((selection: EditorSelection) => void) | undefined): Extension {
  return EditorView.updateListener.of((update) => {
    const listener = onSelection()
    if (listener && (update.selectionSet || update.docChanged)) {
      const { from, to } = update.state.selection.main
      listener({ from, to, text: update.state.sliceDoc(from, to) })
    }
  })
}

export const toneTheme = EditorView.theme({
  // `span` covers syntax colouring nested inside a decorated range, so a {{variable}} inside a string keeps its tone.
  '.cm-tone-success, .cm-tone-success span': { color: 'var(--neon-green)' },
  '.cm-tone-danger, .cm-tone-danger span': { color: 'var(--neon-red)', textDecoration: 'underline wavy var(--neon-red)' },
  '.cm-tone-accent, .cm-tone-accent span': { color: 'var(--neon-magenta)' },
  // The whole line of a problem is tinted so a one-character mistake (a stray comma) is easy to spot.
  '.cm-line:has(.cm-tone-danger)': { backgroundColor: 'rgb(255 61 90 / 10%)' },
  '.cm-tone-info, .cm-tone-info span': { color: 'var(--neon-cyan)' },
  '.cm-tooltip-autocomplete': {
    backgroundColor: 'var(--color-surface-raised)',
    border: '1px solid var(--color-border)',
    boxShadow: 'var(--shadow-raised)',
  },
  '.cm-tooltip-autocomplete > ul > li[aria-selected]': { backgroundColor: 'var(--color-selection)', color: 'var(--color-text)' },
  '.cm-completionDetail': { color: 'var(--color-text-muted)', marginLeft: 'var(--space-3)', fontStyle: 'normal' },
})
