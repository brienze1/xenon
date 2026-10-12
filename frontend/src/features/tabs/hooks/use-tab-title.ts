import { useEnvironmentsStore } from '@/features/environments/store'
import { useRequestBuilderStore } from '@/features/request-builder/store'
import { useTabsStore } from '@/features/tabs/store'

export const UNTITLED_REQUEST = 'Untitled Request'
export const GLOBALS_TITLE = 'Globals'
const UNKNOWN_ENVIRONMENT = 'Environment'

/** The name a tab shows: its request's name, its environment's name, or "Globals". */
export function useTabTitle(tabId: string | null): string {
  const kind = useTabsStore((state) => state.tabs.find((tab) => tab.id === tabId)?.kind)
  const environmentId = useTabsStore((state) => state.tabs.find((tab) => tab.id === tabId)?.environmentId ?? null)
  const requestName = useRequestBuilderStore((state) => (tabId ? state.tabs[tabId]?.draft.name : null) || UNTITLED_REQUEST)
  const environmentName = useEnvironmentsStore((state) =>
    environmentId === null ? null : (state.environments.find((environment) => environment.id === environmentId)?.name ?? null),
  )
  if (kind === 'globals') {
    return GLOBALS_TITLE
  }
  if (kind === 'environment') {
    return environmentName ?? UNKNOWN_ENVIRONMENT
  }
  return requestName
}
