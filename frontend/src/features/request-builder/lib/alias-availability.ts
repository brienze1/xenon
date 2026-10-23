import type { EnvironmentColor, EnvironmentSummary } from '@/platform/mappers/environment'
import type { SavedAlias } from '@/platform/mappers/request'

export const DELETED_ENVIRONMENT = 'Deleted environment'

export interface AliasEnvironmentLabel {
  id: string
  name: string
  color: EnvironmentColor
  deleted: boolean
}

/** Mirrors Go's rule (Go refuses the send anyway): an alias for all environments, or one whose environments include the active one. */
export function isAliasAvailable(alias: Pick<SavedAlias, 'environmentIds'>, activeId: string | null): boolean {
  return alias.environmentIds.length === 0 || (activeId !== null && alias.environmentIds.includes(activeId))
}

/** Each linked environment by its current name and colour, or flagged as deleted when no environment has its id. */
export function aliasEnvironmentLabels(alias: Pick<SavedAlias, 'environmentIds'>, environments: EnvironmentSummary[]): AliasEnvironmentLabel[] {
  return alias.environmentIds.map((id) => {
    const environment = environments.find((candidate) => candidate.id === id && candidate.error === null)
    return environment
      ? { id, name: environment.name, color: environment.color, deleted: false }
      : { id, name: DELETED_ENVIRONMENT, color: 'none', deleted: true }
  })
}
