import { useShallow } from 'zustand/react/shallow'
import type { SavedAlias } from '@/platform/mappers/request'
import { useEnvironmentsStore } from '@/features/environments/store'
import { useRequestBuilderStore } from '@/features/request-builder/store'
import { aliasEnvironmentLabels, isAliasAvailable, type AliasEnvironmentLabel } from '@/features/request-builder/lib/alias-availability'

export interface UnavailableAlias {
  alias: SavedAlias
  labels: AliasEnvironmentLabel[]
}

/** The alias the tab shows when the active environment is not one of its environments, else null. */
export function useUnavailableAlias(tabId: string): UnavailableAlias | null {
  const alias = useRequestBuilderStore((state) => {
    const current = state.tabs[tabId]
    return current?.aliases.find((candidate) => candidate.id === current.aliasId) ?? null
  })
  const activeId = useEnvironmentsStore((state) => state.activeId)
  const environments = useEnvironmentsStore(useShallow((state) => state.environments))
  if (alias === null || isAliasAvailable(alias, activeId)) {
    return null
  }
  return { alias, labels: aliasEnvironmentLabels(alias, environments) }
}

/** "mx-test-user is only for HML, DEV." — or, when every environment it was for is gone, where to fix it. */
export function unavailableAliasMessage({ alias, labels }: UnavailableAlias): string {
  if (labels.every((label) => label.deleted)) {
    return `${alias.name} is only for environments that were deleted. Change its environments from the Aliases menu.`
  }
  return `${alias.name} is only for ${labels.map((label) => label.name).join(', ')}.`
}
