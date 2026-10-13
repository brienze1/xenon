import type {
  VariableInspection as VariableInspectionDto,
  VariableReference as VariableReferenceDto,
  VariableSuggestion as VariableSuggestionDto,
  VariableUsage as VariableUsageDto,
} from '../../../bindings/github.com/brienze1/xenon/internal/integration/entrypoint/dto/models'

export const REFERENCE_STATUSES = ['resolved', 'unresolved', 'missing_secret', 'locked_secret', 'dynamic'] as const
export const VARIABLE_SCOPES = ['environment', 'collection', 'globals', 'dynamic', 'none'] as const
/** The scopes a variable can be added to or edited in. */
export const EDITABLE_SCOPES = ['environment', 'collection', 'globals'] as const

export type ReferenceStatus = (typeof REFERENCE_STATUSES)[number]
export type VariableScope = (typeof VARIABLE_SCOPES)[number]
export type EditableScope = (typeof EDITABLE_SCOPES)[number]

/** One `{{name}}` in a request field; `start`/`end` are string offsets in that field's text. */
export interface VariableReference {
  field: string
  name: string
  start: number
  end: number
  status: ReferenceStatus
  scope: VariableScope
  displayValue: string
  secret: boolean
}

export interface VariableUsage {
  name: string
  status: ReferenceStatus
  scope: VariableScope
  displayValue: string
  secret: boolean
  usedIn: string[]
}

export interface VariableInspection {
  references: VariableReference[]
  variables: VariableUsage[]
  /** The table reference (`env:<id>`, `collection:<id>`, `globals`) of each scope this request can use. */
  scopes: Partial<Record<EditableScope, string>>
}

export interface VariableSuggestion {
  name: string
  scope: VariableScope
  description: string
  secret: boolean
  dynamic: boolean
}

function toStatus(status: string): ReferenceStatus {
  return (REFERENCE_STATUSES as readonly string[]).includes(status) ? (status as ReferenceStatus) : 'unresolved'
}

function toScope(scope: string): VariableScope {
  return (VARIABLE_SCOPES as readonly string[]).includes(scope) ? (scope as VariableScope) : 'none'
}

function isEditableScope(scope: string): scope is EditableScope {
  return (EDITABLE_SCOPES as readonly string[]).includes(scope)
}

function toReference(dto: VariableReferenceDto): VariableReference {
  return {
    field: dto.field,
    name: dto.name,
    start: dto.start,
    end: dto.end,
    status: toStatus(dto.status),
    scope: toScope(dto.scope),
    displayValue: dto.display_value,
    secret: dto.secret,
  }
}

function toUsage(dto: VariableUsageDto): VariableUsage {
  return {
    name: dto.name,
    status: toStatus(dto.status),
    scope: toScope(dto.scope),
    displayValue: dto.display_value,
    secret: dto.secret,
    usedIn: dto.used_in ?? [],
  }
}

export function emptyInspection(): VariableInspection {
  return { references: [], variables: [], scopes: {} }
}

export function toVariableInspection(dto: VariableInspectionDto | null): VariableInspection {
  const scopes: Partial<Record<EditableScope, string>> = {}
  for (const scope of dto?.scopes ?? []) {
    if (isEditableScope(scope.scope)) {
      scopes[scope.scope] = scope.ref
    }
  }
  return {
    references: (dto?.references ?? []).map(toReference),
    variables: (dto?.variables ?? []).map(toUsage),
    scopes,
  }
}

export function toVariableSuggestion(dto: VariableSuggestionDto): VariableSuggestion {
  return { name: dto.name, scope: toScope(dto.scope), description: dto.description, secret: dto.secret, dynamic: dto.dynamic }
}
