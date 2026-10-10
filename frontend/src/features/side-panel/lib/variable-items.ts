import type { EditableScope, ReferenceStatus, VariableInspection, VariableScope } from '@/platform/mappers/inspection'

export const SCOPE_NAMES: Record<EditableScope, string> = {
  environment: 'Environment',
  collection: 'Collection',
  globals: 'Globals',
}

export const STATUS_LABELS: Record<ReferenceStatus, string> = {
  resolved: 'Resolved',
  unresolved: 'Unresolved',
  missing_secret: 'Missing secret',
  locked_secret: 'Locked',
  dynamic: 'Dynamic',
}

/** The scopes a new variable can be added to, strongest first (only those the request can use). */
export function addableScopes(inspection: Pick<VariableInspection, 'scopes'>): EditableScope[] {
  return (['environment', 'collection', 'globals'] as const).filter((scope) => inspection.scopes[scope] !== undefined)
}

export function isEditableScope(scope: VariableScope): scope is EditableScope {
  return scope === 'environment' || scope === 'collection' || scope === 'globals'
}

const VARIABLE_NAME = /^[^{}\r\n]+$/

/** A name Go accepts for a variable: not blank, no braces or line breaks. */
export function isValidVariableName(name: string): boolean {
  return name.trim() !== '' && VARIABLE_NAME.test(name)
}
