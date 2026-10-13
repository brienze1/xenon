import type { VariableSuggestion } from '@/platform/mappers/inspection'
import type { EditorCompletions } from '@/shared/ui/code-editor/editor-extensions'
import { scopeLabel } from './variable-decorations'

const OPEN = '{{'
const CLOSE = '}}'
const BRACES = /[{}]/

export interface VariableCompletionQuery {
  /** Where the typed name starts (right after `{{`). */
  from: number
  prefix: string
  /** The reference is already closed after the cursor, so only the name is inserted. */
  closed: boolean
}

/** Finds the `{{name` being typed at the cursor, or `null` when the cursor is not inside an open reference. */
export function variableCompletionQuery(text: string, position: number): VariableCompletionQuery | null {
  const before = text.slice(0, position)
  const open = before.lastIndexOf(OPEN)
  if (open < 0) {
    return null
  }
  const prefix = before.slice(open + OPEN.length)
  if (BRACES.test(prefix)) {
    return null
  }
  return { from: open + OPEN.length, prefix, closed: text.slice(position).startsWith(CLOSE) }
}

export function toCompletions(query: VariableCompletionQuery, suggestions: VariableSuggestion[], environmentName: string | null): EditorCompletions {
  return {
    from: query.from,
    options: suggestions.map((suggestion) => ({
      label: suggestion.name,
      detail: suggestion.secret ? `${scopeLabel(suggestion.scope, environmentName)} · secret` : scopeLabel(suggestion.scope, environmentName),
      apply: query.closed ? suggestion.name : `${suggestion.name}${CLOSE}`,
    })),
  }
}
