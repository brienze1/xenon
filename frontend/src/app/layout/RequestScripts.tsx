import { useCallback } from 'react'
import { useEnvironmentsStore } from '@/features/environments/store'
import { useRequestBuilderStore } from '@/features/request-builder/store'
import { ScriptsEditor } from '@/features/scripts/components/ScriptsEditor'
import type { VariableSuggestion } from '@/platform/mappers/inspection'

export function scopeDetail(suggestion: VariableSuggestion, environmentName: string | null): string {
  const scope = suggestion.scope === 'environment' && environmentName ? environmentName : suggestion.scope
  return suggestion.secret ? `${scope} · secret` : scope
}

interface RequestScriptsProps {
  tabId: string
}

/** The request's Scripts tab: the scripts editor on the request builder's draft. */
export function RequestScripts({ tabId }: RequestScriptsProps) {
  const scripts = useRequestBuilderStore((state) => state.tabs[tabId]?.draft.scripts)
  const inheritScripts = useRequestBuilderStore((state) => state.tabs[tabId]?.draft.inheritScripts ?? true)
  const updateScripts = useRequestBuilderStore((state) => state.updateScripts)
  const setInheritScripts = useRequestBuilderStore((state) => state.setInheritScripts)
  const suggest = useRequestBuilderStore((state) => state.suggest)
  const suggestVariables = useCallback(
    async (prefix: string) => {
      const environmentName = useEnvironmentsStore.getState().environments.find((environment) => environment.id === useEnvironmentsStore.getState().activeId)?.name ?? null
      return (await suggest(tabId, prefix)).map((suggestion) => ({ name: suggestion.name, detail: scopeDetail(suggestion, environmentName) }))
    },
    [suggest, tabId],
  )
  if (!scripts) {
    return null
  }
  return (
    <ScriptsEditor
      scripts={scripts}
      onChange={(next) => updateScripts(tabId, next)}
      inheritScripts={inheritScripts}
      onInheritChange={(inherit) => setInheritScripts(tabId, inherit)}
      suggestVariables={suggestVariables}
    />
  )
}
