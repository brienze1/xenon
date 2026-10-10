import type { VariableInspection, VariableReference, VariableScope } from '@/platform/mappers/inspection'
import type { EditorDecoration, EditorTone } from '@/shared/ui/code-editor/editor-extensions'

/** How a variable's source scope is named in hovers and lists. */
export function scopeLabel(scope: VariableScope, environmentName: string | null): string {
  switch (scope) {
    case 'environment':
      return environmentName ? `Environment: ${environmentName}` : 'Environment'
    case 'collection':
      return 'Collection'
    case 'globals':
      return 'Globals'
    case 'dynamic':
      return 'Dynamic'
    default:
      return 'Not defined'
  }
}

export function referenceTone(reference: Pick<VariableReference, 'status' | 'secret'>): EditorTone {
  switch (reference.status) {
    case 'resolved':
      return reference.secret ? 'accent' : 'success'
    case 'dynamic':
      return 'info'
    default:
      return 'danger'
  }
}

/** The hover of a reference: its value (masked for secrets) and where it comes from. */
export function referenceTooltip(reference: Pick<VariableReference, 'status' | 'scope' | 'displayValue'>, environmentName: string | null): string {
  const source = scopeLabel(reference.scope, environmentName)
  switch (reference.status) {
    case 'resolved':
      return `${reference.displayValue} — ${source}`
    case 'dynamic':
      return 'Generated at send — Dynamic variable'
    case 'missing_secret':
      return `Secret with no value on this machine — ${source}`
    case 'locked_secret':
      return `Secret in a locked vault — ${source}`
    default:
      return 'Unresolved variable — sent as written'
  }
}

/** Groups the inspected references into editor decorations per request field (`url`, `params.0.value`, …). */
export function decorationsByField(inspection: VariableInspection | null, environmentName: string | null): ReadonlyMap<string, EditorDecoration[]> {
  const byField = new Map<string, EditorDecoration[]>()
  for (const reference of inspection?.references ?? []) {
    const decorations = byField.get(reference.field) ?? []
    decorations.push({
      from: reference.start,
      to: reference.end,
      tone: referenceTone(reference),
      tooltip: referenceTooltip(reference, environmentName),
    })
    byField.set(reference.field, decorations)
  }
  return byField
}
