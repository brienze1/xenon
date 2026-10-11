import { useCallback } from 'react'
import type { ItemScriptsSlot } from '@/features/sidebar/components/SettingsForm'
import { ScriptsEditor } from '@/features/scripts/components/ScriptsEditor'
import { useWorkspacesStore } from '@/features/workspaces/store'
import { suggestVariables } from '@/platform/services/variables'

interface ItemScriptsProps {
  slot: ItemScriptsSlot
}

/** The Scripts tab of collection and folder settings. */
export function ItemScripts({ slot }: ItemScriptsProps) {
  const { itemId } = slot
  const suggest = useCallback(
    async (prefix: string) => {
      const localId = useWorkspacesStore.getState().activeLocalId
      if (!localId) {
        return []
      }
      try {
        return (await suggestVariables(localId, itemId, prefix)).map((suggestion) => ({ name: suggestion.name, detail: suggestion.scope }))
      } catch {
        return []
      }
    },
    [itemId],
  )
  return (
    <ScriptsEditor
      scripts={slot.scripts}
      onChange={slot.onChange}
      inheritScripts={slot.inheritScripts}
      onInheritChange={slot.onInheritChange}
      suggestVariables={suggest}
    />
  )
}
