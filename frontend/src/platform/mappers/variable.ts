import type {
  VariableEdit as VariableEditDto,
  VariableRow as VariableRowDto,
  VariableTable as VariableTableDto,
} from '../../../bindings/github.com/brienze1/xenon/internal/integration/entrypoint/dto/models'

export const GLOBALS_SCOPE = 'globals'
export const GIT_HISTORY_WARNING = 'git_history'

export function environmentScope(environmentId: string): string {
  return `env:${environmentId}`
}

export function collectionScope(collectionId: string): string {
  return `collection:${collectionId}`
}

export const SECRET_STATES = ['set', 'missing', 'locked'] as const

export type SecretState = (typeof SECRET_STATES)[number]

export interface VariableRow {
  key: string
  initialValue: string
  currentValue: string
  hasCurrentOverride: boolean
  secret: boolean
  secretState: SecretState | null
  enabled: boolean
  description: string
  duplicate: boolean
  surroundingSpaces: boolean
}

export interface VariableTable {
  scope: string
  rows: VariableRow[]
  warnings: string[]
}

/**
 * One row of a table save. `currentValue` and `secretValue` are `null` to keep what is stored;
 * an empty `currentValue` clears the local override and an empty `secretValue` clears the secret.
 */
export interface VariableEdit {
  key: string
  previousKey: string | null
  initialValue: string
  currentValue: string | null
  secret: boolean
  secretValue: string | null
  enabled: boolean
  description: string
}

function toSecretState(dto: VariableRowDto): SecretState | null {
  if (!dto.secret) {
    return null
  }
  return (SECRET_STATES as readonly string[]).includes(dto.secret_state ?? '') ? (dto.secret_state as SecretState) : 'missing'
}

export function toVariableRow(dto: VariableRowDto): VariableRow {
  return {
    key: dto.key,
    initialValue: dto.secret ? '' : dto.initial_value,
    currentValue: dto.secret ? '' : dto.current_value,
    hasCurrentOverride: dto.has_current_override,
    secret: dto.secret,
    secretState: toSecretState(dto),
    enabled: dto.enabled,
    description: dto.description,
    duplicate: dto.duplicate,
    surroundingSpaces: dto.surrounding_spaces,
  }
}

export function toVariableTable(dto: VariableTableDto | null, scope: string): VariableTable {
  return {
    scope: dto?.scope ? dto.scope : scope,
    rows: (dto?.rows ?? []).map(toVariableRow),
    warnings: dto?.warnings ?? [],
  }
}

export function toVariableEditDto(edit: VariableEdit): VariableEditDto {
  const dto: VariableEditDto = {
    key: edit.key,
    initial_value: edit.secret ? '' : edit.initialValue,
    description: edit.description,
    enabled: edit.enabled,
    secret: edit.secret,
    set_current: !edit.secret && edit.currentValue !== null,
  }
  if (edit.previousKey !== null) {
    dto.previous_key = edit.previousKey
  }
  if (dto.set_current) {
    dto.current_value = edit.currentValue ?? ''
  }
  if (edit.secret && edit.secretValue !== null) {
    dto.secret_value = edit.secretValue
  }
  return dto
}
